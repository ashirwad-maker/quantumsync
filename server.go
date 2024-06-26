package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/ashirwad-maker/quantumsync/p2p"
)

type FileServerOpts struct {
	EncKey           []byte // EncKey is used to encrypt and decrypt the data in the file.
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
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		log.Printf("[%s] serving file (%s) from local disk", s.Transport.Addr(), key)
		_, r, err := s.store.Read(key)
		return r, err
	}

	log.Printf("[%s] Don't have the file (%s) locally, fetching from network\n", s.Transport.Addr(), key)

	msg := Message{
		Payload: MessageGetFile{
			Key: hashKey(key),
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}
	time.Sleep(500 * time.Millisecond)

	for _, peer := range s.peers {
		// First read the file size from the peers, then use it in the io.LimitReader
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)

		n, err := s.store.WriteDecrypt(s.EncKey, key, io.LimitReader(peer, fileSize))
		// n, err := s.store.Write(key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		log.Printf("[%s] recieved (%d) over tyhe network from -> (%s) \n", s.Transport.Addr(), n, peer.RemoteAddr())

		peer.CloseStream()
	}
	_, r, err := s.store.Read(key)
	return r, err
}

// broadcasting this to all the peers
func (s *FileServer) stream(msg *Message) error {
	buf := new(bytes.Buffer)
	for _, peer := range s.peers {
		if err := gob.NewEncoder(buf).Encode(msg); err != nil {
			return err
		}
		// con.Write() take slice of bytes in input
		peer.Send(buf.Bytes())
	}
	return nil
}

func (s *FileServer) broadcast(msg *Message) error {
	msgBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		time.Sleep(5 * time.Millisecond)
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func (s *FileServer) Store(key string, r io.Reader) error {
	// 1. Store this file to disk.
	// 2. broadcast this file to all know peers in the network.

	fileBuffer := new(bytes.Buffer)

	// io.TeeReader is used where a copy of io.Reader is required. If not used, then after writing the data to the disk,
	// the io.Reader will be at the EOF and no more data can be read and passed to Payload.data, therefore a copy of io.Reader is required.
	tee := io.TeeReader(r, fileBuffer)
	// a copy is being made in buf.

	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  hashKey(key),
			Size: size + 16,
		},
	}

	// The message is broadcasted to all the peers in the network.
	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 5)
	// Here waiting is necessary as it will send the messages at the same instant, that can cause problem.

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingStream})
		n, err := copyEncrypt(s.EncKey, fileBuffer, peer)
		if err != nil {
			return err
		}
		// n, err := io.Copy(peer, fileBuffer)
		// if err != nil {
		// 	return err
		// }
		log.Printf("recieved and written %d bytes to disk\n", n)
	}

	return nil
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
		log.Println("File Server stopping due to error or user quit action .... ")
		s.Transport.Close()
	}()
	for {
		select {
		case rpc := <-s.Transport.Consume():

			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Printf("decoding error :%s\n", err)
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("handle message error : ", err)
			}

		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}
	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("[%s] need to serve the file (%s) but it does not exist on the disk", s.Transport.Addr(), msg.Key)
	}

	log.Printf("[%s] serving (%s) file over the network\n", s.Transport.Addr(), msg.Key)
	fileSize, r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s does not exist in peer map", from)
	}

	// Asserting ReadCloser, and then deferring.
	if rc, ok := r.(io.ReadCloser); ok {
		defer rc.Close()
	}

	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	log.Printf("[%s] Written (%d) bytes over the network to -> %s\n", s.Transport.Addr(), n, from)

	return nil
}
func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {

	log.Printf("In Process of Storing the file on -> [%s]  ...\n", s.Transport.Addr())
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found", from)
	}

	// Here after the broadcasting the message is read, and stored in the file.
	// The io.Limiter is used with a net.Conn object (peer) asking it to read msg.size bytes.
	_, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	// log.Printf("(%s) written %d bytes to disk\n", s.Transport.Addr, n)

	peer.CloseStream()
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
	gob.Register(MessageGetFile{})
}
