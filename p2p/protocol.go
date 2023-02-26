package p2p

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"go.uber.org/zap"
)

type RPC struct {
	From    net.Addr
	payload []byte

	pr *io.PipeReader
	pw *io.PipeWriter
}

func NewRPC(from net.Addr) *RPC {
	p := make([]byte, 0)
	pr, pw := io.Pipe()
	return &RPC{
		From:    from,
		payload: p,
		pr:      pr,
		pw:      pw,
	}
}

func (r *RPC) Write(b []byte) (int, error) {

	return r.pw.Write(b)
}

func (r *RPC) Read(b []byte) (int, error) {

	return r.pr.Read(b)
}

func (r *RPC) Close() error {
	return r.pw.Close()
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
		logger:  l.Named("BinaryDecoder"),
		r:       r,
		bufSize: 1024,
		lenSize: 4,
	}
}

func (d *BinaryProtocolDecoder) Decode(rpc *RPC) error {

	d.logger.Sugar().Debugf("decoding %+v", rpc)
	lenBuf := make([]byte, d.lenSize)

	lb, err := d.r.Read(lenBuf)
	d.logger.Sugar().Debugf("read  %+v %+v", lb, err)
	if err != nil {
		return err
	}
	if lb != d.lenSize {
		return fmt.Errorf("corrupt length prefix")
	}

	length := binary.LittleEndian.Uint32(lenBuf)
	d.logger.Sugar().Debugf("length prefix %d", length)
	// hack. error handling, ctx
	go func() {
		defer rpc.Close()
		d.logger.Sugar().Debugf("copying...")
		n, err := io.CopyN(rpc, d.r, int64(length))
		d.logger.Sugar().Debugf("copied %d", n)
		if err != nil {
			panic(err)
		}
	}()
	//time.Sleep(2 * time.Second)
	return nil
}
