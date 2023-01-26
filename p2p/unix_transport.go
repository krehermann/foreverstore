package p2p

import (
	"context"
	"fmt"
	"net"
	"sync"

	"go.uber.org/zap"
)

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

type UnixTransport struct {
	//addr string
	addr     *net.UnixAddr
	listener net.Listener

	mu sync.RWMutex
	// connections our listener excepted
	incoming map[net.Addr]net.Conn
	// connections this transport has dialed
	outgoing map[net.Addr]net.Conn

	logger *zap.Logger
}

type UnixOpt func(*UnixTransport)

func UnixOptWithLogger(l *zap.Logger) UnixOpt {
	return func(u *UnixTransport) {
		u.logger = l
	}
}

func NewUnixTransport(listenAddr string, opts ...UnixOpt) (*UnixTransport, error) {
	a, err := net.ResolveUnixAddr("unix", listenAddr)
	if err != nil {
		return nil, err
	}

	u := &UnixTransport{
		addr:     a,
		incoming: make(map[net.Addr]net.Conn),
		outgoing: make(map[net.Addr]net.Conn),
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
	u.mu.Lock()
	defer u.mu.Unlock()
	u.logger.Debug("new connection",
		zap.String("local", conn.LocalAddr().String()),
		zap.String("remote", conn.RemoteAddr().String()),
	)

	if conn.RemoteAddr().String() == "@" {
		u.logger.Error("refusing anonyomous connection")
		return fmt.Errorf("no peer in incoming unix connection %s", conn.RemoteAddr().String())
	}
	u.incoming[conn.RemoteAddr()] = conn
	return nil
}
