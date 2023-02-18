package p2p

import (
	"context"
	"errors"
	"net"
	"sync"

	"go.uber.org/zap"
)

type connectionHandlerFunc func(net.Conn) error

// accept takes ownership of the listener
// it accepts new connections and calls connection func
// handleConnFn is responsble for closing the connection
func accept(ctx context.Context,
	listener net.Listener,
	lggr *zap.Logger,
	handleConnFn connectionHandlerFunc) {

	type result struct {
		conn net.Conn
		err  error
	}

	resultCh := make(chan *result, 1)
	defer close(resultCh)

	wg := sync.WaitGroup{}
acceptLoop:
	for {

		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := listener.Accept()
			resultCh <- &result{
				conn: conn,
				err:  err,
			}
		}()

		select {
		case <-ctx.Done():
			break acceptLoop
		case r := <-resultCh:
			if r.err != nil {
				lggr.Sugar().Errorf("Accept error %+v", r.err)
				if errors.Is(r.err, net.ErrClosed) {
					lggr.Sugar().Debug("Listener closed. Stopping acceptLoop")
					break
				}
				continue
			}
			go handleConnFn(r.conn)
		}
	}

	wg.Wait()
}
