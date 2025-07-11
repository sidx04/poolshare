package p2p

import "net"

// Peer is any node that contains data and represents a node.
type Peer interface {
	net.Conn
	Send([]byte) error
	// CloseStream() error
}

// Transport is medium that packets travel over and handles communication
// between nodes in the network. This can be TCP, UDP, Websockets, etc.
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
