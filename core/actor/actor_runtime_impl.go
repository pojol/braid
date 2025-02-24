package actor

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/node"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/lib/mpsc"
	"github.com/pojol/braid/lib/pubsub"
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

	timers    map[core.ITimer]struct{}
	timerChan chan core.ITimer
	//timerWg   sync.WaitGroup // 用于等待所有 timer goroutine 退出

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

	a.timers = make(map[core.ITimer]struct{})
	a.timerChan = make(chan core.ITimer, 1024)

	go a.update()
}

func defaultRecovery(r interface{}) {
	fmt.Printf("Recovered from panic: %v\nStack trace:\n%s\n", r, debug.Stack())
}

func (a *Runtime) Context() core.ActorContext {
	return a.actorCtx
}

func (a *Runtime) OnEvent(ev string, chainFunc func(ctx core.ActorContext) core.IChain) error {
	if _, exists := a.chains[ev]; exists {
		return fmt.Errorf("actor: repeat register event %v", ev)
	}
	a.chains[ev] = chainFunc(a.actorCtx)
	return nil
}

// OnTimer register timer
//
//	dueTime: Delay time before starting the timer (in milliseconds). If 0, starts immediately
//	interval: Time interval between executions (in milliseconds). If 0, executes only once
//	f: Callback function
//	args: Arguments for the callback function
func (a *Runtime) OnTimer(dueTime int64, interval int64, f func(interface{}) error, args interface{}) core.ITimer {
	info := NewTimerInfo(
		time.Duration(dueTime)*time.Millisecond,
		time.Duration(interval)*time.Millisecond,
		f, args)

	a.timers[info] = struct{}{}
	//a.timerWg.Add(1)

	go func() {
		//defer a.timerWg.Done()

		// 如果 dueTime 大于 0，使用 dueTime 进行第一次触发
		if info.dueTime > 0 {
			<-time.After(info.dueTime)
			a.timerChan <- info
		}

		info.ticker = time.NewTicker(info.interval)

		for {
			select {
			case <-info.ticker.C:
				a.timerChan <- info
			case <-a.shutdownCh:
				log.InfoF("[braid.timer] shutdown ch")
				return
			}
		}
	}()

	return info
}

func (a *Runtime) CancelTimer(t core.ITimer) {
	if t == nil {
		return
	}

	log.InfoF("[braid.timer] %v timer cancel", a.Id)

	t.Stop()
	delete(a.timers, t)
}

// Sub subscribes to a message
//
//	If this is the first subscription to this topic, opts will take effect (you can set some options for the topic, such as ttl)
//	topic: A subject that contains a group of channels (e.g., if topic = offline messages, channel = actorId, then each actor can get its own offline messages in this topic)
//	channel: Represents different categories within a topic
//	callback: Callback function for successful subscription
func (a *Runtime) Sub(topic string, channel string, callback func(ctx core.ActorContext) core.IChain, opts ...pubsub.TopicOption) error {

	ch, err := a.Sys.Sub(topic, channel, opts...)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	ch.Arrived(a.q)

	a.OnEvent(channel, callback)

	return nil
}

func (a *Runtime) Call(idOrSymbol, actorType, event string, mw *msg.Wrapper) error {

	if mw.Req.Header.OrgActorID == "" { // Only record the original sender
		mw.Req.Header.OrgActorID = a.Id
		mw.Req.Header.OrgActorType = a.Ty
	}

	// Updated to the latest value on each call
	mw.Req.Header.PrevActorType = a.Ty

	return a.Sys.Call(idOrSymbol, actorType, event, mw)
}

func (a *Runtime) Received(mw *msg.Wrapper) error {

	if mw.Req.Header.OrgActorID != "" {
		if mw.Req.Header.OrgActorID == a.Id {
			return node.ErrSelfCall
		}
	}

	mw.GetWg().Add(1)
	if atomic.LoadInt32(&a.closed) == 0 { // 并不是所有的actor都需要处理退出信号
		a.q.Push(mw)
	}

	return nil
}

