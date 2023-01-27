package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(any) error
}

type Encoder interface {
	Encoder(io.Writer, any)
}

type GOBDecoder struct {
	//r io.Reader
	gdecoder *gob.Decoder
}

func NewGOBDecoder(r io.Reader) *GOBDecoder {
	return &GOBDecoder{
		gdecoder: gob.NewDecoder(r),
	}
}
func (g *GOBDecoder) Decode(v any) error {
	return g.gdecoder.Decode(v)
}
