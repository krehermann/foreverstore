package p2p

// Peer is interface of a remote node
type Peer interface {
}

// Transport is anything that handles the communication
// between nodes in the network.
// TCP, UDP, websockets
type Transport interface {
	Start() error
}
