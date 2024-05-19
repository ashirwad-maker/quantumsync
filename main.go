package main

import (
	"bytes"
	"fmt"
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

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPhandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpOpts)

	fileServerOpts := FileServerOpts{
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

	for c := 0; c < 100; c++ {
		data := bytes.NewReader([]byte("my big data file here!"))
		s2.Store(fmt.Sprintf("myPrivateData_%d", c), data)
	}
	// r, err := s2.Get("myPrivateData")
	// if err != nil {
	// 	log.Println(err)
	// }
	// if err == nil {
	// 	b, err := ioutil.ReadAll(r)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// 	if err == nil {
	// 		fmt.Println(string(b))
	// 	}
	// }
	select {}
	// go func() {
	// 	msg := <-tr.Consume()
	// 	fmt.Printf("%+v\n", msg)
	// }()
	// go func() {
	// 	time.Sleep(time.Second * 10)
	// 	s.Stop()
	// }()

	// if err := s.Start(); err != nil {
	// 	log.Fatal(err)
	// }

}
