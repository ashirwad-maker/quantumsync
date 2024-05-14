package main

import (
	"fmt"
	"log"

	"github.com/ashirwad-maker/quantumsync/p2p"
)

func OnPeer(peer p2p.Peer) error {
	// This loses the connection
	//return fmt.Errorf("failed the onPeer func")
	// OR
	peer.Close()
	// fmt.Println("doing some logic with peer outside of TCPTransport")
	return nil
}

func main() {
	tcpOpts := p2p.TCTTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPhandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	go func() {
		msg := <-tr.Consume()
		fmt.Printf("%+v\n", msg)
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}
