package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	opts := TCPTransportOptions{
		ListenerAddress: ":8080",
		HandshakeFunc:   NOPHandshakeFunc,
		Decoder:         &DefaultDecoder{},
	}

	transport := NewTCPTransport(opts)
	assert.Equal(t, transport.ListenerAddress, ":8080")

	assert.Nil(t, transport.ListenAndAccept())
}
