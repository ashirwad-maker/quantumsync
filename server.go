package main

import (
	"fmt"
	"log"

	"github.com/ashirwad-maker/quantumsync/p2p"
)

type FileServerOpts struct {
	// ListenAddr       string
	StorageRoot      string
	PathTansformFunc PathTansformFunc
	Transport        p2p.Transport
}

type FileServer struct {
	FileServerOpts
	store  *Store
	quitch chan struct{}
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
	}
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("File Server stopping due to user quit action .... ")
		s.Transport.Close()
	}()
	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) Start() error {

	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.loop()
	return nil
}