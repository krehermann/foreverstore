package fileserver

import (
	"context"

	"github.com/krehermann/foreverstore/p2p"
	"github.com/krehermann/foreverstore/store"
	"go.uber.org/zap"
)

type FileServerOpts struct {
	Logger     *zap.Logger
	ListenAddr string
	Store      store.ReadWriteStatFS
	Transport  p2p.Transport
	// StorageRoot string
	// PathTransformFunc store.PathFunc
}

type FileServer struct {
	FileServerOpts
	//store store.ReadWriteStatFS
	lggr *zap.Logger

	// root string
}

func NewFileServer(opts FileServerOpts) (*FileServer, error) {
	// setup logger
	if opts.Logger == nil {
		l, err := zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
		opts.Logger = l
	}
	lggr := opts.Logger.Named("FileServer")
	// setup default store
	if opts.Store == nil {
		str, err := store.NewBlobStore(
			store.BlobStoreConfig{
				PathFunc: store.ContentPath,
				Logger:   lggr,
			},
		)
		if err != nil {
			return nil, err
		}
		opts.Store = str
	}
	// setup default transport
	if opts.Transport == nil {
		tcpTransport, err := p2p.NewTcpTransport(
			opts.ListenAddr,
			p2p.TcpTransportConfig{},
			p2p.TcpOptWithLogger(lggr),
		)
		if err != nil {
			return nil, err
		}
		opts.Transport = tcpTransport
	}

	fs := &FileServer{
		FileServerOpts: opts,
		lggr:           lggr,
	}

	return fs, nil
}

func (s *FileServer) Start(ctx context.Context) error {
	err := s.Transport.Listen(ctx)
	if err != nil {
		return err
	}
}
