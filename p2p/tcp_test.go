package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	address := "8080"
	tr := NewTCPTransport(address)

	assert.Equal(t, tr.listenerAddress, address)
}
