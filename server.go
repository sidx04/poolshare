package main

import (
	"fmt"
	"log"
	"poolshare/p2p"
	"sync"
)

type FileServerOptions struct {
	ListenerAddress   string
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOptions

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store   *Store
	channel chan struct{}
}

func NewFileServer(opts FileServerOptions) *FileServer {
	storageOptions := StoreOptions{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOptions: opts,
		store:             NewStore(storageOptions),
		channel:           make(chan struct{}),
		peers:             make(map[string]p2p.Peer),
	}
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootstrapNetwork()

	s.loop()

	return nil
}

func (s *FileServer) Stop() {
	close(s.channel)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p
	log.Printf("Connected with remote: %s\n", p)

	return nil
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("File server stopped...")
		s.Transport.Close()
	}()

	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)
		case <-s.channel:
			return

		}
	}
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(s.BootstrapNodes) == 0 {
			continue
		}

		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				log.Printf("Dial error: %s\n", err)
			}
		}(addr)
	}
	return nil
}
