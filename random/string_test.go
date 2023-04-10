package random

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	for power := 0.0; power < 25; power++ {
		length := int(math.Pow(2.0, power))
		t.Run(fmt.Sprintf("length %d", length), func(t *testing.T) {
			testRandomString(t, length, RandString(length))
		})
	}
}

func testRandomString(t *testing.T, length int, str string) {
	assert.Len(t, str, length)
	var min float64 = float64(len(str))
	var max float64 = 0.0
	for _, freq := range StringFrequency(str) {
		if freq < min {
			min = freq
		}
		if freq > max {
			max = freq
		}
	}
	fmt.Println(min)
	fmt.Println(max)
}
