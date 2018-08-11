package transports

import (
	"context"
	"crypto/tls"
	"net"

	quicnet "github.com/gfanton/grpc-quic/net"
	"google.golang.org/grpc/credentials"
)

var _ credentials.AuthInfo = (*Info)(nil)

// Info contains the auth information
type Info struct {
	conn *quicnet.Conn
}

func NewInfo(c *quicnet.Conn) *Info {
	return &Info{c}
}

// AuthType returns the type of Info as a string.
func (i *Info) AuthType() string {
	return "quic-tls"
}

func (i *Info) Conn() net.Conn {
	return i.conn
}

var _ credentials.TransportCredentials = (*Credentials)(nil)

type Credentials struct {
	grpcCreds credentials.TransportCredentials
	tlsConfig *tls.Config
}

func NewCredentials(tlsConfig *tls.Config) credentials.TransportCredentials {
	grpcCreds := credentials.NewTLS(tlsConfig)
	return &Credentials{grpcCreds, tlsConfig}
}

// ClientHandshake does the authentication handshake specified by the corresponding
// authentication protocol on rawConn for clients. It returns the authenticated
// connection and the corresponding auth information about the connection.
// Implementations must use the provided context to implement timely cancellation.
// gRPC will try to reconnect if the error returned is a temporary error
// (io.EOF, context.DeadlineExceeded or err.Temporary() == true).
// If the returned error is a wrapper error, implementations should make sure that
// the error implements Temporary() to have the correct retry behaviors.
//
// If the returned net.Conn is closed, it MUST close the net.Conn provided.
func (pt *Credentials) ClientHandshake(ctx context.Context, authority string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	if c, ok := conn.(*quicnet.Conn); ok {
		return conn, NewInfo(c), nil
	}

	return pt.grpcCreds.ClientHandshake(ctx, authority, conn)
}

// ServerHandshake does the authentication handshake for servers. It returns
// the authenticated connection and the corresponding auth information about
// the connection.
//
// If the returned net.Conn is closed, it MUST close the net.Conn provided.
func (pt *Credentials) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	if c, ok := conn.(*quicnet.Conn); ok {
		ainfo := NewInfo(c)
		return conn, ainfo, nil
	}

	return pt.grpcCreds.ServerHandshake(conn)
}

// Info provides the ProtocolInfo of this Credentials.
func (pt *Credentials) Info() credentials.ProtocolInfo {
	return pt.grpcCreds.Info()
}

// Clone makes a copy of this Credentials.
func (pt *Credentials) Clone() credentials.TransportCredentials {
	return &Credentials{
		tlsConfig: pt.tlsConfig.Clone(),
		grpcCreds: pt.grpcCreds.Clone(),
	}
}

// OverrideServerName overrides the server name used to verify the hostname on the returned certificates from the server.
// gRPC internals also use it to override the virtual hosting name if it is set.
// It must be called before dialing. Currently, this is only used by grpclb.
func (pt *Credentials) OverrideServerName(name string) error {
	return nil
}
