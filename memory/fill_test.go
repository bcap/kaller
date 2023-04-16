package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFill(t *testing.T) {
	fill := Fill{}
	kb := 1024
	mb := 1024 * kb

	check := func(size int) {
		assert.Equal(t, size, fill.Size())
		for i := 0; i < len(fill.buf); i++ {
			if fill.buf[i] != byte(i) {
				assert.Equal(t, byte(i), fill.buf[i])
			}
		}
	}

	fill.Grow(kb)
	check(kb)

	fill.Grow(mb)
	check(kb + mb)

	fill.Grow(50 * mb)
	check(kb + mb + 50*mb)

	fill.Grow(100 * mb)
	check(kb + mb + 50*mb + 100*mb)

	fill.Grow(150 * mb)
	check(kb + mb + 50*mb + 100*mb + 150*mb)

	fill.Grow(-100 * mb)
	check(kb + mb + 50*mb + 100*mb + 150*mb - 100*mb)

	fill.Grow(-10 * 1024 * mb)
	check(0)
}
