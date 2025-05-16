package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCP-Peer represents remote node over a connection established over TCP
type TCPPeer struct {
	// `connection` is the underlying connection of the peer
	connection net.Conn
	// outbound -> if we dial and receive a connection;
	// inbound -> if we accept and retreive a connection
	outbound bool
}

func NewTCPPeer(conn net.Conn, out bool) *TCPPeer {
	return &TCPPeer{
		connection: conn,
		outbound:   out,
	}
}

type TCPTransportOptions struct {
	ListenerAddress string
	HandshakeFunc   HandshakeFunc
	Decoder         Decoder
}

type TCPTransport struct {
	TCPTransportOptions
	listener net.Listener

	mutex sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOptions) *TCPTransport {
	return &TCPTransport{
		TCPTransportOptions: opts,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenerAddress)

	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil

}

func (t *TCPTransport) startAcceptLoop() {
	for {
		connection, err := t.listener.Accept()

		if err != nil {
			fmt.Printf("TCP Accept Error: %s\n", err)
		}

		go t.handleConnection(connection)

		fmt.Printf("New incoming connection: %+v...\n", connection)

	}
}

func (t *TCPTransport) handleConnection(connection net.Conn) {
	peer := NewTCPPeer(connection, true)

	if err := t.HandshakeFunc(peer); err != nil {
		connection.Close()
		fmt.Printf("TCP Handshake Error: %s\n", err)
		return
	}

	// read loop
	msg := &Message{}
	for {
		if err := t.Decoder.Decode(connection, msg); err != nil {
			fmt.Printf("TCP Decode Error: %s\n", err)
		}

		msg.From = connection.RemoteAddr()

		fmt.Printf("Message: %+v\n", msg)
	}

}
