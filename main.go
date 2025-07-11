package main

import (
	"bytes"
	"log"
	"poolshare/p2p"
	"time"
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
		StorageRoot:       listenAddr + "_network",
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

	time.Sleep(2 * time.Second)

	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(2 * time.Second)

	go s2.Start()

	time.Sleep(2 * time.Second)

	data := bytes.NewReader([]byte("abc"))

	s2.StoreData("foo", data)

	select {}
}
