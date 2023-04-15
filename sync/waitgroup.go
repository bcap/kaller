package sync

import (
	"sync"
	"sync/atomic"
)

type WaitGroup struct {
	sync.WaitGroup

	current int32
}

func (wg *WaitGroup) Add(delta int) {
	atomic.AddInt32(&wg.current, int32(delta))
	wg.WaitGroup.Add(delta)
}

func (wg *WaitGroup) Done() {
	wg.WaitGroup.Done()
	atomic.AddInt32(&wg.current, -1)
}

func (wg *WaitGroup) Current() int {
	return int(atomic.LoadInt32(&wg.current))
}
