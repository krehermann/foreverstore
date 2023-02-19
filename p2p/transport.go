package p2p

import (
	"context"
	"net"
)

// Peer is interface of a remote node
type Peer interface {
	Close() error
	Addr() net.Addr
	//RemoteAddr() net.Addr
}

// Transport is anything that handles the communication
// between nodes in the network.
// TCP, UDP, websockets
type Transport interface {
	//Start() error
	Listen(context.Context) error
	Recv() <-chan RPC
	//	Close() error
	Peer
	Dial(network, address string) (Peer, error)
	//Dial(net.Addr) (Transport, error)
}

type PeerHandler func(Peer) error

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
