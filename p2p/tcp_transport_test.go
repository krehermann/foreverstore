package p2p

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTCPTransport(t *testing.T) {
	var cnt int
	var mu sync.Mutex
	onPeer := func(p Peer) error {
		mu.Lock()
		defer mu.Unlock()
		cnt++
		return p.Close()
	}
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	u, err := NewTcpTransport(":0", TcpTransportConfig{
		Handshaker:  NOPHandshake{},
		PeerHandler: onPeer,
	}, TcpOptWithLogger(logger))
	assert.NoError(t, err)
	assert.NotEmpty(t, u.addr.String())

	assert.NoError(t, u.Listen(context.Background()))

	nConn := 5
	for i := 0; i < nConn; i++ {

		assert.Nil(t, err)
		c, err := net.Dial(u.listener.Addr().Network(), u.listener.Addr().String())

		t.Logf("conn remote %s, local %s", c.RemoteAddr().String(), c.LocalAddr().String())
		assert.NoError(t, err)
		assert.NotNil(t, c)
	}

	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return cnt == nConn
	},
		100*time.Millisecond, 5*time.Millisecond)

}

func TestTCPTransportRecv(t *testing.T) {
	var cnt int
	var mu sync.Mutex

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	u, err := NewTcpTransport(":0", TcpTransportConfig{
		Handshaker:          NOPHandshake{},
		ProtocolFactoryFunc: NewBinaryProtocolDecoder,
	}, TcpOptWithLogger(logger))
	assert.NoError(t, err)
	assert.NotEmpty(t, u.addr.String())

	assert.NoError(t, u.Listen(context.Background()))

	nConn := 1

	assert.Nil(t, err)
	c, err := net.Dial(u.listener.Addr().Network(), u.listener.Addr().String())

	t.Logf("conn remote %s, local %s", c.RemoteAddr().String(), c.LocalAddr().String())
	assert.NoError(t, err)
	assert.NotNil(t, c)
	r := NewRPC(c.LocalAddr())
	buf := new(bytes.Buffer)
	n, err := buf.WriteString("this is a big message")
	assert.NoError(t, err)

	rpcLenBuf := make([]byte, 4)
	u.logger.Sugar().Debugf("rpc len %d", n)
	binary.LittleEndian.PutUint32(rpcLenBuf, uint32(n))

	r.Write(rpcLenBuf)
	wcnt := 0
	for {
		byt, err := buf.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				u.logger.Sugar().Debugf("read test buf eof %+v", err)
				break
			}
			assert.Failf(t, "error reading byte", "err %+v", err)
		}
		if cnt >= n {
			u.logger.Sugar().Debugf("breaking write loop wrote %d", wcnt)
		}
		n, err := r.Write([]byte{byt})
		assert.Equal(t, 1, n)
		assert.NoError(t, err)
		wcnt += n

	}

	rpcChan := u.Recv()
	u.logger.Sugar().Debug("waiting for rpc")
	got := <-rpcChan

	u.logger.Sugar().Debugf("got rpc %+v %s", got.payload, string(got.payload))
	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return cnt == nConn
	},
		100*time.Millisecond, 5*time.Millisecond)

}
