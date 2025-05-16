package main

import (
	"fmt"
	"log"
	"poolshare/p2p"
)

func main() {
	transportOps := p2p.TCPTransportOptions{
		ListenerAddress: ":8080",
		HandshakeFunc:   p2p.NOPHandshakeFunc,
		Decoder:         &p2p.DefaultDecoder{},
	}
	transport := p2p.NewTCPTransport(transportOps)

	if err := transport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}

	fmt.Println("Hello world!")
}
