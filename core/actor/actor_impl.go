package actor

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/lib/mpsc"
	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/lib/timewheel"
	"github.com/pojol/braid/router"
)

type RecoveryFunc func(interface{})
type EventHandler func(*router.MsgWrapper) error

type Runtime struct {
	Id         string
	Ty         string
	Sys        core.ISystem
	q          *mpsc.Queue
	closed     int32
	closeCh    chan struct{}
	shutdownCh chan struct{}
	chains     map[string]core.IChain
	recovery   RecoveryFunc

	tw       *timewheel.TimeWheel
	lastTick time.Time

	ctx context.Context
}

func (a *Runtime) Type() string {
	return a.Ty
}

func (a *Runtime) ID() string {
	return a.Id
}

func (a *Runtime) Init(ctx context.Context) {
	a.q = mpsc.New()
	atomic.StoreInt32(&a.closed, 0) // 初始化closed状态为0（未关闭）
	a.closeCh = make(chan struct{})
	a.shutdownCh = make(chan struct{})
	a.chains = make(map[string]core.IChain)
	a.recovery = defaultRecovery
	a.ctx = context.Background()

	a.SetContext(core.SystemKey{}, a.Sys)
	a.SetContext(core.ActorKey{}, a)

	a.tw = timewheel.New(10*time.Millisecond, 100) // 100个槽位，每个槽位10ms
	a.lastTick = time.Now()
}

func defaultRecovery(r interface{}) {
	fmt.Printf("Recovered from panic: %v\nStack trace:\n%s\n", r, debug.Stack())
}

func (a *Runtime) SetContext(key, value interface{}) {
	a.ctx = context.WithValue(a.ctx, key, value)
}

func (a *Runtime) RegisterEvent(ev string, chainFunc func(context.Context) core.IChain) error {
	if _, exists := a.chains[ev]; exists {
		return def.ErrActorRepeatRegisterEvent(ev)
	}
	a.chains[ev] = chainFunc(a.ctx)
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

func (a *Runtime) Call(tar router.Target, msg *router.MsgWrapper) error {

	if msg.Req.Header.OrgActorID == "" { // Only record the original sender
		msg.Req.Header.OrgActorID = a.Id
		msg.Req.Header.OrgActorType = a.Ty
	}

	// Updated to the latest value on each call
	msg.Req.Header.PrevActorType = a.Ty

	return a.Sys.Call(tar, msg)
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
		if atomic.CompareAndSwapInt32(&a.closed, 1, 2) {
			close(a.closeCh)
		}
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

				if chain, ok := a.chains[msg.Req.Header.Event]; ok {
					err := chain.Execute(msg)
					if err != nil {
						fmt.Printf("actor %v execute %v chain err %v\n", a.Id, msg.Req.Header.Event, err)
					}
				} else {
					fmt.Printf("actor %v No handlers for message type: %s\n", a.Id, msg.Req.Header.Event)
				}
			}()

			if msg.Req.Header.Event == "exit" {
				if atomic.CompareAndSwapInt32(&a.closed, 0, 1) {
					go checkClose()
				}
				return
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
	log.Info("[braid.actor] %s has exited", a.Id)
}
