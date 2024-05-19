package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultFolderName = "quantumsyncnetwork"

func CASPathTransformFunc(key string) Pathkey {
	hash := sha1.Sum([]byte(key))

	// sha1.Sum() returns a [20]byte array and to convert it to a
	// slice => hash[:]
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLength := len(hashStr) / blockSize

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
	// Root is the folder name of the root, contaning all the folders/files of the system.
	Root             string
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
	if len(opts.Root) == 0 {
		opts.Root = defaultFolderName
	}
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(key string) error {
	pathKey := s.PathTansformFunc(key)
	defer func() {
		log.Printf("delted [%s] from disk", pathKey.Filename)
	}()
	firstPathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FirstPathName())

	return os.RemoveAll(firstPathNameWithRoot)
}

func (s *Store) Has(key string) bool {
	pathKey := s.PathTansformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

func (s *Store) Read(key string) (int64, io.Reader, error) {
	return s.readStream(key)
}

func (s *Store) readStream(key string) (int64, io.ReadCloser, error) {
	pathkey := s.PathTansformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.FullPath())

	fi, err := os.Stat(fullPathWithRoot)
	if err != nil {
		return 0, nil, err
	}
	file, err := os.Open(fullPathWithRoot)
	if err != nil {
		return 0, nil, err
	}
	return fi.Size(), file, nil

}

func (s *Store) writeStream(key string, r io.Reader) (int64, error) {
	pathKey := s.PathTansformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return 0, err
	}
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	f, err := os.Create(fullPathWithRoot)
	if err != nil {
		return 0, err
	}
	// io.Copy() keeps on copying from the souce till the EOF is not found, leading to a blocking and disallowing streaming.
	// Therefor a io.LimitReader is passed while calling.
	n, err := io.Copy(f, r)
	if err != nil {
		return 0, err
	}
	f.Close()
	log.Printf("written (%d) bytes to disk : %s", n, fullPathWithRoot)

	return n, nil
}
