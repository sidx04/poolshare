package p2p

import (
	"net"
	"sync"
)

type TCPTransport struct {
	listenerAddress string
	listener        net.Listener
	mutex           sync.RWMutex
	peers           map[net.Addr]Peer
}

func NewTCPTransport(address string) *TCPTransport {
	return &TCPTransport{
		listenerAddress: address,
	}
}
