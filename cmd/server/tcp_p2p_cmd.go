package main

import (
	"context"
	"time"

	"github.com/krehermann/foreverstore/p2p"
	"go.uber.org/zap"
)

type TCPP2PServerCmd struct {
	Decoder string `help:"rpc protocol to use"`
	Addr    string `help:"address to listen on"`
	logger  *zap.Logger
}

func (s *TCPP2PServerCmd) Run() error {
	l, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	t, err := p2p.NewTcpTransport(s.Addr,
		p2p.TcpTransportConfig{
			Handshaker: p2p.NOPHandshake{},
		},
		p2p.TcpOptWithLogger(l),
	)

	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 2*time.Minute)
	err = t.Listen(ctx)
	if err != nil {
		return err
	}

	time.Sleep(3 * time.Minute)
	return nil
}
