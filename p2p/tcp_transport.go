package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer represents the remote node over a estabilished connection.
// Before a new peer is going to be accepted it needs to "handshake", if the
// "handshake" fails then the connection is dropped otherwise accepted
type TCPPeer struct {
	// conn in the underlying connection of the peer
	conn net.Conn

	//if we dial and retrieve a conn => outbound == true
	//if we accept  and retrieve a conn => outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCTTransportOpts struct {
	// Exported fields.
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
}

type TCPTransport struct {
	TCTTransportOpts // Using strcture embedding.
	//listenAddress string ----> moved to TCPTransportOpts
	//	handshakeFunc HandshakeFunc ----> moved to TCPTransportOpts
	// decoder       Decoder ----> moved to TCPTransportOpts
	listener net.Listener
	mu       sync.RWMutex
	peers    map[net.Addr]Peer
}

func NewTCPTransport(opts TCTTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCTTransportOpts: opts,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAccepLoop()
	return nil

}

func (t *TCPTransport) startAccepLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error %s", err)
		}
		fmt.Printf("New Incoming Connection %+v\n", conn)
		go t.handleConn(conn)
	}
}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("TCP handshake error: %s\n", err)
		return
	}

	// Read Loop
	msg := &Message{}
	// buff := make([]byte, 2000)
	for {
		// n, err := conn.Read(buff)
		// if err != nil {
		// 	fmt.Printf("TCP error : %s\n", err)
		// }
		// fmt.Printf("message: %v\n", string(buff[:n]))

		if err := t.Decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP error:  %s\n", err)
			continue
		}
		fmt.Printf("message : %v\n", msg)
	}

}
