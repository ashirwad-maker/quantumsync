package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
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

// This function is required as os.RemoveAll() removes the folder and all its childern,
// so this functions gives the folder where the storage starts.
func (p Pathkey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (p Pathkey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

type StoreOpts struct {
	PathTansformFunc PathTansformFunc
}

var DefaultPathTransformFunc = func(key string) Pathkey {
	return Pathkey{
		PathName: key,
		Filename: key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTansformFunc == nil {
		opts.PathTansformFunc = DefaultPathTransformFunc
	}
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Delete(key string) error {
	pathKey := s.PathTansformFunc(key)
	defer func() {
		log.Printf("delted [%s] from disk", pathKey.Filename)
	}()

	return os.RemoveAll(pathKey.FirstPathName())
}

func (s *Store) Has(key string) bool {
	pathKey := s.PathTansformFunc(key)

	_, err := os.Stat(pathKey.FullPath())
	if err == fs.ErrNotExist {
		return false
	}
	return true
}

func (s *Store) Read(key string) (io.Reader, error) {
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
	f.Close()
	log.Printf("written (%d) bytes to disk : %s", n, fullPath)

	return nil
}
