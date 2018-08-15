package opts

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/tap"
)

func withGrpcDialOptions(opts ...grpc.DialOption) DialOption {
	return func(o *ClientConfig) error {
		o.GrpcDialOptions = append(o.GrpcDialOptions, opts...)
		return nil
	}
}

func withGrpcServerOptions(opts ...grpc.ServerOption) ServerOption {
	return func(o *ServerConfig) error {
		o.GrpcServerOptions = append(o.GrpcServerOptions, opts...)
		return nil
	}
}

// WithWaitForHandshake blocks until the initial settings frame is received from
// the server before assigning RPCs to the connection. Experimental API.

// UNAVAILABLE
// func WithWaitForHandshake() DialOption {
// 	return withGrpcDialOptions(grpc.WithWaitForHandshake())
// }

// WithWriteBufferSize determines how much data can be batched before doing a
// write on the wire. The corresponding memory allocation for this buffer will
// be twice the size to keep syscalls low. The default value for this buffer is
// 32KB.
//
// Zero will disable the write buffer such that each write will be on underlying
// connection. Note: A Send call may not directly translate to a write.

// UNAVAILABLE
// func WithWriteBufferSize(s int) DialOption {
// 	return withGrpcDialOptions(grpc.WithWriteBufferSize(s))
// }

// WithReadBufferSize lets you set the size of read buffer, this determines how
// much data can be read at most for each read syscall.
//
// The default value for this buffer is 32KB. Zero will disable read buffer for
// a connection so data framer can access the underlying conn directly.

// UNAVAILABLE
// func WithReadBufferSize(s int) DialOption {
// 	return withGrpcDialOptions(grpc.WithReadBufferSize(s))
// }

// WithInitialWindowSize returns a DialOption which sets the value for initial
// window size on a stream. The lower bound for window size is 64K and any value
// smaller than that will be ignored.

// UNAVAILABLE
// func WithInitialWindowSize(s int32) DialOption {
// 	return withGrpcDialOptions(grpc.WithInitialWindowSize(s))
// }

// WithInitialConnWindowSize returns a DialOption which sets the value for
// initial window size on a connection. The lower bound for window size is 64K
// and any value smaller than that will be ignored.

// UNAVAILABLE
// func WithInitialConnWindowSize(s int32) DialOption {
// 	return withGrpcDialOptions(grpc.WithInitialConnWindowSize(s))
// }

// WithMaxMsgSize returns a DialOption which sets the maximum message size the
// client can receive.
//
// Deprecated: use WithDefaultCallOptions(MaxCallRecvMsgSize(s)) instead.

// UNAVAILABLE
// func WithMaxMsgSize(s int) DialOption {
// 	return withGrpcDialOptions(grpc.WithMaxMsgSize(s))
// }

// WithDefaultCallOptions returns a DialOption which sets the default
// CallOptions for calls over the connection.

// UNAVAILABLE
// func WithDefaultCallOptions(cos ...CallOption) DialOption {
// 	return withGrpcDialOptions(grpc.WithDefaultCallOptions(cos...))
// }

// WithCodec returns a DialOption which sets a codec for message marshaling and
// unmarshaling.
//
// Deprecated: use WithDefaultCallOptions(CallCustomCodec(c)) instead.
func WithCodec(c grpc.Codec) DialOption {
	return withGrpcDialOptions(grpc.WithCodec(c))
}

// WithCompressor returns a DialOption which sets a Compressor to use for
// message compression. It has lower priority than the compressor set by the
// UseCompressor CallOption.
//
// Deprecated: use UseCompressor instead.
func WithCompressor(cp grpc.Compressor) DialOption {
	return withGrpcDialOptions(grpc.WithCompressor(cp))
}

// WithDecompressor returns a DialOption which sets a Decompressor to use for
// incoming message decompression.  If incoming response messages are encoded
// using the decompressor's Type(), it will be used.  Otherwise, the message
// encoding will be used to look up the compressor registered via
// encoding.RegisterCompressor, which will then be used to decompress the
// message.  If no compressor is registered for the encoding, an Unimplemented
// status error will be returned.
//
// Deprecated: use encoding.RegisterCompressor instead.
func WithDecompressor(dc grpc.Decompressor) DialOption {
	return withGrpcDialOptions(grpc.WithDecompressor(dc))
}

// WithBalancer returns a DialOption which sets a load balancer with the v1 API.
// Name resolver will be ignored if this DialOption is specified.
//
// Deprecated: use the new balancer APIs in balancer package and
// WithBalancerName.

// UNAVAILABLE
// func WithBalancer(b Balancer) DialOption {
//      return withGrpcDialOptions(grpc.WithBalancer(b))
// }

