package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "foogen"

func CASPathTransform(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blocksize := 10
	sliceLen := len(hashStr) / blocksize
	paths := make([]string, sliceLen)

	for i := range sliceLen {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

type PathTransform func(string) PathKey

type PathKey struct {
	PathName string
	Filename string
}

func (p *PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (p *PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename[:10]) // use first 10 characters of the hash as the filename
}

type StoreOptions struct {
	// Root is the folder name of the root directory,
	// containing all the files of the system.
	Root                  string
	PathTransformFunction PathTransform
}

var DefaultPathTransformFunction = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

type Store struct {
	StoreOptions
}

func NewStore(options StoreOptions) *Store {
	if options.PathTransformFunction == nil {
		options.PathTransformFunction = DefaultPathTransformFunction
	}

	if len(options.Root) == 0 {
		options.Root = defaultRootFolderName
	}
	return &Store{
		StoreOptions: options,
	}
}

func (s *Store) HasKey(key string) bool {
	pathKey := s.PathTransformFunction(key)
	fullPath := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	_, err := os.Stat(fullPath)

	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Write(key string, r io.Reader) error {
	return s.writeStream(key, r)

}

func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunction(key)
	return os.Open(fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath()))

}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunction(key)

	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return err
	}

	fullPath := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("Written (%d) bytes to disk: %s", n, fullPath)

	return nil
}

func (s *Store) Delete(key string) error {
	path := s.PathTransformFunction(key)

	fullPath := fmt.Sprintf("%s/%s", s.Root, path.FirstPathName())

	defer func() {
		log.Printf("Deleted [%s] from disk", fullPath)
	}()

	return os.RemoveAll(s.Root)
}
