package memory

import (
	"fmt"
	"os"
	"runtime"
	"sync"
)

const (
	kb uint64 = 1024
	mb uint64 = kb * 1024
	gb uint64 = mb * 1024
)

var chunk []byte

// If the memory fill is dealing with a large amount of data (> 50mb),
// operations for both growing or shrinking the Fill struct will also involve a
// call to the garbage collector. This is to avoid high memory waste overall
var GCThreshold = 50 * 1024 * 1024 // 50mb

func init() {
	chunk = make([]byte, 1024)
	for i := 0; i < len(chunk); i++ {
		chunk[i] = byte(i)
	}
}

type Fill struct {
	DebugPrintStats bool

	buf   []byte
	mutex sync.RWMutex
}

func (m *Fill) Grow(bytes int) int {
	if m.DebugPrintStats {
		fmt.Fprintf(os.Stderr, "Before Fill.Grow(%s):\t", HumanizeBytesInt64(int64(bytes)))
		PrintStats(os.Stderr)
	}
	newSize := m.grow(bytes)
	oldSize := newSize - bytes
	if oldSize > GCThreshold || newSize > GCThreshold {
		runtime.GC()
	}
	if m.DebugPrintStats {
		fmt.Fprintf(os.Stderr, "After  Fill.Grow(%s):\t", HumanizeBytesInt64(int64(bytes)))
		PrintStats(os.Stderr)
	}
	return newSize
}

func (m *Fill) grow(bytes int) int {
	if bytes == 0 {
		return m.Size()
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	newSize := len(m.buf) + bytes
	if newSize <= 0 {
		m.buf = make([]byte, 0)
	} else {
		new := make([]byte, newSize)
		copy(new, m.buf)
		for i := len(m.buf); i < len(new); i += len(chunk) {
			copy(new[i:], chunk)
		}
		m.buf = new
	}
	return len(m.buf)
}

func (m *Fill) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.buf)
}