// WithBalancerName sets the balancer that the ClientConn will be initialized
// with. Balancer registered with balancerName will be used. This function
// panics if no balancer was registered by balancerName.
//
// The balancer cannot be overridden by balancer option specified by service
// config.
//
// This is an EXPERIMENTAL API.
func WithBalancerName(balancerName string) DialOption {
	return withGrpcDialOptions(grpc.WithBalancerName(balancerName))
}

// withResolverBuilder is only for grpclb.

// UNAVAILABLE
// func withResolverBuilder(b resolver.Builder) DialOption {
// 	return withGrpcDialOptions(grpc.WithResolverBuilder(b))
// }

// WithServiceConfig returns a DialOption which has a channel to read the
// service configuration.
//
// Deprecated: service config should be received through name resolver, as
// specified here.
// https://github.com/grpc/grpc/blob/master/doc/service_config.md
func WithServiceConfig(c <-chan grpc.ServiceConfig) DialOption {
	return withGrpcDialOptions(grpc.WithServiceConfig(c))
}

// WithBackoffMaxDelay configures the dialer to use the provided maximum delay
// when backing off after failed connection attempts.
func WithBackoffMaxDelay(md time.Duration) DialOption {
	return withGrpcDialOptions(grpc.WithBackoffMaxDelay(md))
}

// WithBackoffConfig configures the dialer to use the provided backoff
// parameters after connection failures.
//
// Use WithBackoffMaxDelay until more parameters on BackoffConfig are opened up
// for use.
func WithBackoffConfig(b grpc.BackoffConfig) DialOption {
	return withGrpcDialOptions(grpc.WithBackoffConfig(b))
}

// WithBlock returns a DialOption which makes caller of Dial blocks until the
// underlying connection is up. Without this, Dial returns immediately and
// connecting the server happens in background.
func WithBlock() DialOption {
	return withGrpcDialOptions(grpc.WithBlock())
}

// WithInsecure returns a DialOption which disables transport security for this
// ClientConn. Note that transport security is required unless WithInsecure is
// set.

// UNAVAILABLE
// func WithInsecure() DialOption {
// 	return withGrpcDialOptions(grpc.WithInsecure())
// }

// WithTransportCredentials returns a DialOption which configures a connection
// level security credentials (e.g., TLS/SSL).

// UNAVAILABLE
// func WithTransportCredentials(creds credentials.TransportCredentials) DialOption {
// 	return withGrpcDialOptions(grpc.WithTransportCredentials(creds))
// }

// WithPerRPCCredentials returns a DialOption which sets credentials and places
// auth state on each outbound RPC.

// UNAVAILABLE
// func WithPerRPCCredentials(creds credentials.PerRPCCredentials) DialOption {
// 	return withGrpcDialOptions(grpc.WithPerRpcCredentials(creds))
// }

// WithTimeout returns a DialOption that configures a timeout for dialing a
// ClientConn initially. This is valid if and only if WithBlock() is present.
//
// Deprecated: use DialContext and context.WithTimeout instead.
func WithTimeout(d time.Duration) DialOption {
	return withGrpcDialOptions(grpc.WithTimeout(d))
}

// UNAVAILABLE
// func withContextDialer(f func(context.Context, string) (net.Conn, error)) DialOption {
// 	return withGrpcDialOptions(grpc.WithContextDialer(f))
// }

// WithDialer returns a DialOption that specifies a function to use for dialing
// network addresses. If FailOnNonTempDialError() is set to true, and an error
// is returned by f, gRPC checks the error's Temporary() method to decide if it
// should try to reconnect to the network address.

// UNAVAILABLE
// func WithDialer(f func(string, time.Duration) (net.Conn, error)) DialOption {
// 	return withGrpcDialOptions(grpc.WithDialer(f))
// }

// WithStatsHandler returns a DialOption that specifies the stats handler for
// all the RPCs and underlying network connections in this ClientConn.
func WithStatsHandler(h stats.Handler) DialOption {
	return withGrpcDialOptions(grpc.WithStatsHandler(h))
}

// FailOnNonTempDialError returns a DialOption that specifies if gRPC fails on
// non-temporary dial errors. If f is true, and dialer returns a non-temporary
// error, gRPC will fail the connection to the network address and won't try to
// reconnect. The default value of FailOnNonTempDialError is false.
//
// This is an EXPERIMENTAL API.
func FailOnNonTempDialError(f bool) DialOption {
	return withGrpcDialOptions(grpc.FailOnNonTempDialError(f))
}

// WithUserAgent returns a DialOption that specifies a user agent string for all
// the RPCs.

// UNAVAILABLE
// func WithUserAgent(s string) DialOption {
// 	return withGrpcDialOptions(grpc.WithUserAgent(s))
// }

