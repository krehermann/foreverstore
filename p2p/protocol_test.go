package p2p

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestBinaryProtocolDecoder_Decode(t *testing.T) {
	type fields struct {
		r       io.Reader
		logger  *zap.Logger
		bufSize int
		lenSize int
	}
	type args struct {
		rpc *RPC
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &BinaryProtocolDecoder{
				r:       tt.fields.r,
				logger:  tt.fields.logger,
				bufSize: tt.fields.bufSize,
				lenSize: tt.fields.lenSize,
			}
			if err := d.Decode(tt.args.rpc); (err != nil) != tt.wantErr {
				t.Errorf("BinaryProtocolDecoder.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	testMsg := []byte("this is a test. this is only a test")
	rdr := new(bytes.Buffer)

	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(testMsg)))

	n, err := rdr.Write(lenBuf)
	assert.Equal(t, 4, n)
	assert.NoError(t, err)

	n, err = rdr.Write(testMsg)
	assert.NoError(t, err)
	assert.Equal(t, len(testMsg), n)

	expectedLen := rdr.Len()
	decoder := NewBinaryProtocolDecoder(rdr, zap.Must(zap.NewDevelopment()))

	rpc := NewRPC(nil)
	err = decoder.Decode(rpc)
	assert.NoError(t, err)

	buf := make([]byte, expectedLen)
	n, err = rpc.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, expectedLen-4, n)

	assert.Equal(t, testMsg, buf[:n])
	t.Fatalf("hack. need to fix decoder and remove sleep. io.Pipe?")
}
