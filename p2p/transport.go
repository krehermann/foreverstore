package p2p

import (
	"context"
)

// Peer is interface of a remote node
type Peer interface {
	Close() error
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
