package p2p

// Peer is any node that contains data and represents a node.
type Peer interface {
	Close() error
}

// Transport is medium that packets travel over and handles communication
// between nodes in the network. This can be TCP, UDP, Websockets, etc.
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
}
