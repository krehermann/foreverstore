package main

import (
	"context"
	"time"

	"github.com/alecthomas/kong"
	"github.com/krehermann/foreverstore/p2p"
	"go.uber.org/zap"
)

type debugFlag bool

func (d debugFlag) BeforeApply(logger *zap.SugaredLogger) error {
	l, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	logger = l.Sugar()
	return nil
}

var CLI struct {
	Debug debugFlag `help:"Enable debug logging."`

	Serve ServerCmd `cmd:"" help:"start server"`
}

func main() {
	logger := zap.NewNop().Sugar()
	ctx := kong.Parse(&CLI, kong.Bind(logger))

	// Call the Run() method of the selected parsed command.
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

type ServerCmd struct {
	Decoder string `help:"rpc protocol to use"`
	Addr    string `help:"address to listen on"`
	logger  *zap.Logger
}

func (s *ServerCmd) Run() error {
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
