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
	lggr   *zap.Logger
	quitCh chan struct{}
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
		quitCh:         make(chan struct{}),
	}

	return fs, nil
}

func (s *FileServer) Start(ctx context.Context) error {
	s.lggr.Sugar().Info("Starting...")
	err := s.Transport.Listen(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *FileServer) Stop(ctx context.Context) error {
	close(s.quitCh)
	return nil
}

func (s *FileServer) handleProtocol(ctx context.Context) {
	defer s.Transport.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.quitCh:
			return
		case msg := <-s.Transport.Recv():
			s.lggr.Sugar().Debugf("recieved msg: %+v", msg)
		}
	}

}
