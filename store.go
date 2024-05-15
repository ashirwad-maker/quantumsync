package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func CASPathTransformFunc(key string) Pathkey {
	hash := sha1.Sum([]byte(key))

	// sha1.Sum() returns a [20]byte array and to convert it to a
	// slice => hash[:]
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLength := len(hashStr) / blockSize
	fmt.Println(hashStr)

	// Converting a key number of folders and finally merging them to create a path .
	paths := make([]string, sliceLength)
	for i := 0; i < sliceLength; i++ {
		from, to := i*blockSize, (i+1)*blockSize
		paths[i] = hashStr[from:to]
	}

	return Pathkey{
		PathName: strings.Join(paths, "/"),
		Filename: hashStr,
	}

}

type PathTansformFunc func(string) Pathkey

type Pathkey struct {
	PathName string
	Filename string
}

func (p Pathkey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

type StoreOpts struct {
	PathTansformFunc PathTansformFunc
}

var DefaultPathTransformFunc = func(key string) string {
	return key
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buff := new(bytes.Buffer)
	_, err = io.Copy(buff, f)

	return buff, err
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathkey := s.PathTansformFunc(key)
	return os.Open(pathkey.FullPath())
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.PathTansformFunc(key)

	if err := os.MkdirAll(pathKey.PathName, os.ModePerm); err != nil {
		return err
	}
	fullPath := pathKey.FullPath()

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("written (%d) bytes to disk : %s", n, fullPath)

	return nil
}
