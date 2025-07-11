package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCP Peer represents remote node over a connection established over TCP.
type TCPPeer struct {
	// `connection` is the underlying connection of the peer;
	// in this case, it is a TCP connection.
	net.Conn
	// outbound -> if we dial and receive a connection;
	// inbound -> if we accept and retreive a connection
	outbound bool

	Wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, out bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: out,
		Wg:       &sync.WaitGroup{},
	}
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

	log.Printf("TCP Transport listening on port: %s\n", t.ListenerAddress)

	return nil

}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

// Close implements the Transport interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial implements the Transport interface
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConnection(conn, true)

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		connection, err := t.listener.Accept()

		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			fmt.Printf("TCP Accept Error: %s\n", err)
		}

		go t.handleConnection(connection, false)
	}
}

func (t *TCPTransport) handleConnection(connection net.Conn, outbound bool) {
	var err error

	defer func() {
		fmt.Printf("Dropping Peer connection %x\n", err)
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
		peer.Wg.Add(1)

		fmt.Println("Waiting till stream is done...")
		t.rpcChannel <- rpc

		peer.Wg.Wait()
		fmt.Println("Stream done, continuing normal read loop")

		// fmt.Printf("RPC: %+v\n", rpc)
	}
}
