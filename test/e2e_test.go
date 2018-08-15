package test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	qgrpc "github.com/gfanton/grpc-quic"
	quicbalancer "github.com/gfanton/grpc-quic/balancer"
	"github.com/gfanton/grpc-quic/opts"
	"github.com/gfanton/grpc-quic/proto/hello"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
)

type Hello struct{}

func (h *Hello) SayHello(ctx context.Context, in *hello.HelloRequest) (*hello.HelloReply, error) {
	rep := new(hello.HelloReply)
	rep.Message = "Hello " + in.GetName()
	return rep, nil
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}, nil
}

func testDial(t *testing.T, target string) {
	var (
		client *grpc.ClientConn
		server *grpc.Server

		err error
	)

	defer func() {
		if client != nil {
			client.Close()
		}

		if server != nil {
			server.Stop()
		}
	}()

	Convey("Setup server", t, func(c C) {
		//setup server
		tlsConf, err := generateTLSConfig()
		So(err, ShouldBeNil)

		server, l, err := qgrpc.NewServer(target, opts.TLSConfig(tlsConf))
		So(err, ShouldBeNil)

		hello.RegisterGreeterServer(server, &Hello{})

		go func() {
			err := server.Serve(l)
			c.So(err, ShouldBeNil)
		}()
	})

	Convey("Setup client", t, func() {
		tlsConf := &tls.Config{InsecureSkipVerify: true}

		// Take a random port to listen from udp server
		client, err = qgrpc.Dial(target, opts.WithTLSConfig(tlsConf))
		So(err, ShouldBeNil)
	})

	Convey("Test basic dial", t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		greet := hello.NewGreeterClient(client)
		req := new(hello.HelloRequest)
		req.Name = "World"

		rep, err := greet.SayHello(ctx, req)
		So(err, ShouldBeNil)
		So(rep.GetMessage(), ShouldEqual, "Hello World")
	})
}

func TestDialUDP(t *testing.T) {
	target := "/ip4/127.0.0.1/udp/5847"
	testDial(t, target)
}

func TestDialTCP(t *testing.T) {
	target := "/ip4/127.0.0.1/tcp/5847"
	testDial(t, target)
}

type testHandler func(*manual.Resolver, hello.GreeterClient, []*grpc.Server)

func testBalancerProgressiveClose(mresolver *manual.Resolver, client hello.GreeterClient, servers []*grpc.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration((len(servers)+1)))
	defer cancel()

	req := new(hello.HelloRequest)
	req.Name = "World"

	for i, s := range servers {
		s.Stop()

		// Wait for shutdown
		time.Sleep(time.Microsecond * 5)

		rep, err := client.SayHello(ctx, req)
		if i == len(servers)-1 {
			So(err, ShouldNotBeNil)
		} else {
			So(err, ShouldBeNil)
			So(rep.GetMessage(), ShouldEqual, "Hello World")
		}
	}
}

func testBalancerDial(mresolver *manual.Resolver, client hello.GreeterClient, servers []*grpc.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := new(hello.HelloRequest)
	req.Name = "World"

	rep, err := client.SayHello(ctx, req)
	So(err, ShouldBeNil)
	So(rep.GetMessage(), ShouldEqual, "Hello World")
}

func testBalancer(t *testing.T, balancerName string, clientAddrs []string, serverAddrs []string, handlers ...testHandler) {
	var (
		client    *grpc.ClientConn
		servers   []*grpc.Server
		mresolver *manual.Resolver
		rcleanup  func()

		err error
	)

	// cleanup
	defer func() {
		if client != nil {
			client.Close()
		}

		for _, server := range servers {
			if server != nil {
				server.Stop()
			}
		}

		if rcleanup != nil {
			rcleanup()
		}

	}()

	Convey("Setup servers", t, func(c C) {
		//setup server
		tlsConf, err := generateTLSConfig()
		So(err, ShouldBeNil)

		for _, addr := range serverAddrs {
			server, l, err := qgrpc.NewServer(addr, opts.TLSConfig(tlsConf))
			So(err, ShouldBeNil)

			hello.RegisterGreeterServer(server, &Hello{})

			go func() {
				err := server.Serve(l)
				c.So(err, ShouldBeNil)
			}()

			servers = append(servers, server)
		}
	})

	Convey("Setup client", t, func() {
		var addrs []resolver.Address
		for _, addr := range clientAddrs {
			addrs = append(addrs, resolver.Address{Addr: addr})
		}

		tlsConf := &tls.Config{InsecureSkipVerify: true}
		mresolver, rcleanup = manual.GenerateAndRegisterManualResolver()

		// Take a random port to listen from udp server
		client, err = qgrpc.Dial(mresolver.Scheme()+":///",
			opts.WithTLSConfig(tlsConf),
			opts.WithBalancerName(balancerName),
		)
		So(err, ShouldBeNil)

		mresolver.NewAddress(addrs)
	})

	Convey("Test handler", t, func() {
		for _, th := range handlers {
			greet := hello.NewGreeterClient(client)
			th(mresolver, greet, servers)
		}
	})
}

func TestBalancerTCPtoUDP(t *testing.T) {
	balancerName := quicbalancer.Name
	serverAddrs := []string{
		// UDP servers
		"/ip4/127.0.0.1/udp/6850",
	}

	clientAddrs := []string{
		// Fake (down)
		"/ip4/127.0.0.1/tcp/6950",

		// Real (up)
		"/ip4/127.0.0.1/udp/6850",
	}

	testBalancer(t, balancerName, clientAddrs, serverAddrs, testBalancerDial)
}

func TestBalancerUDPtoTCP(t *testing.T) {
	balancerName := quicbalancer.Name
	serverAddrs := []string{
		// UDP servers
		"/ip4/127.0.0.1/tcp/6851",
	}

	clientAddrs := []string{
		// Fake (down)
		"/ip4/127.0.0.1/udp/6951",

		// Real (up)
		"/ip4/127.0.0.1/tcp/6851",
	}

	testBalancer(t, balancerName, clientAddrs, serverAddrs, testBalancerDial)
}

func TestBalancerProgressiveClose(t *testing.T) {
	balancerName := quicbalancer.Name
	addrs := []string{
		"/ip4/127.0.0.1/tcp/6852",
		"/ip4/127.0.0.1/udp/6853",
		"/ip4/127.0.0.1/tcp/6854",

		"/ip4/127.0.0.1/udp/6855",
		"/ip4/127.0.0.1/tcp/6856",
		"/ip4/127.0.0.1/udp/6857",
	}

	testBalancer(t, balancerName, addrs, addrs, testBalancerProgressiveClose)
}
