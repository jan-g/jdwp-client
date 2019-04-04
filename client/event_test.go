package client

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestParseCompositeEvent(t *testing.T) {
	data := []byte{
		1,
		0, 0, 0, 1,
		2,          // EventKind.Breakpoint
		0, 0, 0, 2, // requestId
		// threadId
		0, 0, 0, 0, 0, 0, 0, 3,
		// location
		1,                      // TypeTag = CLASS
		0, 0, 0, 0, 0, 0, 0, 2, // ClassId
		0, 0, 127, 148, 117, 194, 124, 16, // MethodId
		0, 0, 0, 0, 0, 0, 0, 0, // Index
	}
	var comp Composite
	err := Parse(data, &comp)
	assert.Nil(t, err)
}
