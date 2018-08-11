package quic

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"

	"github.com/bifurcation/mint"
	"github.com/lucas-clemente/quic-go/internal/crypto"
	"github.com/lucas-clemente/quic-go/internal/handshake"
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/internal/wire"
	"github.com/lucas-clemente/quic-go/qerr"
)

type tlsSession struct {
	connID protocol.ConnectionID
	sess   quicSession
}

type serverTLS struct {
	conn              net.PacketConn
	config            *Config
	supportedVersions []protocol.VersionNumber
	mintConf          *mint.Config
	params            *handshake.TransportParameters
	newMintConn       func(*handshake.CryptoStreamConn, protocol.VersionNumber) (handshake.MintTLS, <-chan handshake.TransportParameters, error)

	newSession func(connection, sessionRunner, protocol.ConnectionID, protocol.ConnectionID, protocol.PacketNumber, *Config, handshake.MintTLS, *handshake.CryptoStreamConn, crypto.AEAD, *handshake.TransportParameters, protocol.VersionNumber, utils.Logger) (quicSession, error)

	sessionRunner sessionRunner
	sessionChan   chan<- tlsSession

	logger utils.Logger
}

func newServerTLS(
	conn net.PacketConn,
	config *Config,
	runner sessionRunner,
	cookieHandler *handshake.CookieHandler,
	tlsConf *tls.Config,
	logger utils.Logger,
) (*serverTLS, <-chan tlsSession, error) {
	mconf, err := tlsToMintConfig(tlsConf, protocol.PerspectiveServer)
	if err != nil {
		return nil, nil, err
	}
	mconf.RequireCookie = true
	cs, err := mint.NewDefaultCookieProtector()
	if err != nil {
		return nil, nil, err
	}
	mconf.CookieProtector = cs
	mconf.CookieHandler = cookieHandler

	sessionChan := make(chan tlsSession)
	s := &serverTLS{
		conn:              conn,
		config:            config,
		supportedVersions: config.Versions,
		mintConf:          mconf,
		sessionRunner:     runner,
		sessionChan:       sessionChan,
		params: &handshake.TransportParameters{
			StreamFlowControlWindow:     protocol.ReceiveStreamFlowControlWindow,
			ConnectionFlowControlWindow: protocol.ReceiveConnectionFlowControlWindow,
			IdleTimeout:                 config.IdleTimeout,
			MaxBidiStreams:              uint16(config.MaxIncomingStreams),
			MaxUniStreams:               uint16(config.MaxIncomingUniStreams),
			DisableMigration:            true,
		},
		newSession: newTLSServerSession,
		logger:     logger,
	}
	s.newMintConn = s.newMintConnImpl
	return s, sessionChan, nil
}

func (s *serverTLS) HandleInitial(remoteAddr net.Addr, hdr *wire.Header, data []byte) {
	// TODO: add a check that DestConnID == SrcConnID
	s.logger.Debugf("Received a Packet. Handling it statelessly.")
	sess, connID, err := s.handleInitialImpl(remoteAddr, hdr, data)
	if err != nil {
		s.logger.Errorf("Error occurred handling initial packet: %s", err)
		return
	}
	if sess == nil { // a stateless reset was done
		return
	}
	s.sessionChan <- tlsSession{
		connID: connID,
		sess:   sess,
	}
}

// will be set to s.newMintConn by the constructor
func (s *serverTLS) newMintConnImpl(bc *handshake.CryptoStreamConn, v protocol.VersionNumber) (handshake.MintTLS, <-chan handshake.TransportParameters, error) {
	extHandler := handshake.NewExtensionHandlerServer(s.params, s.config.Versions, v, s.logger)
	conf := s.mintConf.Clone()
	conf.ExtensionHandler = extHandler
	return newMintController(bc, conf, protocol.PerspectiveServer), extHandler.GetPeerParams(), nil
}

func (s *serverTLS) sendConnectionClose(remoteAddr net.Addr, clientHdr *wire.Header, aead crypto.AEAD, closeErr error) error {
	ccf := &wire.ConnectionCloseFrame{
		ErrorCode:    qerr.HandshakeFailed,
		ReasonPhrase: closeErr.Error(),
	}
	replyHdr := &wire.Header{
		IsLongHeader:     true,
		Type:             protocol.PacketTypeHandshake,
		SrcConnectionID:  clientHdr.DestConnectionID,
		DestConnectionID: clientHdr.SrcConnectionID,
		PacketNumber:     1, // random packet number
		PacketNumberLen:  protocol.PacketNumberLen1,
		Version:          clientHdr.Version,
	}
	data, err := packUnencryptedPacket(aead, replyHdr, ccf, protocol.PerspectiveServer, s.logger)
	if err != nil {
		return err
	}
	_, err = s.conn.WriteTo(data, remoteAddr)
	return err
}

