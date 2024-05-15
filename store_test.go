package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "HelloWorld"
	Pathkey := CASPathTransformFunc(key)
	expectedPath := "db8ac/1c259/eb89d/4a131/b253b/acfca/5f319/d54f2"
	expectedOriginalKey := "db8ac1c259eb89d4a131b253bacfca5f319d54f2"
	if Pathkey.PathName != expectedPath {
		t.Error(t, "have %s want this %s", expectedPath, Pathkey.PathName)
	}
	if Pathkey.Filename != expectedOriginalKey {
		t.Error(t, "have %s want this %s", expectedOriginalKey, Pathkey.Filename)
	}
	fmt.Println(Pathkey)
}

func TestStoreDeleteKey(t *testing.T) {
	opts := StoreOpts{
		PathTansformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)

	key := "myspecialpicture"
	data := []byte("some jpeg bytes")
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTansformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)

	key := "myspecialpicture"
	data := []byte("some jpeg bytes")
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}
	b, _ := ioutil.ReadAll(r)

	if string(b) != string(data) {
		t.Errorf("want %s have %s", string(data), string(b))
	}
	s.Delete(key)

}
