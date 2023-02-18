package main

import (
	"context"

	"github.com/krehermann/foreverstore/fileserver"
	"go.uber.org/zap"
)

type FileServerCmd struct {
	Decoder string `help:"rpc protocol to use"`
	Addr    string `help:"address to listen on"`
	logger  *zap.Logger
}

func (s *FileServerCmd) Run() error {
	l, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	opts := fileserver.FileServerOpts{
		Logger: l,
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
