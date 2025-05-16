package p2p

import "net"

// Message contains any data that is being sent over each transport
// between two nodes in the network
type Message struct {
	From    net.Addr
	Payload []byte
}
