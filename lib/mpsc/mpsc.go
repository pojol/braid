// Package mpsc provides an efficient implementation of a multi-producer, single-consumer lock-free queue.
//
// The Push function is safe to call from multiple goroutines. The Pop and Empty APIs must only be
// called from a single, consumer goroutine.
package mpsc

// This implementation is based on http://www.1024cores.net/home/lock-free-algorithms/queues/non-intrusive-mpsc-node-based-queue

import (
	"sync/atomic"
	"unsafe"
)

type node struct {
	next *node
	val  interface{}
}

type Queue struct {
	head, tail *node
	C          chan int32
	count      int32
}

func New() *Queue {
	q := &Queue{
		C:     make(chan int32, 1),
		count: 0,
	}
	stub := &node{}
	q.head = stub
	q.tail = stub
	return q
}

// Push adds x to the back of the queue.
//
// Push can be safely called from multiple goroutines
func (q *Queue) Push(x interface{}) {
	n := new(node)
	n.val = x
	// current producer acquires head node
	prev := (*node)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.head)), unsafe.Pointer(n)))

	// release node to consumer
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&prev.next)), unsafe.Pointer(n))

	q.setCount(1)
}

// Pop removes the item from the front of the queue or nil if the queue is empty
//
// Pop must be called from a single, consumer goroutine
func (q *Queue) Pop() interface{} {
	tail := q.tail
	next := (*node)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&tail.next)))) // acquire
	if next != nil {
		q.tail = next
		v := next.val
		next.val = nil
		q.setCount(-1)
		return v
	}
	return nil
}

// Empty returns true if the queue is empty
//
// Empty must be called from a single, consumer goroutine
func (q *Queue) Empty() bool {
	tail := q.tail
	next := (*node)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&tail.next))))
	return next == nil
}

func (q *Queue) Count() int32 {
	return atomic.LoadInt32(&q.count)
}

func (q *Queue) setCount(delta int32) {
	count := atomic.AddInt32(&q.count, delta)
	if count > 0 {
		select {
		case q.C <- count:
		default:
		}
	}
}
