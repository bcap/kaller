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

// If the memory fill is dealing with a large amount of data (> 50mb),
// operations for both growing or shrinking the Fill struct will also involve a
// call to the garbage collector. This is to avoid high memory waste overall
var GCThreshold = 50 * 1024 * 1024 // 50mb

var chunk []byte

func init() {
	chunk = make([]byte, 1024)
	for i := 0; i < len(chunk); i++ {
		chunk[i] = byte(i)
	}
}

type Fill struct {
	buf             []byte
	mutex           sync.RWMutex
	debugPrintStats bool
}

func (f *Fill) DebugPrintStats(set bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.debugPrintStats = set
}

type mode int8

const (
	set mode = 0
	add mode = 1
)

func (f *Fill) Set(bytes int) (int, int) {
	return f.adjust(set, bytes, "Set")
}

func (f *Fill) Add(bytes int) (int, int) {
	return f.adjust(add, bytes, "Add")
}

func (f *Fill) adjust(mode mode, bytes int, debugfnName string) (int, int) {
	var fnCall string
	if f.debugPrintStats {
		fnCall = fmt.Sprintf("Fill.%s(%s)", debugfnName, HumanizeBytesInt64(int64(bytes)))
		fmt.Fprintf(os.Stderr, "Before %s:\t", fnCall)
		PrintStats(os.Stderr)
	}
	f.mutex.Lock()
	oldSize := len(f.buf)
	var newSize int
	switch mode {
	case set:
		newSize = bytes
	case add:
		newSize = oldSize + bytes
	default:
		panic(fmt.Sprintf("invalid mode %d", mode))
	}
	adjustBuffer(&f.buf, newSize)
	if oldSize > GCThreshold || newSize > GCThreshold {
		runtime.GC()
	}
	f.mutex.Unlock()
	if f.debugPrintStats {
		fmt.Fprintf(os.Stderr, "After %s:\t", fnCall)
		PrintStats(os.Stderr)
	}
	return oldSize, newSize
}

func adjustBuffer(buffer *[]byte, bytes int) {
	if bytes == len(*buffer) {
		return
	}
	if bytes <= 0 {
		*buffer = []byte{}
		return
	}
	new := make([]byte, bytes)
	copy(new, *buffer)
	for i := len(*buffer); i < len(new); i += len(chunk) {
		copy(new[i:], chunk)
	}
	*buffer = new
}

func (f *Fill) Size() int {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return len(f.buf)
}
