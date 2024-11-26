package actor

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/lib/mpsc"
	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/lib/timewheel"
	"github.com/pojol/braid/router"
	"github.com/pojol/braid/router/msg"
)

type RecoveryFunc func(interface{})
type EventHandler func(*msg.Wrapper) error

type systemKey struct{}
type actorKey struct{}

type Runtime struct {
	Id           string
	Ty           string
	Sys          core.ISystem
	q            *mpsc.Queue
	reenterQueue *mpsc.Queue
	closed       int32
	closeCh      chan struct{}
	shutdownCh   chan struct{}
	chains       map[string]core.IChain
	recovery     RecoveryFunc

	tw       *timewheel.TimeWheel
	lastTick time.Time

	actorCtx *actorContext
}

func (a *Runtime) Type() string {
	return a.Ty
}

func (a *Runtime) ID() string {
	return a.Id
}

func (a *Runtime) Init(ctx context.Context) {
	a.q = mpsc.New()
	a.reenterQueue = mpsc.New()
	atomic.StoreInt32(&a.closed, 0) // 初始化closed状态为0（未关闭）
	a.closeCh = make(chan struct{})
	a.shutdownCh = make(chan struct{})
	a.chains = make(map[string]core.IChain)
	a.recovery = defaultRecovery
	a.actorCtx = &actorContext{
		ctx: ctx,
	}

	a.actorCtx.ctx = context.WithValue(a.actorCtx.ctx, systemKey{}, a.Sys)
	a.actorCtx.ctx = context.WithValue(a.actorCtx.ctx, actorKey{}, a)

	a.tw = timewheel.New(100*time.Millisecond, 100) // 100个槽位，每个槽位10ms
	a.lastTick = time.Now()

	go a.update()
}

func defaultRecovery(r interface{}) {
	fmt.Printf("Recovered from panic: %v\nStack trace:\n%s\n", r, debug.Stack())
}

func (a *Runtime) Context() core.ActorContext {
	return a.actorCtx
}

func (a *Runtime) RegisterEvent(ev string, chainFunc func(ctx core.ActorContext) core.IChain) error {
	if _, exists := a.chains[ev]; exists {
		return fmt.Errorf("actor: repeat register event %v", ev)
	}
	a.chains[ev] = chainFunc(a.actorCtx)
	return nil
}

