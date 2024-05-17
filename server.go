package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/ashirwad-maker/quantumsync/p2p"
)

type FileServerOpts struct {
	StorageRoot      string
	PathTansformFunc PathTansformFunc
	Transport        p2p.Transport
	BootstrapNodes   []string // Bootstrap nodes in context of p2p, are specific nodes that serve as initial contact points
	// for new nodes joining the network, they are repsonsible for connection of peers in decentralized network.
}

type FileServer struct {
	FileServerOpts
	peerLock sync.Mutex
	peers    map[string]p2p.Peer
	store    *Store
	quitch   chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:             opts.StorageRoot,
		PathTansformFunc: opts.PathTansformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key string
}

// broadcasting this to all the peers
func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	for _, peer := range s.peers {
		if err := gob.NewEncoder(buf).Encode(msg); err != nil {
			return err
		}
		// con.Write() take slice of bytes in input
		peer.Send(buf.Bytes())
	}

	// Since Peer Interface embeds the Conn interface, which also implements the
	// writer, reader interface so peers can be a slice of io.writer and pass this to the encoder
	// peers := []io.Writer{}
	// for _, peer := range s.peers {
	// 	peers = append(peers, peer)
	// }
	// mw := io.MultiWriter(peers...)

	// It will encode the payload p and will write on mw. Note here a conn.Write() action is happ
	// return gob.NewEncoder(mw).Encode(p)
	return nil
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	msg := Message{
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
	// Here waiting is necessary as it will send the messages at the same instant, that can cause problem.
	// Wait afer sending the first message, so that it can be read at the reciever and then send another message.
	// time.Sleep(time.Second * 3)

	payload := []byte("This large file")
	for _, peer := range s.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}

	return nil
	// 1. Store this file to disk.
	// 2. broadcast this file to all know peers in the network.

	// buf := new(bytes.Buffer)

	// io.TeeReader is used where a copy of io.Reader is required. If not used, then after writing the data to the disk,
	// the io.Reader wi be at the EOF and no more data can be read and passed to Payload.data, therefore a copy of io.Reader is required.

	// tee := io.TeeReader(r, buf)
	// if err := s.store.Write(key, tee); err != nil {
	// 	return err
	// }

	// _, err := io.Copy(buf, r)
	// if err != nil {
	// 	return err
	// }

	// return s.broadcast(&Message{
	// 	Payload: buf,
	// })
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("Connected with remote %s", p.RemoteAddr())

	return nil
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			fmt.Println("Attempting to connect with the remote: ", addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("Dial Error: ", err)

			}
		}(addr)
	}
	return nil
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("File Server stopping due to user quit action .... ")
		s.Transport.Close()
	}()
	for {
		select {
		case rpc := <-s.Transport.Consume():

			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Fatal(err)
				return
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println(err)
				return
			}
			fmt.Printf("rcv message : %+v\n", msg.Payload)

			peer, ok := s.peers[rpc.From]
			if !ok {
				panic("Peer not found")
			}
			b := make([]byte, 1024)
			if _, err := peer.Read(b); err != nil {
				panic(err)
			}

			fmt.Printf("%s\n", string(b))
			peer.(*p2p.TCPPeer).Wg.Done()

		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		s.handleMessageStoreFile(from, v)
	}
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	fmt.Printf("recv store file msg : %+v", msg)
	return nil
}

func (s *FileServer) Start() error {

	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.bootstrapNetwork()

	s.loop()
	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
}
