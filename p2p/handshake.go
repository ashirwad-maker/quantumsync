package p2p

// Hanshake func is responsible for accepting the users ...?

type HandshakeFunc func(Peer) error

func NOPhandshakeFunc(Peer) error { return nil }
