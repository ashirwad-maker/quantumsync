package p2p

// Message represents any abriitary data that is being sent over the each transport
// between two nodes in the network.
type Message struct {
	Payload []byte
	From    string
}
