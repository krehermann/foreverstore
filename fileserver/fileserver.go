package fileserver

import (
	"context"
	"fmt"
	"net"

	"github.com/krehermann/foreverstore/p2p"
	"github.com/krehermann/foreverstore/store"
	"github.com/krehermann/foreverstore/util"
	"go.uber.org/zap"
)

type FileServerOpts struct {
	Logger     *zap.Logger
	ListenAddr string
	Store      store.ReadWriteStatFS
	Transport  p2p.Transport
	Bootstraps []net.Addr //*util.Iterable[net.Addr]

	// PathTransformFunc store.PathFunc
}

type FileServer struct {
	FileServerOpts
	lggr   *zap.Logger
	quitCh chan struct{}

	peers *util.ConcurrentMap[string, p2p.Peer]
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
	lggr := opts.Logger.Named(fmt.Sprintf("FileServer%s", opts.ListenAddr))
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
		peers:          util.NewConcurrentMap[string, p2p.Peer](),
	}

	return fs, nil
}

func (s *FileServer) Start(ctx context.Context) error {
	s.lggr.Sugar().Info("Starting...")
	err := s.Transport.Listen(ctx)
	if err != nil {
		return err
	}
	return s.bootstrap()
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

func (s *FileServer) bootstrap() error {
	s.lggr.Sugar().Debug("bootstrapping...")
	defer s.lggr.Sugar().Debug("done bootstrapping...")
	if s.Bootstraps == nil {
		return nil
	}
	for _, boot := range s.Bootstraps {
		s.lggr.Sugar().Debugf("dialing %s:%s", boot.Network(), boot.String())
		peer, err := s.Transport.Dial(boot.Network(), boot.String())
		if err != nil {
			panic(err)
		}
		s.lggr.Sugar().Debugf("added peer %s", boot.String())
		s.peers.Put(boot.String(), peer)
	}

	return nil
}
