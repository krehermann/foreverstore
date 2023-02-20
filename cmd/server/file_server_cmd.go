package main

import (
	"context"
	"net"

	"github.com/krehermann/foreverstore/fileserver"
	"github.com/krehermann/foreverstore/p2p"
	"go.uber.org/zap"
)

type FileServerCmd struct {
	Decoder   string   `help:"rpc protocol to use"`
	Addr      string   `help:"address to listen on"`
	Bootstrap []string `help:"bootstrap addresses"`
	//logger    *zap.Logger
}

func (s *FileServerCmd) Run() error {
	l, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	bootStrapAddrs := make([]net.Addr, 0)
	for _, b := range s.Bootstrap {
		l.Sugar().Debugf("bootstrap add %s", b)
		bootStrapAddrs = append(bootStrapAddrs, p2p.TCPTransportAddr{Addr: b})
	}

	err = startBootstraps(bootStrapAddrs, l)
	if err != nil {
		return err
	}
	opts := fileserver.FileServerOpts{
		Logger:     l,
		Bootstraps: bootStrapAddrs,
		ListenAddr: s.Addr,
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

func startBootstraps(addrs []net.Addr, logger *zap.Logger) error {
	opts := fileserver.FileServerOpts{
		Logger: logger,
	}

	for _, addr := range addrs {

		opts.ListenAddr = addr.String()
		bootStrapServer, err := fileserver.NewFileServer(opts)
		if err != nil {
			return err
		}
		return bootStrapServer.Start(context.Background())
	}
	return nil
}
