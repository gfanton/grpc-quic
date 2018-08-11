package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"

	gquic "github.com/gfanton/grpc-quic"
	"github.com/gfanton/grpc-quic/example/proto/hello"
)

type Hello struct{}

func (h *Hello) SayHello(ctx context.Context, in *hello.HelloRequest) (*hello.HelloReply, error) {
	fmt.Println("Receiving " + in.GetName())
	rep := &hello.HelloReply{}
	rep.Message = "Hello " + in.GetName()
	fmt.Println("Sending " + rep.GetMessage())
	return rep, nil
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}

func main() {
	if len(os.Args) == 1 {
		panic("Not enough arguments")
	}

	addr := os.Args[1]

	tlsConf := generateTLSConfig()
	t, err := gquic.NewGrpcQuicTransport(addr, tlsConf)
	if err != nil {
		panic("Transport error " + err.Error())
	}

	gs := t.NewGrpcServer()
	hello.RegisterGreeterServer(gs, &Hello{})

	fmt.Println("Listening on " + addr)
	if err := t.GrpcServe(gs); err != nil {
		panic("Serve error " + err.Error())
	}
}
