package types

import "net"

type ComparableAddr struct {
	addr    net.Addr
	Network string
	Str     string
}

func NewComparableAddr(addr net.Addr) *ComparableAddr {
	return &ComparableAddr{
		addr:    addr,
		Network: addr.Network(),
		Str:     addr.String(),
	}
}

func (c *ComparableAddr) Addr() net.Addr {
	return c.addr
}
