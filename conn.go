package grpcquic

import (
	"context"
	"fmt"
	"net"
	"time"

	quic "github.com/lucas-clemente/quic-go"
	"google.golang.org/grpc/credentials"
)

var _ net.Conn = (*Conn)(nil)

type Conn struct {
	sess   quic.Session
	stream quic.Stream
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *Conn) Read(b []byte) (n int, err error) {
	return c.stream.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *Conn) Write(b []byte) (n int, err error) {
	return c.stream.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *Conn) Close() error {
	// @TODO: log this
	c.stream.Close()

	return c.sess.Close()
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.sess.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.sess.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c *Conn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)

}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

var _ net.Listener = (*Listener)(nil)

type Listener struct {
	ql quic.Listener
}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	sess, err := l.ql.Accept()
	if err != nil {
		return nil, err
	}

	s, err := sess.AcceptStream()
	if err != nil {
		return nil, err
	}

	return &Conn{sess, s}, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *Listener) Close() error {
	return l.ql.Close()
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.ql.Addr()
}

// Info is a credentials.AuthInfo

var _ credentials.AuthInfo = (*Info)(nil)

// Info contains the auth information
type Info struct {
	conn *Conn
}

func NewInfo(c *Conn) *Info {
	return &Info{c}
}

// AuthType returns the type of Info as a string.
func (i *Info) AuthType() string {
	return "quic-tls"
}

func (i *Info) Conn() *Conn {
	return i.conn
}

var _ credentials.TransportCredentials = (*TransportCredentials)(nil)

type TransportCredentials struct{}

var p2pinfo Info = Info{}

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
func (pt *TransportCredentials) ClientHandshake(ctx context.Context, authority string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	fmt.Print("Client Handshake...")
	if c, ok := conn.(*Conn); ok {
		ainfo := NewInfo(c)
		return conn, ainfo, nil
	}
	fmt.Println("Done")
	return nil, nil, fmt.Errorf("Not a valid quic conn")
}

// ServerHandshake does the authentication handshake for servers. It returns
// the authenticated connection and the corresponding auth information about
// the connection.
//
// If the returned net.Conn is closed, it MUST close the net.Conn provided.
func (pt *TransportCredentials) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {

	if c, ok := conn.(*Conn); ok {
		ainfo := NewInfo(c)
		return conn, ainfo, nil
	}

	return nil, nil, fmt.Errorf("Not a valid p2p conn")
}

// Info provides the ProtocolInfo of this TransportCredentials.
func (pt *TransportCredentials) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{
		// ProtocolVersion is the gRPC wire protocol version.
		ProtocolVersion: "/quic/1.0.0",
		// SecurityProtocol is the security protocol in use.
		SecurityProtocol: "quic-tls",
		// SecurityVersion is the security protocol version.
		SecurityVersion: "1.2.0",
		// ServerName is the user-configured server name.
		ServerName: "test",
	}
}

// Clone makes a copy of this TransportCredentials.
func (pt *TransportCredentials) Clone() credentials.TransportCredentials {
	return &TransportCredentials{}
}

// OverrideServerName overrides the server name used to verify the hostname on the returned certificates from the server.
// gRPC internals also use it to override the virtual hosting name if it is set.
// It must be called before dialing. Currently, this is only used by grpclb.
func (pt *TransportCredentials) OverrideServerName(name string) error {
	return nil
}
