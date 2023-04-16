package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFill(t *testing.T) {
	fill := Fill{}
	kb := 1024
	mb := 1024 * kb

	fill.Grow(kb)
	assert.Equal(t, kb, fill.Size())

	fill.Grow(mb)
	assert.Equal(t, kb+mb, fill.Size())

	fill.Grow(100 * mb)
	assert.Equal(t, kb+mb+100*mb, fill.Size())

	fill.Grow(100 * mb)
	assert.Equal(t, kb+mb+100*mb+100*mb, fill.Size())

	fill.Grow(1024 * mb)
	assert.Equal(t, kb+mb+100*mb+100*mb+1024*mb, fill.Size())

	fill.Grow(-100 * mb)
	assert.Equal(t, kb+mb+100*mb+100*mb+1024*mb-100*mb, fill.Size())

	fill.Grow(-1024 * 2 * mb)
	assert.Equal(t, 0, fill.Size())
}
