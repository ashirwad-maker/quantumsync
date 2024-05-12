package main

import (
	"fmt"
	"log"

	"github.com/ashirwad-maker/quantumsync/p2p"
)

func main() {
	tcpOpts := p2p.TCTTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPhandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
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
