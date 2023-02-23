package p2p

import (
	"context"
	"io"
	"net"

	"github.com/krehermann/foreverstore/types"
	"github.com/krehermann/foreverstore/util"
	"go.uber.org/zap"
)

var _ Transport = (*TcpTransport)(nil)

type TcpTransport struct {
	addr     *net.TCPAddr
	listener net.Listener

	// connections our listener excepted
	incoming *util.ConcurrentMap[*types.ComparableAddr, net.Conn]
	outgoing *util.ConcurrentMap[*types.ComparableAddr, net.Conn]

	config TcpTransportConfig
	logger *zap.Logger

	rpcCh chan RPC
}

type TcpOpt func(*TcpTransport)

func TcpOptWithLogger(l *zap.Logger) TcpOpt {
	return func(u *TcpTransport) {
		u.logger = l
	}
}

type TcpTransportConfig struct {
	Handshaker Handshaker
	ProtocolFactoryFunc
	PeerHandler
}

func NewTcpTransport(listenAddr string, config TcpTransportConfig, opts ...TcpOpt) (*TcpTransport, error) {
	a, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	if config.ProtocolFactoryFunc == nil {
		config.ProtocolFactoryFunc = NewBinaryProtocolDecoder
	}
	if config.Handshaker == nil {
		config.Handshaker = NOPHandshake{}
	}
	u := &TcpTransport{
		addr:     a,
		incoming: util.NewConcurrentMap[*types.ComparableAddr, net.Conn](),
		outgoing: util.NewConcurrentMap[*types.ComparableAddr, net.Conn](),
		config:   config,
		rpcCh:    make(chan RPC),
	}

	for _, opt := range opts {
		opt(u)
	}

	return u, nil
}

func (u *TcpTransport) Recv() <-chan RPC {
	return u.rpcCh
}

func (u *TcpTransport) Close() error {
	close(u.rpcCh)
	return u.listener.Close()
}

func (u *TcpTransport) Dial(network, address string) (Peer, error) {

	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return remotePeer{Conn: conn}, nil
}

func (u *TcpTransport) Addr() net.Addr {
	return u.listener.Addr()
}

func (u *TcpTransport) Listen(ctx context.Context) error {

	var err error

	u.listener, err = net.Listen(u.addr.Network(), u.addr.String())
	if err != nil {
		return err
	}

	u.logger.Sugar().Infof("Listening at %+v", u.listener.Addr())
	go accept(ctx, u.listener, u.logger, u.handleConn)

	return nil
}

func (u *TcpTransport) handleConn(conn net.Conn) error {
	u.logger.Debug("new connection",
		zap.String("local", conn.LocalAddr().String()),
		zap.String("remote", conn.RemoteAddr().String()),
	)

	defer conn.Close()
	peer := remotePeer{Conn: conn}
	if u.config.PeerHandler != nil {
		err := u.config.PeerHandler(peer)
		if err != nil {

			return err
		}
	}

	err := u.config.Handshaker.Handshake(conn)
	if err != nil {
		u.logger.Error("handshake failed. closing connection", zap.Error(err))
		return err
	}
	d := u.config.ProtocolFactoryFunc(conn, u.logger)
	for {
		var rpc RPC
		err := d.Decode(&rpc)
		if err != nil {
			if err == io.EOF {
				break
			}
			u.logger.Error("decode error", zap.Error(err))
			return err
		}
		rpc.From = conn.RemoteAddr()
		u.rpcCh <- rpc
		u.logger.Debug("got rpc", zap.Any("raw", rpc), zap.String("payload", string(rpc.payload)))
	}

	return nil
}

type TCPTransportAddr struct {
	Addr string
}

var _ net.Addr = TCPTransportAddr{}

func (a TCPTransportAddr) Network() string {
	return "tcp"
}
func (a TCPTransportAddr) String() string {
	return a.Addr
}
