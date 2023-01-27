package p2p

import (
	"context"
	"net"
)

type connectionHandlerFunc func(net.Conn) error

// accept takes ownership of the listener
// it accepts new connections and calls connection func
func accept(ctx context.Context,
	listener net.Listener,
	handleConnFn connectionHandlerFunc) {

	defer listener.Close()
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
			conn, err := listener.Accept()
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
				// todo error channel?
				//u.logger.Error("connection error", zap.Error(r.err))
				continue
			}
			go handleConnFn(r.conn)
		}
	}

}