// WithKeepaliveParams returns a DialOption that specifies keepalive parameters
// for the client transport.

// UNAVAILABLE
// func WithKeepaliveParams(kp keepalive.ClientParameters) DialOption {
// 	return withGrpcDialOptions(grpc.WithKeepaliveParams(kp))
// }

// WithUnaryInterceptor returns a DialOption that specifies the interceptor for
// unary RPCs.
func WithUnaryInterceptor(f grpc.UnaryClientInterceptor) DialOption {
	return withGrpcDialOptions(grpc.WithUnaryInterceptor(f))
}

// WithStreamInterceptor returns a DialOption that specifies the interceptor for
// streaming RPCs.
func WithStreamInterceptor(f grpc.StreamClientInterceptor) DialOption {
	return withGrpcDialOptions(grpc.WithStreamInterceptor(f))
}

// WithAuthority returns a DialOption that specifies the value to be used as the
// :authority pseudo-header. This value only works with WithInsecure and has no
// effect if TransportCredentials are present.
func WithAuthority(a string) DialOption {
	return withGrpcDialOptions(grpc.WithAuthority(a))
}

// WithChannelzParentID returns a DialOption that specifies the channelz ID of
// current ClientConn's parent. This function is used in nested channel creation
// (e.g. grpclb dial).
func WithChannelzParentID(id int64) DialOption {
	return withGrpcDialOptions(grpc.WithChannelzParentID(id))
}

// WithDisableServiceConfig returns a DialOption that causes grpc to ignore any
// service config provided by the resolver and provides a hint to the resolver
// to not fetch service configs.
func WithDisableServiceConfig() DialOption {
	return withGrpcDialOptions(grpc.WithDisableServiceConfig())
}

// WithDisableRetry returns a DialOption that disables retries, even if the
// service config enables them.  This does not impact transparent retries, which
// will happen automatically if no data is written to the wire or if the RPC is
// unprocessed by the remote server.
//
// Retry support is currently disabled by default, but will be enabled by
// default in the future.  Until then, it may be enabled by setting the
// environment variable "GRPC_GO_RETRY" to "on".
//
// This API is EXPERIMENTAL.
func WithDisableRetry() DialOption {
	return withGrpcDialOptions(grpc.WithDisableRetry())
}

// WithMaxHeaderListSize returns a DialOption that specifies the maximum
// (uncompressed) size of header list that the client is prepared to accept.

// UNAVAILABLE
// func WithMaxHeaderListSize(s uint32) DialOption {
// 	return withGrpcDialOptions(grpc.WithMaxHeaderListSize(s))
// }

// Server Option

// WriteBufferSize determines how much data can be batched before doing a write on the wire.
// The corresponding memory allocation for this buffer will be twice the size to keep syscalls low.
// The default value for this buffer is 32KB.
// Zero will disable the write buffer such that each write will be on underlying connection.
// Note: A Send call may not directly translate to a write.
func WriteBufferSize(s int) ServerOption {
	return withGrpcServerOptions(grpc.WriteBufferSize(s))
}

// ReadBufferSize lets you set the size of read buffer, this determines how much data can be read at most
// for one read syscall.
// The default value for this buffer is 32KB.
// Zero will disable read buffer for a connection so data framer can access the underlying
// conn directly.
func ReadBufferSize(s int) ServerOption {
	return withGrpcServerOptions(grpc.ReadBufferSize(s))
}

// InitialWindowSize returns a ServerOption that sets window size for stream.
// The lower bound for window size is 64K and any value smaller than that will be ignored.
func InitialWindowSize(s int32) ServerOption {
	return withGrpcServerOptions(grpc.InitialWindowSize(s))
}

// InitialConnWindowSize returns a ServerOption that sets window size for a connection.
// The lower bound for window size is 64K and any value smaller than that will be ignored.
func InitialConnWindowSize(s int32) ServerOption {
	return withGrpcServerOptions(grpc.InitialConnWindowSize(s))
}

// KeepaliveParams returns a ServerOption that sets keepalive and max-age parameters for the server.
func KeepaliveParams(kp keepalive.ServerParameters) ServerOption {
	return withGrpcServerOptions(grpc.KeepaliveParams(kp))
}

// KeepaliveEnforcementPolicy returns a ServerOption that sets keepalive enforcement policy for the server.
func KeepaliveEnforcementPolicy(kep keepalive.EnforcementPolicy) ServerOption {
	return withGrpcServerOptions(grpc.KeepaliveEnforcementPolicy(kep))
}

// CustomCodec returns a ServerOption that sets a codec for message marshaling and unmarshaling.
//
// This will override any lookups by content-subtype for Codecs registered with RegisterCodec.
func CustomCodec(codec grpc.Codec) ServerOption {
	return withGrpcServerOptions(grpc.CustomCodec(codec))
}

