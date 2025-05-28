package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathTransformFunc(t *testing.T) {
	key := "foobar"
	pathname := CASPathTransform(key)
	expectedFilename := "8843d7f92416211de9ebb963ff4ce28125932878"
	expectedPathName := "8843d7f924/16211de9eb/b963ff4ce2/8125932878"

	assert.Equal(t, pathname.Filename, expectedFilename)
	assert.Equal(t, pathname.PathName, expectedPathName)

}

func TestStoreWrite(t *testing.T) {
	s := newStore()
	defer teardownStore(t, s)

	for i := range 10 {
		key := fmt.Sprintf("foobar_%d", i)

		data := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")

		// write data
		if err := s.Write(key, bytes.NewReader(data[i:])); err != nil {
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

		if string(b) != string(data[i:]) {
			t.Errorf("Wanted: %s, Got: %s", data, b)

		}

		if err := s.Delete(key); err != nil {
			t.Error(err)
		}

		if ok := s.HasKey(key); ok {
			t.Errorf("Expected to NOT have [%s] key...\n", key)
		}
	}
}

func newStore() *Store {
	storeOpts := StoreOptions{
		Root:              "foobar",
		PathTransformFunc: CASPathTransform,
	}

	return NewStore(storeOpts)
}

func teardownStore(t *testing.T, s *Store) {
	log.Printf("Tearing down store [%s]...", s.Root)
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
