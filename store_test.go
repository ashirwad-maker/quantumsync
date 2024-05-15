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
	s := newStore()
	key := "myspecialpicture"
	data := []byte("some jpeg bytes")
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

// TestStore function is responsible for testing the whole store.go functionalities
// 50 times with different keys.
func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTansformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)

	defer teardown(t, s)
	for count := 0; count < 50; count++ {

		key := fmt.Sprintf("myspecialpicture %d", count)
		data := []byte("some jpeg bytes")
		if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); !ok {
			t.Errorf("expected to have the key %s", key)
		}

		r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}
		b, _ := ioutil.ReadAll(r)

		if string(b) != string(data) {
			t.Errorf("want %s have %s", string(data), string(b))
		}
		// fmt.Println(string(b))
		if err := s.Delete(key); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); ok {
			t.Errorf("expected to NOT have the key: %s", key)
		}
	}

}

func newStore() *Store {
	opts := StoreOpts{
		PathTansformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)

}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
