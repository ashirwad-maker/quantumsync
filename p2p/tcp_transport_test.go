package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	opts := TCTTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: NOPhandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	tr := NewTCPTransport(opts)
	assert.Equal(t, tr.ListenAddr, ":3000")

	//Server
	assert.Nil(t, tr.ListenAndAccept())
}
