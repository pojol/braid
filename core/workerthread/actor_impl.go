package workerthread

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/mpsc"
	"github.com/pojol/braid/router"
)

type RecoveryFunc func(interface{})
type MiddlewareHandler func(context.Context, *router.MsgWrapper) error
type EventHandler func(context.Context, *router.MsgWrapper) error

type BaseActor struct {
	Id  string
	Ty  string
	Sys ISystem
	//msgCh    *unbounded.Unbounded
	q        *mpsc.Queue
	closed   int32
	closeCh  chan struct{}
	chains   map[string]IChain
	recovery RecoveryFunc
}

func (a *BaseActor) Type() string {
	return a.Ty
}

func (a *BaseActor) ID() string {
	return a.Id
}

func (a *BaseActor) Init() {
	//a.msgCh = unbounded.NewUnbounded()
	a.q = mpsc.New()
	atomic.StoreInt32(&a.closed, 0) // 初始化closed状态为0（未关闭）
	a.closeCh = make(chan struct{})
	a.chains = make(map[string]IChain)
	a.recovery = defaultRecovery
}

func defaultRecovery(r interface{}) {
	fmt.Printf("Recovered from panic: %v\nStack trace:\n%s\n", r, debug.Stack())
}

func (a *BaseActor) RegisterEvent(ev string, chain IChain) error {
	if _, exists := a.chains[ev]; exists {
		return def.ErrActorRepeatRegisterEvent(ev)
	}
	a.chains[ev] = chain
	return nil
}

func (a *BaseActor) RegisterTimer(dueTime int64, interval int64, f func() error, args interface{}) {

}

func (a *BaseActor) Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error {
	return a.Sys.Call(ctx, tar, msg)
}

func (a *BaseActor) Received(msg *router.MsgWrapper) error {

	msg.Wg.Add(1)

	if atomic.LoadInt32(&a.closed) == 0 { // 并不是所有的actor都需要处理退出信号
		a.q.Push(msg)
		//a.msgCh.Put(msg)
	}

	return nil
}

func (a *BaseActor) Update() {
	checkClose := func() {
		for !a.q.Empty() {
			time.Sleep(10 * time.Millisecond)
		}
		//for a.msgCh.Len() > 0 {
		//	time.Sleep(10 * time.Millisecond)
		//}
		close(a.closeCh)
	}

	for {
		select {
		//case msgInterface := <-a.msgCh.Get():
		case <-a.q.C:
			//a.msgCh.Load() // 处理完之后从队列中丢弃
			msgInterface := a.q.Pop()

			msg, ok := msgInterface.(*router.MsgWrapper)
			if !ok {
				fmt.Println(a.Id, "Received non-Message type")
				continue
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						a.recovery(r)
					}

					// 通知调用者消息处理完成
					msg.Wg.Done()
				}()

				ctx := context.Background()
				if chain, ok := a.chains[msg.Req.Header.Event]; ok {
					err := chain.Execute(ctx, msg)
					if err != nil {
						fmt.Printf("actor %v execute %v chain err %v\n", a.Id, msg.Req.Header.Event, err)
					}
				} else {
					fmt.Printf("actor %v No handlers for message type: %s\n", a.Id, msg.Req.Header.Event)
				}
			}()

			if msg.Req.Header.Event == "exit" {
				atomic.StoreInt32(&a.closed, 1) // 设置closed状态为1（关闭中）
				go checkClose()
				return
			}
		case <-a.closeCh:
			atomic.StoreInt32(&a.closed, 2) // 设置closed状态为2（已关闭）
			fmt.Println(a.Id, "Actor is now closed")
			return
		}
	}
}

func (a *BaseActor) Exit() {

}
