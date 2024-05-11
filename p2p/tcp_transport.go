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

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	handshakeFunc HandshakeFunc

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAddr,
		handshakeFunc: func(any) error { return nil },
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.listenAddress)
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
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	fmt.Printf("New Incoming Connection %+v\n", peer)
}
