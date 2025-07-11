package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
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

type Message struct {
	From    string
	Payload any
}

type MessageStoreFile struct {
	Key string
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

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// 1. store file to disk
	// 2. broadcast file all known peers in the network

	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)

	// if err := s.store.Write(key, tee); err != nil {
	// 	return err
	// }

	// pl := &DataEnvelope{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }

	// return s.broadcast(&Message{
	// 	From:    s.ListenerAddress,
	// 	Payload: pl,
	// })

	buf := new(bytes.Buffer)

	msg := Message{
		// From: ,
		Payload: MessageStoreFile{
			Key: key,
		},
	}

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	// time.Sleep(1 * time.Second)

	// payload := []byte("file")
	// for _, peer := range s.peers {
	// 	if err := peer.Send(payload); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p
	log.Printf("Connected with remote: %+v\n", p)

	return nil
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("File server stopped...")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message

			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%+v\n", msg.Payload)

			// fmt.Printf("Recv: %s\n", string(msg.Payload.([]byte)))

			peer, ok := s.peers[rpc.From.String()]
			if !ok {
				panic("Peer not found in peer map.")
			}

			buf := make([]byte, 1024)

			if _, err := peer.Read(buf); err != nil {
				panic(err)
			}

			fmt.Printf("Recv: %s\n", string(buf))

			peer.(*p2p.TCPPeer).Wg.Done()

			// if err := s.handleMessage(&m); err != nil {
			// 	log.Fatal(err)
			// }

		case <-s.channel:
			return
		}
	}
}

// func (s *FileServer) handleMessage(msg *Message) error {
// 	switch v := msg.Payload.(type) {

// 	case *DataEnvelope:
// 		fmt.Printf("%+v", v)
// 	}

// 	return nil
// }

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
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

func init() {
	gob.Register(MessageStoreFile{})
}
