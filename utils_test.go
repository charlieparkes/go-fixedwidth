package fixedwidth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPadLeft(t *testing.T) {
	bs := []byte("foobar")
	bs = PadLeft(bs, 10, 'x')
	assert.Equal(t, []byte("xxxxfoobar"), bs)
}

func TestPadRight(t *testing.T) {
	bs := []byte("foobar")
	bs = PadRight(bs, 10, 'x')
	assert.Equal(t, []byte("foobarxxxx"), bs)
}
