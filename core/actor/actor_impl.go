package actor

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/mpsc"
	"github.com/pojol/braid/lib/timewheel"
	"github.com/pojol/braid/router"
)

type RecoveryFunc func(interface{})
type MiddlewareHandler func(context.Context, *router.MsgWrapper) error
type EventHandler func(context.Context, *router.MsgWrapper) error

type Runtime struct {
	Id       string
	Ty       string
	Sys      core.ISystem
	q        *mpsc.Queue
	closed   int32
	closeCh  chan struct{}
	chains   map[string]core.IChain
	recovery RecoveryFunc

	tw       *timewheel.TimeWheel
	lastTick time.Time
}

func (a *Runtime) Type() string {
	return a.Ty
}

func (a *Runtime) ID() string {
	return a.Id
}

func (a *Runtime) Init() {
	a.q = mpsc.New()
	atomic.StoreInt32(&a.closed, 0) // 初始化closed状态为0（未关闭）
	a.closeCh = make(chan struct{})
	a.chains = make(map[string]core.IChain)
	a.recovery = defaultRecovery

	a.tw = timewheel.New(10*time.Millisecond, 100) // 100个槽位，每个槽位10ms
	a.lastTick = time.Now()
}

func defaultRecovery(r interface{}) {
	fmt.Printf("Recovered from panic: %v\nStack trace:\n%s\n", r, debug.Stack())
}

func (a *Runtime) RegisterEvent(ev string, chain core.IChain) error {
	if _, exists := a.chains[ev]; exists {
		return def.ErrActorRepeatRegisterEvent(ev)
	}
	a.chains[ev] = chain
	return nil
}

// RegisterTimer register timer
//
//	dueTime: Delay time before starting the timer (in milliseconds). If 0, starts immediately
//	interval: Time interval between executions (in milliseconds). If 0, executes only once
//	f: Callback function
//	args: Arguments for the callback function
func (a *Runtime) RegisterTimer(dueTime int64, interval int64, f func() error, args interface{}) *timewheel.Timer {
	return a.tw.AddTimer(
		time.Duration(dueTime)*time.Millisecond,
		time.Duration(interval)*time.Millisecond,
		f,
		args,
	)
}

func (a *Runtime) RemoveTimer(t *timewheel.Timer) {
	a.tw.RemoveTimer(t)
}

func (a *Runtime) Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error {
	return a.Sys.Call(ctx, tar, msg)
}

func (a *Runtime) Received(msg *router.MsgWrapper) error {

	msg.Wg.Add(1)

	if atomic.LoadInt32(&a.closed) == 0 { // 并不是所有的actor都需要处理退出信号
		a.q.Push(msg)
	}

	return nil
}

func (a *Runtime) Update() {
	checkClose := func() {
		for !a.q.Empty() {
			time.Sleep(10 * time.Millisecond)
		}
		close(a.closeCh)
	}

	for {
		now := time.Now()
		if now.Sub(a.lastTick) >= a.tw.Interval() {
			a.tw.Tick()
			a.lastTick = now
		}

		select {
		case <-a.q.C:
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

func (a *Runtime) Exit() {
	a.tw.Shutdown()
}
