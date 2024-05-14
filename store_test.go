package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "HelloWorld"
	path := CASPathTransformFunc(key)
	fmt.Println(path)
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTansformFunc: DefaultPathTransformFunc,
	}
	s := NewStore(opts)

	data := bytes.NewReader([]byte("lesgoooo"))
	s.writeStream("gg", data)
}