func (a *Runtime) ReenterCall(idOrSymbol, actorType, event string, rmw *msg.Wrapper) core.IFuture {
	if rmw.Req.Header.OrgActorID == "" {
		rmw.Req.Header.OrgActorID = a.Id
		rmw.Req.Header.OrgActorType = a.Ty
	}
	rmw.Req.Header.PrevActorType = a.Ty

	reenterFuture := NewFuture()
	callFuture := NewFuture()

	deadline, ok := rmw.Ctx.Deadline()
	var timeout time.Duration
	if ok {
		timeout = time.Until(deadline)
	} else {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	go func() {
		select {
		case <-rmw.Ctx.Done():
			log.WarnF("[ReenterCall] Context canceled: %v", rmw.Ctx.Err())
			cancel()

			errWrapper := &msg.Wrapper{
				Ctx: rmw.Ctx,
				Err: rmw.Ctx.Err(),
			}

			reenterFuture.Complete(errWrapper)
		case <-callFuture.done:
		}
	}()

	go func() {
		defer cancel()
		log.InfoF("[ReenterCall] Starting call to %s.%s", actorType, event)

		swappedWrapper := msg.Swap(rmw)
		//swappedWrapper.Ctx = ctx

		err := a.Sys.Call(idOrSymbol, actorType, event, swappedWrapper)
		if err != nil {
			callFuture.Complete(&msg.Wrapper{
				Ctx: ctx,
				Err: err,
			})
			return
		}

		callFuture.Complete(swappedWrapper)
	}()

	// 设置回调，将处理放入重入队列
	callFuture.Then(func(ret *msg.Wrapper) {

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

	return reenterFuture
}

func (a *Runtime) update() {
	checkClose := func() {
		timeout := time.After(10 * time.Second)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for !a.q.Empty() || !a.reenterQueue.Empty() {
			select {
			case <-timeout:
				log.WarnF("[braid.actor] %s force close due to timeout waiting for queue to empty remaining %v", a.Id, a.q.Count())
				goto ForceClose
			case <-ticker.C:
				continue
			}
		}

	ForceClose:
		if atomic.CompareAndSwapInt32(&a.closed, 1, 2) {
			log.InfoF("[braid.actor] %s closing channel", a.Id)
			close(a.closeCh)
		}
	}

	for {
		select {
		case timerInfo := <-a.timerChan:
			if atomic.LoadInt32(&a.closed) != 0 {
				continue
			}
			if err := timerInfo.Execute(); err != nil {
				log.WarnF("actor %v timer callback error: %v", a.Id, err)
			}
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

					mw.GetWg().Done()
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

		case <-a.reenterQueue.C:
			reenterMsgInterface := a.reenterQueue.Pop()
			if reenterMsg, ok := reenterMsgInterface.(*reenterMessage); ok {
				reenterMsg.action(reenterMsg.msg.(*msg.Wrapper))
			}

		case <-a.shutdownCh:
			if atomic.CompareAndSwapInt32(&a.closed, 0, 1) {
				log.DebugF("[braid.actor] %s exiting check close %v", a.Id, atomic.LoadInt32(&a.closed))
				go checkClose()
			}
		case <-a.closeCh:
			log.DebugF("[braid.actor] %s exiting closed", a.Id)
			return
		}
	}
}

func (a *Runtime) Exit() {
	log.DebugF("[braid.actor] %s exiting state %v remaining msg %v", a.Id, atomic.LoadInt32(&a.closed), a.q.Count())
	close(a.shutdownCh) // 发送关闭信号
	<-a.closeCh         // 等待所有消息处理完毕

	for t := range a.timers {
		a.CancelTimer(t)
	}
	//a.timerWg.Wait()

	log.InfoF("[braid.actor] %s has exited", a.Id)
}
