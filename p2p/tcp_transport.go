package p2p

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/krehermann/foreverstore/types"
	"github.com/krehermann/foreverstore/util"
	"go.uber.org/zap"
)

var _ Transport = (*TcpTransport)(nil)

type TcpTransport struct {
	//addr string
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
	Handshaker       Handshaker
	AllowAnonynomous bool
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

func (u *TcpTransport) Recv() chan<- RPC {
	return u.rpcCh
}

func (u *TcpTransport) Listen(ctx context.Context) error {
	var err error

	u.listener, err = net.Listen(u.addr.Network(), u.addr.String())
	if err != nil {
		return err
	}

	go accept(ctx, u.listener, u.handleConn)

	return nil
}

func (u *TcpTransport) handleConn(conn net.Conn) error {
	u.logger.Debug("new connection",
		zap.String("local", conn.LocalAddr().String()),
		zap.String("remote", conn.RemoteAddr().String()),
	)

	if !u.config.AllowAnonynomous && conn.RemoteAddr().String() == "@" {
		u.logger.Error("refusing anonyomous connection", zap.Bool("allow", u.config.AllowAnonynomous))
		return fmt.Errorf("no peer in incoming unix connection %s", conn.RemoteAddr().String())
	}

	if u.config.PeerHandler != nil {
		err := u.config.PeerHandler(conn)
		if err != nil {

			return err
		}
	}
	//raddr := types.NewComparableAddr(conn.RemoteAddr())
	/*
		err := u.incoming.Put(raddr, conn)
		if err != nil {
			return err
		}

		defer func() {
			u.incoming.Delete(raddr)
			conn.Close()
		}()
	*/
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
		u.logger.Debug("got rpc", zap.Any("raw", rpc), zap.String("payload", string(rpc.Payload)))
	}

	return nil
}
