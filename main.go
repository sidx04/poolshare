package main

import (
	"fmt"
	"log"
	"poolshare/p2p"
)

func onPeer(p p2p.Peer) error {
	p.Close()
	return nil
}

func main() {
	transportOps := p2p.TCPTransportOptions{
		ListenerAddress: ":8080",
		HandshakeFunc:   p2p.NOPHandshakeFunc,
		Decoder:         &p2p.DefaultDecoder{},
		OnPeer:          onPeer,
	}
	transport := p2p.NewTCPTransport(transportOps)

	go func() {
		for {
			<-transport.Consume()
		}
	}()

	if err := transport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}

	fmt.Println("Hello world!")
}
