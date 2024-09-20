package core

import (
	"context"

	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/lib/timewheel"
	"github.com/pojol/braid/router"
)

type IChain interface {
	Execute(context.Context, *router.MsgWrapper) error
}

/*
IActor 对线程（协程）的抽象，在Node(进程)中，是由1～N个actor执行具体的业务逻辑
  - 每一个actor对象代表一个逻辑计算单元，由mailbox去和外部进行交互
*/
type IActor interface {
	Init()

	ID() string
	Type() string

	// 向 actor 的 mailbox 压入一条消息
	Received(msg *router.MsgWrapper) error

	// RegisterEvent registers an event handling chain for the actor
	RegisterEvent(ev string, chain IChain) error

	// RegisterTimer registers a timer function for the actor (Note: all times used here are in milliseconds)
	//  dueTime: delay before execution, 0 for immediate execution
	//  interval: time between each tick
	//  f: callback function
	//  args: can be used to pass the actor entity to the timer callback
	RegisterTimer(dueTime int64, interval int64, f func() error, args interface{}) *timewheel.Timer

	// SubscriptionEvent subscribes to a message
	//  If this is the first subscription to this topic, opts will take effect (you can set some options for the topic, such as ttl)
	//  topic: A subject that contains a group of channels (e.g., if topic = offline messages, channel = actorId, then each actor can get its own offline messages in this topic)
	//  channel: Represents different categories within a topic
	//  succ: Callback function for successful subscription
	SubscriptionEvent(topic string, channel string, succ func(), opts ...pubsub.TopicOption) error

	// Actor 的主循环，它在独立的 goroutine 中运行
	Update()

	// Call 发送一个事件给另外一个 actor
	Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	Exit()
}