// RPCCompressor returns a ServerOption that sets a compressor for outbound
// messages.  For backward compatibility, all outbound messages will be sent
// using this compressor, regardless of incoming message compression.  By
// default, server messages will be sent using the same compressor with which
// request messages were sent.
//
// Deprecated: use encoding.RegisterCompressor instead.
func RPCCompressor(cp grpc.Compressor) ServerOption {
	return withGrpcServerOptions(grpc.RPCCompressor(cp))
}

// RPCDecompressor returns a ServerOption that sets a decompressor for inbound
// messages.  It has higher priority than decompressors registered via
// encoding.RegisterCompressor.
//
// Deprecated: use encoding.RegisterCompressor instead.
func RPCDecompressor(dc grpc.Decompressor) ServerOption {
	return withGrpcServerOptions(grpc.RPCDecompressor(dc))
}

// MaxMsgSize returns a ServerOption to set the max message size in bytes the server can receive.
// If this is not set, gRPC uses the default limit.
//
// Deprecated: use MaxRecvMsgSize instead.
func MaxMsgSize(m int) ServerOption {
	return withGrpcServerOptions(grpc.MaxMsgSize(m))
}

// MaxRecvMsgSize returns a ServerOption to set the max message size in bytes the server can receive.
// If this is not set, gRPC uses the default 4MB.
func MaxRecvMsgSize(m int) ServerOption {
	return withGrpcServerOptions(grpc.MaxRecvMsgSize(m))
}

// MaxSendMsgSize returns a ServerOption to set the max message size in bytes the server can send.
// If this is not set, gRPC uses the default 4MB.
func MaxSendMsgSize(m int) ServerOption {
	return withGrpcServerOptions(grpc.MaxSendMsgSize(m))
}

// MaxConcurrentStreams returns a ServerOption that will apply a limit on the number
// of concurrent streams to each ServerTransport.
func MaxConcurrentStreams(n uint32) ServerOption {
	return withGrpcServerOptions(grpc.MaxConcurrentStreams(n))
}

// Creds returns a ServerOption that sets credentials for server connections.

// UNAVAILABLE
// func Creds(c credentials.TransportCredentials) ServerOption {
// 	return withGrpcServerOptions(grpc.Creds(c))
// }

// UnaryInterceptor returns a ServerOption that sets the UnaryServerInterceptor for the
// server. Only one unary interceptor can be installed. The construction of multiple
// interceptors (e.g., chaining) can be implemented at the caller.
func UnaryInterceptor(i grpc.UnaryServerInterceptor) ServerOption {
	return withGrpcServerOptions(grpc.UnaryInterceptor(i))
}

// StreamInterceptor returns a ServerOption that sets the StreamServerInterceptor for the
// server. Only one stream interceptor can be installed.
func StreamInterceptor(i grpc.StreamServerInterceptor) ServerOption {
	return withGrpcServerOptions(grpc.StreamInterceptor(i))
}

// InTapHandle returns a ServerOption that sets the tap handle for all the server
// transport to be created. Only one can be installed.
func InTapHandle(h tap.ServerInHandle) ServerOption {
	return withGrpcServerOptions(grpc.InTapHandle(h))
}

// StatsHandler returns a ServerOption that sets the stats handler for the server.
func StatsHandler(h stats.Handler) ServerOption {
	return withGrpcServerOptions(grpc.StatsHandler(h))
}

// UnknownServiceHandler returns a ServerOption that allows for adding a custom
// unknown service handler. The provided method is a bidi-streaming RPC service
// handler that will be invoked instead of returning the "unimplemented" gRPC
// error whenever a request is received for an unregistered service or method.
// The handling function has full access to the Context of the request and the
// stream, and the invocation bypasses interceptors.

// UNAVAILABLE
// func UnknownServiceHandler(streamHandler StreamHandler) ServerOption {
// 	return withGrpcServerOptions(grpc.UnknownServiceHandler(streamHandler))
// }

// ConnectionTimeout returns a ServerOption that sets the timeout for
// connection establishment (up to and including HTTP/2 handshaking) for all
// new connections.  If this is not set, the default is 120 seconds.  A zero or
// negative value will result in an immediate timeout.
//
// This API is EXPERIMENTAL.
func ConnectionTimeout(d time.Duration) ServerOption {
	return withGrpcServerOptions(grpc.ConnectionTimeout(d))
}

// MaxHeaderListSize returns a ServerOption that sets the max (uncompressed) size
// of header list that the server is prepared to accept.
func MaxHeaderListSize(s uint32) ServerOption {
	return withGrpcServerOptions(grpc.MaxHeaderListSize(s))
}
