package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/ashirwad-maker/quantumsync/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPhandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpOpts)

	fileServerOpts := FileServerOpts{
		EncKey:           newEncryptionKey(),
		StorageRoot:      listenAddr[1:] + "quantumsyncnetwork",
		PathTansformFunc: CASPathTransformFunc,
		Transport:        tcpTransport,
		BootstrapNodes:   nodes,
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")
	go func() {
		s1.Start()
	}()
	time.Sleep(2 * time.Second)

	go s2.Start()
	time.Sleep(2 * time.Second)

	for c := 0; c < 20; c++ {
		key := fmt.Sprintf("myCoolPicture_%d", c)
		data := bytes.NewReader([]byte("my big data file is here!"))
		s2.Store(key, data)
		if err := s2.store.Delete(key); err != nil {
			log.Fatal(err)
		}
		r, err := s2.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))

	}

	select {}

}
