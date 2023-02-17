package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/krehermann/foreverstore/p2p"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("Good to go")

	l, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalln(err)
	}
	f := "/tmp/unix-socket-test-1"
	defer os.Remove(f)
	t, err := p2p.NewTcpTransport(f,
		p2p.TcpTransportConfig{
			Handshaker:          p2p.NOPHandshake{},
			ProtocolFactoryFunc: p2p.NewNewlineDecoder,
		},
		p2p.TcpOptWithLogger(l))

	if err != nil {
		l.Sugar().Fatal(err)
	}
	err = t.Listen(context.Background())
	if err != nil {
		l.Sugar().Fatalf("error listening %s", err)
	}
	time.Sleep(300 * time.Second)
}
