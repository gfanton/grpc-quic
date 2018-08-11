package grpcquic

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	quic "github.com/lucas-clemente/quic-go"
	"google.golang.org/grpc"
)

var quicConfig = &quic.Config{
	// MaxReceiveStreamFlowControlWindow:     3 * (1 << 20),   // 3 MB
	// MaxReceiveConnectionFlowControlWindow: 4.5 * (1 << 20), // 4.5 MB
	// Versions: []quic.VersionNumber{101},
	// AcceptCookie: func(clientAddr net.Addr, cookie *quic.Cookie) bool {
	// 	// TODO(#6): require source address validation when under load
	// 	return true
	// },
	KeepAlive: true,
}

// var quicConfig *quic.Config

type GrpcQuicTransport struct {
	tlsConf *tls.Config
	pconn   net.PacketConn
}

// NewTransport creates a new QUIC transport
func NewGrpcQuicTransport(addr string, tlsConf *tls.Config) (*GrpcQuicTransport, error) {
	// create a packet conn for outgoing connections
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	return &GrpcQuicTransport{
		tlsConf: tlsConf,
		pconn:   conn,
	}, nil
}

func (t *GrpcQuicTransport) Dial(target string, td time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), td)
	defer cancel()

	udpAddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		return nil, err
	}

	sess, err := quic.DialContext(ctx, t.pconn, udpAddr, target, t.tlsConf, quicConfig)
	if err != nil {
		return nil, err
	}

	s, err := sess.OpenStreamSync()
	if err != nil {
		return nil, err
	}

	return &Conn{sess, s}, nil
}

func (t *GrpcQuicTransport) Listener() (net.Listener, error) {
	ql, err := quic.Listen(t.pconn, t.tlsConf, quicConfig)
	if err != nil {
		return nil, err
	}

	return &Listener{ql}, nil
}

func (t *GrpcQuicTransport) GrpcDial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	fmtOpts := append([]grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDialer(t.Dial)}, opts...)
	return grpc.Dial(target, fmtOpts...)

}

func (t *GrpcQuicTransport) GrpcServe(gs *grpc.Server) error {
	l, err := t.Listener()
	if err != nil {
		return err
	}

	return gs.Serve(l)
}

func (t *GrpcQuicTransport) NewGrpcServer(opts ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(opts...)
}
