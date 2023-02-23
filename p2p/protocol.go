package p2p

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"go.uber.org/zap"
)

type RPC struct {
	From    net.Addr
	payload []byte

	mu  sync.RWMutex
	buf *bytes.Buffer
}

func (r *RPC) Write(b []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.payload == nil {
		r.payload = make([]byte, 0)
	}
	if r.buf == nil {
		r.buf = bytes.NewBuffer(r.payload)
	}

	return r.buf.Write(b)
}

func (r *RPC) Read(b []byte) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.payload == nil {
		r.payload = make([]byte, 0)
	}
	if r.buf == nil {
		r.buf = bytes.NewBuffer(r.payload)
	}

	return r.buf.Read(b)
}

// ProtocolFactoryFunc is type to generate a protocol decoder
// decoder implementations need to have constructor that matches this signature
type ProtocolFactoryFunc func(io.Reader, *zap.Logger) ProtocolDecoder

// ProtocolDecoder is responsible for interepting raw protocol reads into an rpc
type ProtocolDecoder interface {
	Decode(*RPC) error
}

// NewlineProtocolDecoder is useful for manual testing with tools like netcat
type NewlineProtocolDecoder struct {
	r       io.Reader
	logger  *zap.Logger
	bufSize int
}

func NewNewlineDecoder(r io.Reader, l *zap.Logger) ProtocolDecoder {
	return &NewlineProtocolDecoder{
		logger:  l.Named("line-decoder"),
		r:       r,
		bufSize: 1024,
	}
}

// Decode parses the incoming bytes into new-line delimited rpcs
func (d *NewlineProtocolDecoder) Decode(r *RPC) error {

	buf := make([]byte, d.bufSize)
	// todo handle messages large than the buffer size
	// would be a loop until last byte is a new line
	// and a result buffer for appending

	n, err := d.r.Read(buf)
	if n < d.bufSize && n > 0 {
		if buf[n-1] != '\n' {
			return fmt.Errorf("invalid read: no newline delimiter %s", string(buf[n-3:n]))
		}
		_, err = r.Write(buf[:n-1])
		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return nil
}

// BinaryProtocalDecoder is for len prefixed messages
type BinaryProtocolDecoder struct {
	r       io.Reader
	logger  *zap.Logger
	bufSize int
	lenSize int
}

func NewBinaryProtocolDecoder(r io.Reader, l *zap.Logger) ProtocolDecoder {
	return &BinaryProtocolDecoder{
		logger:  l.Named("binary-decoder"),
		r:       r,
		bufSize: 1024,
		lenSize: 4,
	}
}

func (d *BinaryProtocolDecoder) Decode(rpc *RPC) error {
	//buf := make([]byte, d.bufSize)
	lenBuf := make([]byte, d.lenSize)

	lb, err := d.r.Read(lenBuf)
	if err != nil {
		return err
	}
	if lb != d.lenSize {
		return fmt.Errorf("corrupt length prefix")
	}

	length := binary.LittleEndian.Uint32(lenBuf)

	// hack. error handling, ctx
	go func() {
		_, err := io.CopyN(rpc, d.r, int64(length))
		if err != nil {
			panic(err)
		}
	}()

	return nil
}
