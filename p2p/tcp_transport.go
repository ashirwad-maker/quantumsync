package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer represents the remote node over a estabilished connection.
// Before a new peer is going to be accepted it needs to "handshake", if the
// "handshake" fails then the connection is dropped otherwise accepted
type TCPPeer struct {
	// The underlying connection of the peer, which in this case is a TCP connection.
	net.Conn
	//if we dial and retrieve a conn => outbound == true
	//if we accept  and retrieve a conn => outbound == false
	outbound bool

	// We are reading from the conn in the handleConn function and we are also reading it in the
	// loop function from the peer, which are basically the same connection, so to avoid race conditions we are using a waitgroup.
	Wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{},
	}
}

func (peer *TCPPeer) Send(b []byte) error {
	_, err := peer.Conn.Write(b)
	return err
}

// The OnPeer() notifies the server what needs to be done with a new peer
// attaching to the server.(cache, drop, etc...)
// Here if the OnPeer() returns error we drop the connection.

type TCPTransportOpts struct {
	// Exported fields.
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts // Using strcture embedding.
	rpcch            chan RPC
	listener         net.Listener
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

// Consume() only reads from the channel for reading the incoming message
// from another peer in the network and implements the Transport Interface.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// Close implements the Transport Interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial implements the Transport Interface
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil
}

// ListenAndAccept implements the Transport Interface
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()
	log.Printf("TCP transport listening on port : %s\n", t.ListenAddr)
	return nil

}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept() // blocking call, till the time a connection does not come
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			fmt.Printf("TCP accept error %s", err)
		}
		// fmt.Printf("New Incoming Connection: %+v \n", conn)
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		fmt.Printf("Dropping Peer connection: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	// First the handshake is called, if the handshake is successful then we will
	// check the t.OnPeer() if that is also fine then we will go in the Read Loop()
	// If either of them fails we will drop the connection.

	if err = t.HandshakeFunc(peer); err != nil {
		fmt.Printf("TCP handshake error: %s\n", err)
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read Loop
	rpc := RPC{}
	for {

		// If the peer is closed then this loop should end/return
		// otherwise if it is a decoder error then it should be keep on going.

		// Note that the message is being decoded in rpc.PayLoad which is a slice of bytes.
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Printf("TCP error:  %s\n", err)
			return
		}
		rpc.From = conn.RemoteAddr().String() // Storing the address of a endpoint in the network
		peer.Wg.Add(1)
		log.Printf("Waiting till stream is done\n")

		t.rpcch <- rpc
		peer.Wg.Wait()
		log.Printf("Stream is done\n")
		// log.Printf("message : %+v\n", rpc)
	}

}
