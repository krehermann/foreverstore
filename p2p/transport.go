package p2p

import (
	"context"
	"net"
)

// Peer is interface of a remote node
type Peer interface {
	Write([]byte) (int, error)
	Close() error
	Addr() net.Addr
}

// Transport is anything that handles the communication
// between nodes in the network.
// TCP, UDP, websockets
type Transport interface {
	Listen(context.Context) error
	Recv() <-chan *RPC
	Close() error
	Addr() net.Addr
	Dial(network, address string) (Peer, error)
}

type PeerHandler func(Peer) error

var _ Peer = remotePeer{}
var _ Peer = localPeer{}

type remotePeer struct {
	net.Conn
}

func (rp remotePeer) Addr() net.Addr {
	return rp.RemoteAddr()
}

type localPeer struct {
	net.Conn
}

func (lp localPeer) Addr() net.Addr {
	return lp.LocalAddr()
}
