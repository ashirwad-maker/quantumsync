package main

import (
	"log"
	"time"

	"github.com/ashirwad-maker/quantumsync/p2p"
)

func OnPeer(peer p2p.Peer) error {
	// This loses the connection
	//return fmt.Errorf("failed the onPeer func")
	// OR
	// peer.Close()
	// fmt.Println("doing some logic with peer outside of TCPTransport")
	return nil
}

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPhandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// OnPeer:        OnPeer,
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:      ":3000_network",
		PathTansformFunc: CASPathTransformFunc,
		Transport:        tr,
	}
	s := NewFileServer(fileServerOpts)

	// go func() {
	// 	msg := <-tr.Consume()
	// 	fmt.Printf("%+v\n", msg)
	// }()
	go func() {
		time.Sleep(time.Second * 10)
		s.Stop()
	}()

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}

}
