package p2p

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestUnixTransport(t *testing.T) {
	d := t.TempDir()

	f := filepath.Join(d, "abd")
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	u, err := NewTcpTransport(f, TcpOptWithLogger(logger))
	assert.NoError(t, err)
	assert.NotEmpty(t, u.addr.String())

	assert.NoError(t, u.Listen(context.Background()))

	nConn := 5
	for i := 0; i < nConn; i++ {
		localAddr, err := net.ResolveUnixAddr("unix", filepath.Join(d,
			fmt.Sprintf("dialer-%d", i)))
		assert.Nil(t, err)
		//c, err := net.Dial(u.listener.Addr().Network(), u.listener.Addr().String())
		c, err := net.DialUnix("unix", localAddr, u.addr)

		t.Logf("conn remote %s, local %s", c.RemoteAddr().String(), c.LocalAddr().String())
		assert.NoError(t, err)
		assert.NotNil(t, c)
	}

	assert.Eventually(t, func() bool {
		return len(u.incoming) == nConn
	},
		100*time.Millisecond, 5*time.Millisecond)

}
