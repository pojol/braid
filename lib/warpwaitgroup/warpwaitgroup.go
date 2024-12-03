package warpwaitgroup

import (
	"sync"
	"sync/atomic"
)

// WrapWaitGroup 包装了 sync.WaitGroup，增加了一个计数器。
type WrapWaitGroup struct {
	wg sync.WaitGroup
	// 使用一个 int32 变量来记录当前的计数
	counter int32
}

// Add 添加 delta 到 WaitGroup 计数器并更新内部计数器。
func (w *WrapWaitGroup) Add(delta int) {
	atomic.AddInt32(&w.counter, int32(delta))

	//fmt.Printf("%p cnt %v\n", w, atomic.LoadInt32(&w.counter))
	w.wg.Add(delta)
}

// Done 完成一个 WaitGroup 任务并减少内部计数器。
func (w *WrapWaitGroup) Done() {
	atomic.AddInt32(&w.counter, -1)
	w.wg.Done()
}

// Wait 等待所有的 WaitGroup 任务完成。
func (w *WrapWaitGroup) Wait() {
	w.wg.Wait()
}

// Count 返回当前 WaitGroup 中等待的 goroutine 数量。
func (w *WrapWaitGroup) Count() int32 {
	return atomic.LoadInt32(&w.counter)
}
