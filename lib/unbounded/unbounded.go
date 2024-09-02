package unbounded

import "sync"

// form https://github.com/grpc/grpc-go/blob/master/internal/buffer/unbounded.go

// Unbounded 一个支持任意长度的channel实现
// 如果有性能要求，将interface类型实例出来，避免一次转换。
type Unbounded struct {
	c       chan interface{}
	backlog []interface{}

	sync.Mutex
}

// NewUnbounded 构建Unbounded
func NewUnbounded() *Unbounded {
	return &Unbounded{c: make(chan interface{}, 1)}
}

// Put 输入一个新的信息
func (b *Unbounded) Put(t interface{}) {
	b.Lock()
	if len(b.backlog) == 0 {
		select {
		case b.c <- t:
			b.Unlock()
			return
		default:
		}
	}
	b.backlog = append(b.backlog, t)
	b.Unlock()
}

// Load 将积压队列中的头部数据提取到channel，并将队列整体前移一位。
func (b *Unbounded) Load() {
	b.Lock()
	if len(b.backlog) > 0 {
		select {
		case b.c <- b.backlog[0]:
			b.backlog[0] = nil
			b.backlog = b.backlog[1:]
		default:
		}
	}
	b.Unlock()
}

// Get 获取unbounded的读channel
func (b *Unbounded) Get() <-chan interface{} {
	return b.c
}

func (b *Unbounded) Len() int {
	return len(b.backlog)
}
