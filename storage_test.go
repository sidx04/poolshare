package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathTransformFunction(t *testing.T) {
	key := "foobar"
	pathname := CASPathTransform(key)
	expectedFilename := "8843d7f92416211de9ebb963ff4ce28125932878"
	expectedPathName := "8843d7f924/16211de9eb/b963ff4ce2/8125932878"

	assert.Equal(t, pathname.Filename, expectedFilename)
	assert.Equal(t, pathname.PathName, expectedPathName)

}

func TestStoreWrite(t *testing.T) {
	storeOpts := StoreOptions{
		Root:                  "foobaz",
		PathTransformFunction: CASPathTransform,
	}
	s := NewStore(storeOpts)

	key := "foobar"
	data := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")

	// write data
	if err := s.writeStream("foobar", bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	// read the same data
	r, err := s.readStream(key)

	if err != nil {
		t.Error(err)
	}

	if ok := s.HasKey(key); !ok {
		t.Errorf("Expected to have [%s] key:\n", key)
	}

	b, _ := io.ReadAll(r)

	fmt.Println(string(b))

	if string(b) != string(data) {
		t.Errorf("Wanted: %s, Got: %s", data, b)
	}
}

func TestStoreDelete(t *testing.T) {
	storeOpts := StoreOptions{
		Root:                  "foobar",
		PathTransformFunction: CASPathTransform,
	}
	s := NewStore(storeOpts)

	key := "foobar"
	data := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")

	// write data
	if err := s.writeStream("foobar", bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	// delete the key
	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}
