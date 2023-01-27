package p2p

import (
	"errors"
	"net"
)

var ErrHandshakeFailed = errors.New("handshake failed")

type Handshaker interface {
	Handshake(net.Conn) error
}

type NOPHandshake struct{}

func (n NOPHandshake) Handshake(net.Conn) error { return nil }
