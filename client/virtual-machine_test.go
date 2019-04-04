package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseClassesBySignature(t *testing.T) {
	data := []byte{
		0, 0, 0, 1,
		1, 0, 0, 0, 0, 0, 0, 0, 1,
		0, 0, 0, 7,
	}
	var cws ClassesBySignatureReply
	err := Parse(data, &cws)
	assert.Nil(t, err)
}
