package p2p

import (
	"context"
	"fmt"
	"net"

	"github.com/krehermann/foreverstore/types"
	"github.com/krehermann/foreverstore/util"
	"go.uber.org/zap"
)

var _ Peer = (*UnixPeer)(nil)

type UnixPeer struct {
	id   string
	conn net.Conn
}

func NewUnixPeer(conn net.Conn, id string) *UnixPeer {
	return &UnixPeer{
		id:   id,
		conn: conn,
	}
}

func (p *UnixPeer) Close() error {
	return p.conn.Close()
}

type UnixTransport struct {
	//addr string
	addr     *net.UnixAddr
	listener net.Listener

	// connections our listener excepted
	incoming *util.ConcurrentMap[*types.ComparableAddr, net.Conn] //map[net.Addr]net.Conn
	// connections this transport has dialed
	outgoing *util.ConcurrentMap[*types.ComparableAddr, net.Conn]

	//Handshaker
	//	decoder Decoder
	config UnixTransportConfig
	logger *zap.Logger
}

type UnixOpt func(*UnixTransport)

func UnixOptWithLogger(l *zap.Logger) UnixOpt {
	return func(u *UnixTransport) {
		u.logger = l
	}
}

type UnixTransportConfig struct {
	Handshaker Handshaker
	//RPCDecoder    Decoder
	AllowAnonynomous bool
}

func NewUnixTransport(listenAddr string, config UnixTransportConfig, opts ...UnixOpt) (*UnixTransport, error) {
	a, err := net.ResolveUnixAddr("unix", listenAddr)
	if err != nil {
		return nil, err
	}

	u := &UnixTransport{
		addr:     a,
		incoming: util.NewConcurrentMap[*types.ComparableAddr, net.Conn](),
		outgoing: util.NewConcurrentMap[*types.ComparableAddr, net.Conn](),
		config:   config,
	}

	for _, opt := range opts {
		opt(u)
	}

	return u, nil
}

func (u *UnixTransport) Listen(ctx context.Context) error {
	var err error

	u.listener, err = net.Listen(u.addr.Net, u.addr.Name)
	if err != nil {
		return err
	}

	go u.accept(ctx)

	return nil
}

func (u *UnixTransport) accept(ctx context.Context) {

	defer u.listener.Close()
	type result struct {
		conn net.Conn
		err  error
		id   int
	}

	resultCh := make(chan *result, 1)

	connId := 0
acceptLoop:
	for {

		go func(id int) {
			conn, err := u.listener.Accept()
			resultCh <- &result{
				conn: conn,
				err:  err,
			}
		}(connId)
		select {
		case <-ctx.Done():
			break acceptLoop
		case r := <-resultCh:
			if r.err != nil {
				u.logger.Error("connection error", zap.Error(r.err))
				continue
			}
			go u.handleConn(r.conn, r.id)
		}
	}
}

func (u *UnixTransport) handleConn(conn net.Conn, id int) error {
	u.logger.Debug("new connection",
		zap.String("local", conn.LocalAddr().String()),
		zap.String("remote", conn.RemoteAddr().String()),
	)

	if !u.config.AllowAnonynomous && conn.RemoteAddr().String() == "@" {
		u.logger.Error("refusing anonyomous connection", zap.Bool("allow", u.config.AllowAnonynomous))
		return fmt.Errorf("no peer in incoming unix connection %s", conn.RemoteAddr().String())
	}

	err := u.incoming.Put(
		types.NewComparableAddr(conn.RemoteAddr()),
		conn,
	)
	if err != nil {
		return err
	}

	err = u.config.Handshaker.Handshake(conn)
	if err != nil {
		u.logger.Error("handshake failed. closing connection", zap.Error(err))
		conn.Close()
		return err
	}

	received := make([]byte, 0)
	for {
		buf := make([]byte, 512)

		n, err := conn.Read(buf)
		if err != nil {
			return err
		}
		received = append(received, buf[:n]...)
		if err != nil {
			u.logger.Sugar().Infof("read %d %s", n, string(received))

		}

	}
}
