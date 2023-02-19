package main

import (
	"context"
	"net"

	"github.com/krehermann/foreverstore/fileserver"
	"github.com/krehermann/foreverstore/p2p"
	"github.com/krehermann/foreverstore/util"
	"go.uber.org/zap"
)

type FileServerCmd struct {
	Decoder   string   `help:"rpc protocol to use"`
	Addr      string   `help:"address to listen on"`
	Bootstrap []string `help:"bootstrap addresses"`
	logger    *zap.Logger
}

func (s *FileServerCmd) Run() error {
	l, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	addrs := make([]net.Addr, 0)
	for _, b := range s.Bootstrap {
		l.Sugar().Debugf("bootstrap add %s", b)
		addrs = append(addrs, p2p.TCPTransportAddr{Addr: b})
	}

	opts := fileserver.FileServerOpts{
		Logger:     l,
		Bootstraps: util.NewIterable[net.Addr](addrs),
	}

	srvr, err := fileserver.NewFileServer(opts)
	if err != nil {
		return err
	}

	err = srvr.Start(context.Background())
	if err != nil {
		return err
	}

	// hack. should handle signals
	waitForever := make(chan struct{})
	<-waitForever
	return nil
}
