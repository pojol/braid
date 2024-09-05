package workerthread

import (
	"context"

	"github.com/pojol/braid/router"
)

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

	// 为 actor 注册一个事件
	RegisterEvent(ev string, chain IChain) error

	// 为 actor 注册一个定时函数（注：这边用到的时间都是毫秒
	//  dueTime 延迟执行 0 为立即执行
	//  interval 每次tick的间隔时间
	//  f 回调函数
	//  args 可以将 actor 实体传递给 timer 回调
	RegisterTimer(dueTime int64, interval int64, f func() error, args interface{})

	// Actor 的主循环，它在独立的 goroutine 中运行
	Update()

	// Call 发送一个事件给另外一个 actor
	Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	Exit()
}

type CreateFunc func(p *CreateActorParm) IActor

type ISystem interface {
	CreateActor() (IActor, error)
	Register(ctx context.Context, ty string, opts ...CreateActorOption)
	Actors() []IActor

	FindActor(id string) (IActor, error)

	// 同步调用语义（实际实现是异步的，每个调用都是在独立的goroutine中）
	Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	// 异步调用语义，不阻塞当前的goroutine，用于耗时较长的rpc调用
	Send(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	// pubsub 的pub语义，用于将消息发布到某个 actor 的消息缓存队列中
	Pub(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error
}

// chlidren ...
