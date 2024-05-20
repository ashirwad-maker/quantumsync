package main

import (
	"bytes"
	"testing"
)

func TestCopyEncrypt(t *testing.T) {
	payLoad := "Hello World"
	src := bytes.NewReader([]byte(payLoad))
	dst := new(bytes.Buffer)
	key := newEncryptionKey()

	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	out := new(bytes.Buffer)
	nw, err := copyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}
	if nw != len(payLoad)+16 {
		t.Fail()
	}
	if out.String() != payLoad {
		t.Errorf("decryption Failed")
	}
}