func (s *serverTLS) handleInitialImpl(remoteAddr net.Addr, hdr *wire.Header, data []byte) (quicSession, protocol.ConnectionID, error) {
	if hdr.DestConnectionID.Len() < protocol.MinConnectionIDLenInitial {
		return nil, nil, errors.New("dropping Initial packet with too short connection ID")
	}
	if len(hdr.Raw)+len(data) < protocol.MinInitialPacketSize {
		return nil, nil, errors.New("dropping too small Initial packet")
	}
	// check version, if not matching send VNP
	if !protocol.IsSupportedVersion(s.supportedVersions, hdr.Version) {
		s.logger.Debugf("Client offered version %s, sending VersionNegotiationPacket", hdr.Version)
		vnp, err := wire.ComposeVersionNegotiation(hdr.SrcConnectionID, hdr.DestConnectionID, s.supportedVersions)
		if err != nil {
			return nil, nil, err
		}
		_, err = s.conn.WriteTo(vnp, remoteAddr)
		return nil, nil, err
	}

	// unpack packet and check stream frame contents
	aead, err := crypto.NewNullAEAD(protocol.PerspectiveServer, hdr.DestConnectionID, protocol.VersionTLS)
	if err != nil {
		return nil, nil, err
	}
	frame, err := unpackInitialPacket(aead, hdr, data, s.logger, hdr.Version)
	if err != nil {
		s.logger.Debugf("Error unpacking initial packet: %s", err)
		return nil, nil, nil
	}
	sess, connID, err := s.handleUnpackedInitial(remoteAddr, hdr, frame, aead)
	if err != nil {
		if ccerr := s.sendConnectionClose(remoteAddr, hdr, aead, err); ccerr != nil {
			s.logger.Debugf("Error sending CONNECTION_CLOSE: %s", ccerr)
		}
		return nil, nil, err
	}
	return sess, connID, nil
}

func (s *serverTLS) handleUnpackedInitial(remoteAddr net.Addr, hdr *wire.Header, frame *wire.StreamFrame, aead crypto.AEAD) (quicSession, protocol.ConnectionID, error) {
	version := hdr.Version
	bc := handshake.NewCryptoStreamConn(remoteAddr)
	bc.AddDataForReading(frame.Data)
	tls, paramsChan, err := s.newMintConn(bc, version)
	if err != nil {
		return nil, nil, err
	}
	alert := tls.Handshake()
	if alert == mint.AlertStatelessRetry {
		// the HelloRetryRequest was written to the bufferConn
		// Take that data and write send a Retry packet
		f := &wire.StreamFrame{
			StreamID: version.CryptoStreamID(),
			Data:     bc.GetDataForWriting(),
		}
		srcConnID, err := protocol.GenerateConnectionID(s.config.ConnectionIDLength)
		if err != nil {
			return nil, nil, err
		}
		replyHdr := &wire.Header{
			IsLongHeader:     true,
			Type:             protocol.PacketTypeRetry,
			DestConnectionID: hdr.SrcConnectionID,
			SrcConnectionID:  srcConnID,
			PayloadLen:       f.Length(version) + protocol.ByteCount(aead.Overhead()),
			PacketNumber:     hdr.PacketNumber, // echo the client's packet number
			PacketNumberLen:  hdr.PacketNumberLen,
			Version:          version,
		}
		data, err := packUnencryptedPacket(aead, replyHdr, f, protocol.PerspectiveServer, s.logger)
		if err != nil {
			return nil, nil, err
		}
		_, err = s.conn.WriteTo(data, remoteAddr)
		return nil, nil, err
	}
	if alert != mint.AlertNoAlert {
		return nil, nil, alert
	}
	if tls.State() != mint.StateServerNegotiated {
		return nil, nil, fmt.Errorf("Expected mint state to be %s, got %s", mint.StateServerNegotiated, tls.State())
	}
	if alert := tls.Handshake(); alert != mint.AlertNoAlert {
		return nil, nil, alert
	}
	if tls.State() != mint.StateServerWaitFlight2 {
		return nil, nil, fmt.Errorf("Expected mint state to be %s, got %s", mint.StateServerWaitFlight2, tls.State())
	}
	params := <-paramsChan
	connID, err := protocol.GenerateConnectionID(s.config.ConnectionIDLength)
	if err != nil {
		return nil, nil, err
	}
	s.logger.Debugf("Changing source connection ID to %s.", connID)
	sess, err := s.newSession(
		&conn{pconn: s.conn, currentAddr: remoteAddr},
		s.sessionRunner,
		hdr.SrcConnectionID,
		connID,
		protocol.PacketNumber(1), // TODO: use a random packet number here
		s.config,
		tls,
		bc,
		aead,
		&params,
		version,
		s.logger,
	)
	if err != nil {
		return nil, nil, err
	}
	cs := sess.getCryptoStream()
	cs.setReadOffset(frame.DataLen())
	bc.SetStream(cs)
	go sess.run()
	return sess, connID, nil
}
