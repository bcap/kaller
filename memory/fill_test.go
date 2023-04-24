package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFill(t *testing.T) {
	fill := Fill{}
	fill.DebugPrintStats(true)
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

	fill.Add(kb)
	check(kb)

	fill.Add(mb)
	check(kb + mb)

	fill.Add(50 * mb)
	check(kb + mb + 50*mb)

	fill.Add(100 * mb)
	check(kb + mb + 50*mb + 100*mb)

	fill.Add(150 * mb)
	check(kb + mb + 50*mb + 100*mb + 150*mb)

	fill.Add(-100 * mb)
	check(kb + mb + 50*mb + 100*mb + 150*mb - 100*mb)

	fill.Add(-10 * 1024 * mb)
	check(0)

	fill.Set(-10)
	check(0)

	fill.Set(0)
	check(0)

	fill.Set(10 * mb)
	check(10 * mb)

	fill.Set(5 * mb)
	check(5 * mb)

	fill.Set(15 * mb)
	check(15 * mb)
}
