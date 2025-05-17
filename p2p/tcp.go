package p2p

import (
	"fmt"
	"net"
)

// TCP Peer represents remote node over a connection established over TCP.
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

// Close implements the Peer interface and closes the connection of the peer.
func (p *TCPPeer) Close() error {
	return p.connection.Close()
}

type TCPTransportOptions struct {
	ListenerAddress string
	HandshakeFunc   HandshakeFunc
	Decoder         Decoder
	OnPeer          func(Peer) error
}

type TCPTransport struct {
	TCPTransportOptions
	listener   net.Listener
	rpcChannel chan RPC
}

func NewTCPTransport(opts TCPTransportOptions) *TCPTransport {
	return &TCPTransport{
		TCPTransportOptions: opts,
		rpcChannel:          make(chan RPC),
	}
}

// Consume implements the transport interface, which will return
// a read-onlt channel for reading incoming messages received
// from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChannel
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
	var err error

	defer func() {
		fmt.Printf("Dropping Peer connection %s\n", err)
		connection.Close()
	}()

	peer := NewTCPPeer(connection, true)

	// first, perform tcp handshake for the peer
	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	// next, do onPeer
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// if both hadnshake and onPeer suceed, commence read loop
	rpc := RPC{}
	for {
		err := t.Decoder.Decode(connection, &rpc)

		if err != nil {
			return
		}

		rpc.From = connection.RemoteAddr()
		t.rpcChannel <- rpc

		fmt.Printf("RPC: %+v\n", rpc)
	}

}
