package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	gquic "github.com/gfanton/grpc-quic"
	"github.com/gfanton/grpc-quic/example/proto/hello"
)

func main() {
	if len(os.Args) == 1 {
		panic("Not enough arguments")
	}

	addr := os.Args[1]
	fmt.Println("dialing " + addr)

	tlsConf := &tls.Config{InsecureSkipVerify: true}
	t, err := gquic.NewGrpcQuicTransport("localhost:0", tlsConf)
	if err != nil {
		panic("Transport error " + err.Error())
	}

	c, err := t.GrpcDial(addr)
	if err != nil {
		panic("Dial error " + err.Error())
	}

	greet := hello.NewGreeterClient(c)

	req := &hello.HelloRequest{}
	req.Name = "World"

	rep, err := greet.SayHello(context.Background(), req)
	if err != nil {
		panic("SayHello error " + err.Error())
	}

	fmt.Println("Response " + rep.GetMessage())
}
