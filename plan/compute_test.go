package plan

import (
	"context"
	"testing"
	"time"

	"github.com/bcap/kaller/memory"
	"github.com/stretchr/testify/assert"
)

func TestCompute(t *testing.T) {
	duration := 3 * time.Second
	compute := Compute{CPU: 1.8, Min: duration, Max: duration, MemoryDeltaKB: 1024}
	start := time.Now()
	fill := memory.Fill{}
	compute.Do(context.Background(), &fill)
	timeTaken := time.Since(start)
	assert.Greater(t, timeTaken, duration)
	assert.Less(t, timeTaken, duration+100*time.Millisecond)
	assert.Equal(t, 1024*1024, fill.Size())
}
