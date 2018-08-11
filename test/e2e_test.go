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

	gq "github.com/gfanton/grpc-quic"
	"github.com/gfanton/grpc-quic/example/proto/hello"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
)

const ADDR = "127.0.0.1:5678"

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

func TestDial(t *testing.T) {
	var (
		client *grpc.ClientConn
		server *grpc.Server

		tlsclient *tls.Config
		tlsserver *tls.Config

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
		tlsserver, err = generateTLSConfig()
		So(err, ShouldBeNil)

		ts, err := gq.NewGrpcQuicTransport(ADDR, tlsserver)
		So(err, ShouldBeNil)

		server = ts.NewGrpcServer()
		hello.RegisterGreeterServer(server, &Hello{})

		go func() {
			err := ts.GrpcServe(server)
			c.So(err, ShouldBeNil)
		}()
	})

	Convey("Setup client", t, func() {
		tlsclient = &tls.Config{InsecureSkipVerify: true}

		// Take a random port to listen from udp server
		tc, err := gq.NewGrpcQuicTransport("localhost:0", tlsclient)
		So(err, ShouldBeNil)

		client, err = tc.GrpcDial(ADDR)
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