// RegisterTimer register timer
//
//	dueTime: Delay time before starting the timer (in milliseconds). If 0, starts immediately
//	interval: Time interval between executions (in milliseconds). If 0, executes only once
//	f: Callback function
//	args: Arguments for the callback function
func (a *Runtime) RegisterTimer(dueTime int64, interval int64, f func(interface{}) error, args interface{}) *timewheel.Timer {
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

// SubscriptionEvent subscribes to a message
//
//	If this is the first subscription to this topic, opts will take effect (you can set some options for the topic, such as ttl)
//	topic: A subject that contains a group of channels (e.g., if topic = offline messages, channel = actorId, then each actor can get its own offline messages in this topic)
//	channel: Represents different categories within a topic
//	succ: Callback function for successful subscription
func (a *Runtime) SubscriptionEvent(topic string, channel string, succ func(), opts ...pubsub.TopicOption) error {

	ch, err := a.Sys.Sub(topic, channel, opts...)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	ch.Arrived(a.q)

	if succ != nil {
		succ()
	}

	return nil
}

func (a *Runtime) Call(tar router.Target, mw *msg.Wrapper) error {

	if mw.Req.Header.OrgActorID == "" { // Only record the original sender
		mw.Req.Header.OrgActorID = a.Id
		mw.Req.Header.OrgActorType = a.Ty
	}

	// Updated to the latest value on each call
	mw.Req.Header.PrevActorType = a.Ty

	return a.Sys.Call(tar, mw)
}

func (a *Runtime) Received(mw *msg.Wrapper) error {

	mw.Wg.Add(1)
	if atomic.LoadInt32(&a.closed) == 0 { // 并不是所有的actor都需要处理退出信号
		a.q.Push(mw)
	}

	return nil
}

func (a *Runtime) ReenterCall(ctx context.Context, tar router.Target, rmw *msg.Wrapper) core.IFuture {
	future := NewFuture()

	// 准备消息头
	if rmw.Req.Header.OrgActorID == "" {
		rmw.Req.Header.OrgActorID = a.Id
		rmw.Req.Header.OrgActorType = a.Ty
	}
	rmw.Req.Header.PrevActorType = a.Ty

	// 创建一个带有取消功能的新 context
	ctxWithCancel, cancel := context.WithCancel(ctx)

	// 创建一个 channel 来通知调用完成
	done := make(chan struct{})

	go func() {
		defer cancel()

		// 执行异步调用
		err := a.Sys.Call(tar, rmw)
		if err != nil {
			future.Complete(rmw)
			return
		}

		// 等待消息处理完成或上下文取消
		select {
		case <-ctxWithCancel.Done():
			rmw.Err = ctxWithCancel.Err()
		case <-rmw.Done:
			// 消息处理完成
		}

		future.Complete(rmw)
	}()

	// 创建一个新的 Future 用于重入
	reenterFuture := NewFuture()

	future.Then(func(ret *msg.Wrapper) {
		reenterMsg := &reenterMessage{
			action: func(mw *msg.Wrapper) error {
				defer func() {
					if r := recover(); r != nil {
						log.ErrorF("panic in ReenterCall: %v", r)
						rmw.Err = fmt.Errorf("panic in ReenterCall: %v", r)
						reenterFuture.Complete(rmw)
					}
				}()

				if mw.Err != nil {
					rmw.Err = mw.Err
					reenterFuture.Complete(rmw)
					return mw.Err
				}

				rmw.Res = mw.Res
				reenterFuture.Complete(rmw)
				return nil
			},
			msg: ret,
		}
		a.reenterQueue.Push(reenterMsg)
	})

	// 监听 context 取消
	go func() {
		select {
		case <-ctx.Done():
			cancel()
			if !future.IsCompleted() {
				future.Complete(rmw)
			}
		case <-done:
			// 调用已完成，不需要做任何事
		}
	}()

	return reenterFuture
}

func (a *Runtime) update() {
	ticker := time.NewTicker(a.tw.Interval())
	defer ticker.Stop()

	checkClose := func() {
		for !a.q.Empty() || !a.reenterQueue.Empty() {
			time.Sleep(10 * time.Millisecond)
		}
		if atomic.CompareAndSwapInt32(&a.closed, 1, 2) {
			close(a.closeCh)
		}
	}

	for {
		select {
		case <-ticker.C:
			a.tw.Tick()
			a.lastTick = time.Now()
		case <-a.q.C:
			msgInterface := a.q.Pop()

			mw, ok := msgInterface.(*msg.Wrapper)
			if !ok {
				log.WarnF("actor %v received non-Message type %v", a.Id, reflect.TypeOf(msgInterface))
				continue
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						a.recovery(r)
					}

					// 通知调用者消息处理完成
					mw.Wg.Done()
				}()

				if chain, ok := a.chains[mw.Req.Header.Event]; ok {
					err := chain.Execute(mw)
					if err != nil {
						log.WarnF("actor %v event %v execute err %v", a.Id, mw.Req.Header.Event, err)
					}
				} else {
					log.WarnF("actor %v No handlers for message type: %s", a.Id, mw.Req.Header.Event)
				}
			}()

			if mw.Req.Header.Event == "exit" {
				if atomic.CompareAndSwapInt32(&a.closed, 0, 1) {
					go checkClose()
				}
				return
			}
		case <-a.reenterQueue.C:
			reenterMsgInterface := a.reenterQueue.Pop()
			if reenterMsg, ok := reenterMsgInterface.(*reenterMessage); ok {
				reenterMsg.action(reenterMsg.msg.(*msg.Wrapper))
			}
		case <-a.shutdownCh:
			if atomic.CompareAndSwapInt32(&a.closed, 0, 1) {
				go checkClose()
			}

		case <-a.closeCh:
			atomic.StoreInt32(&a.closed, 2)
			return
		}
	}
}

func (a *Runtime) Exit() {
	close(a.shutdownCh) // 发送关闭信号
	<-a.closeCh         // 等待所有消息处理完毕

	a.tw.Shutdown()
	log.InfoF("[braid.actor] %s has exited", a.Id)
}
