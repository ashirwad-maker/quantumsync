package p2p

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// RPC (Remote Procedure Call) represents any abriitary data that is being sent
// over the each transport between two nodes in the network.
// RPC in context of p2p(peer to peer) can invoke procedures or functions on remote
// peers allowing a computer program to execute on a remote node.
type RPC struct {
	Payload []byte
	From    string
	Stream  bool
}
