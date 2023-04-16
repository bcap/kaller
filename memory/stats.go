package memory

import (
	"fmt"
	"io"
	"runtime"
)

func Stats() runtime.MemStats {
	// https://golang.org/pkg/runtime/#MemStats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

func PrintStats(w io.Writer) {
	stats := Stats()
	fmt.Fprintf(w, ""+
		"HeapAlloc: %s\tTotalHeapAlloc: %s\tHeapInUse: %s\t"+
		"HeapIdle: %s\tHeapReleased: %s\tHeapObjects: %d\t"+
		"Mallocs: %d\tFrees: %d\tGCs: %d\tNextGC: %s\tGCPauseTotalMicros: %d\t"+
		"StackInuse: %s\n",
		HumanizeBytesUint64(stats.HeapAlloc),
		HumanizeBytesUint64(stats.TotalAlloc),
		HumanizeBytesUint64(stats.HeapInuse),
		HumanizeBytesUint64(stats.HeapIdle),
		HumanizeBytesUint64(stats.HeapReleased),
		stats.HeapObjects,
		stats.Mallocs,
		stats.Frees,
		stats.NumGC,
		HumanizeBytesUint64(stats.NextGC),
		stats.PauseTotalNs/1000,
		HumanizeBytesUint64(stats.StackInuse),
	)
}

func HumanizeBytesUint64(bytes uint64) string {
	if bytes < kb {
		return fmt.Sprintf("%db", bytes)
	} else if bytes >= kb && bytes < mb {
		return fmt.Sprintf("%dKb", bytes/kb)
	} else {
		return fmt.Sprintf("%.02fMb", float64(bytes/kb)/float64(kb))
	}
}

func HumanizeBytesInt64(bytes int64) string {
	var signal int64 = 1
	if bytes < 0 {
		signal = -1
	}
	bytes = bytes * signal
	if bytes < int64(kb) {
		return fmt.Sprintf("%db", signal*bytes)
	} else if bytes >= int64(kb) && bytes < int64(mb) {
		return fmt.Sprintf("%dKb", signal*bytes/int64(kb))
	} else {
		return fmt.Sprintf("%.02fMb", float64(signal*bytes/int64(kb))/float64(kb))
	}
}
