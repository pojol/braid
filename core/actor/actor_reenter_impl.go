package actor

import (
	"sync"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/router/msg"
)

// Future represents an asynchronous operation
type Future struct {
	result    *msg.Wrapper
	err       error
	done      chan struct{}
	callbacks []func(mw *msg.Wrapper)
	mutex     sync.Mutex
}

func NewFuture() *Future {
	return &Future{
		done: make(chan struct{}),
	}
}

func (f *Future) Then(callback func(mw *msg.Wrapper)) core.IFuture {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.IsCompleted() {
		go callback(f.result)
		return NewFuture()
	}

	newFuture := NewFuture()
	f.callbacks = append(f.callbacks, func(mw *msg.Wrapper) {
		callback(mw)
		newFuture.Complete(mw)
	})

	return newFuture
}

func (f *Future) Complete(result *msg.Wrapper) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.IsCompleted() {
		return // 已经完成
	}

	f.result = result
	close(f.done)

	for _, callback := range f.callbacks {
		go callback(f.result)
	}
	f.callbacks = nil
}

func (f *Future) IsCompleted() bool {
	select {
	case <-f.done:
		return true
	default:
		return false
	}
}

type reenterMessage struct {
	action EventHandler
	msg    interface{}
}
