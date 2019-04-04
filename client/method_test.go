package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLineTable(t *testing.T) {
	data := []byte{
		0, 0, 0, 39, 76, 111, 114, 103,
		47, 105, 111, 99, 116, 108, 47, 100,
		101, 98, 117, 103,
		47, 97, 112, 112, 47,
		87, 101, 98, 83, 101, 114, 118, 101, 114, 36, 72, 97, 110, 100, 108, 101, 114, 59,
	}
	var ltr LineTableReply
	err := Parse(data, &ltr)
	assert.Nil(t, err)
}
