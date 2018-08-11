package grpcquic

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	quicnet "github.com/gfanton/grpc-quic/net"
	options "github.com/gfanton/grpc-quic/opts"
	"github.com/gfanton/grpc-quic/transports"
	quic "github.com/lucas-clemente/quic-go"
	ma "github.com/multiformats/go-multiaddr"
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

type addr struct {
	code   int
	target string
	port   string
	laddr  string

	m ma.Multiaddr
}

func newAddr(m ma.Multiaddr) (addr, error) {
	var err error

	a := addr{}
	a.m = m
	for _, p := range m.Protocols() {
		switch p.Code {
		case ma.P_IP4:
			a.target, err = m.ValueForProtocol(ma.P_IP4)
		case ma.P_UDP:
			a.code = ma.P_UDP
			a.port, err = m.ValueForProtocol(ma.P_UDP)
		default:
			return addr{}, fmt.Errorf("Protocol `%s` not supported", p.Name)
		}

		if err != nil {
			return addr{}, fmt.Errorf("Protocol value error: %s", err)
		}
	}

	a.laddr = a.target + ":" + a.port
	return a, nil
}

func (a *addr) String() string {
	return a.laddr
}

func newPacketConn(addr string) (net.PacketConn, error) {
	// create a packet conn for outgoing connections
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	return net.ListenUDP("udp", udpAddr)
}

func newQuicDialer(pconn net.PacketConn, tlsConf *tls.Config) func(target string, td time.Duration) (net.Conn, error) {
	return func(target string, td time.Duration) (net.Conn, error) {
		ctx, cancel := context.WithTimeout(context.Background(), td)
		defer cancel()

		udpAddr, err := net.ResolveUDPAddr("udp", target)
		if err != nil {
			return nil, err
		}

		sess, err := quic.DialContext(ctx, pconn, udpAddr, target, tlsConf, quicConfig)
		if err != nil {
			return nil, err
		}

		return quicnet.NewConn(sess)
	}
}

func Dial(target string, opts ...options.DialOption) (*grpc.ClientConn, error) {
	m, err := ma.NewMultiaddr(target)
	if err != nil {
		return nil, err
	}

	addr, err := newAddr(m)
	if err != nil {
		return nil, err
	}

	cfg := options.NewClientConfig()
	if err := cfg.Apply(opts...); err != nil {
		return nil, err
	}

	pconn, err := newPacketConn("127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	creds := new(transports.Credentials)
	dialer := newQuicDialer(pconn, cfg.TLSConf)
	grpcOpts := []grpc.DialOption{
		grpc.WithDialer(dialer),
		grpc.WithTransportCredentials(creds),
	}

	grpcOpts = append(grpcOpts, cfg.GrpcDialOptions...)
	return grpc.Dial(addr.String(), grpcOpts...)
}

func NewListener(addr string, tlsConf *tls.Config) (net.Listener, error) {
	pconn, err := newPacketConn(addr)
	if err != nil {
		return nil, err
	}

	ql, err := quic.Listen(pconn, tlsConf, quicConfig)
	if err != nil {
		return nil, err
	}

	return quicnet.Listen(ql), nil
}

func NewServer(laddr string, opts ...options.ServerOption) (*grpc.Server, net.Listener, error) {
	m, err := ma.NewMultiaddr(laddr)
	if err != nil {
		return nil, nil, err
	}

	addr, err := newAddr(m)
	if err != nil {
		return nil, nil, err
	}

	cfg := options.NewServerConfig()
	if err := cfg.Apply(opts...); err != nil {
		return nil, nil, err
	}

	creds := new(transports.Credentials)
	l, err := NewListener(addr.String(), cfg.TLSConf)
	if err != nil {
		return nil, nil, err
	}

	return grpc.NewServer(grpc.Creds(creds)), l, err
}
