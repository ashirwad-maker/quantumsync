package p2p

// Hanshake func is responsible for accepting the users ...?
type HandshakeFunc func(any) error

func NOPhandshakeFunc(any) error { return nil }
