package main

import (
	"bytes"
	"fmt"
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

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTansformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)

	data := bytes.NewReader([]byte("some jpeg bytes"))
	s.writeStream("myspecialpicture", data)
}
