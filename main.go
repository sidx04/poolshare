package main

import (
	"log"
	"poolshare/p2p"
)

func onPeer(p p2p.Peer) error {
	p.Close()
	return nil
}

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOptions := p2p.TCPTransportOptions{
		ListenerAddress: listenAddr,
		HandshakeFunc:   p2p.NOPHandshakeFunc,
		Decoder:         &p2p.DefaultDecoder{},
		// onPeer: 			 ?
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOptions)

	fileServerOptions := FileServerOptions{
		StorageRoot:       listenAddr + "network",
		PathTransformFunc: CASPathTransform,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileServerOptions)

	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer(":8080", "")
	s2 := makeServer(":8081", ":8080")

	go func() {
		log.Fatal(s1.Start())
	}()

	s2.Start()
}
