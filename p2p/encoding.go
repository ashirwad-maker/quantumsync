package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

func (decoder GOBDecoder) Decode(r io.Reader, rpc *RPC) error {
	return gob.NewDecoder(r).Decode(rpc)
}

type DefaultDecoder struct{}

// This is being used in the handleConn
// net.Conn and io.Reader have the same function signatute of Read function therefore they
// can be used interchanebily , this is know as interface substitution.
func (dec DefaultDecoder) Decode(r io.Reader, rpc *RPC) error {
	buf := make([]byte, 2048)
	n, err := r.Read(buf) // blocking call mmove forward after reading from io reader.
	if err != nil {
		return err
	}
	rpc.Payload = buf[:n]
	return nil
}
