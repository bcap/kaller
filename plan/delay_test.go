package plan

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDelay(t *testing.T) {
	duration := 3 * time.Second
	delay := Delay{CPU: 1.8, Min: duration, Max: duration}
	start := time.Now()
	delay.Do(context.Background())
	timeTaken := time.Since(start)
	assert.Greater(t, timeTaken, duration)
	assert.Less(t, timeTaken, duration+100*time.Millisecond)
}
