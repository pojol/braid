package actor

import (
	"context"

	"github.com/pojol/braid/router"
)

// 从一个 actor 发给另外一个 actor (理论上绝大部份的场景都应该避免这种形势的调用
// a.system.call(ctx, addr:{target : $id, ev : ""}, msg, res interface{})

// 领取邮件
// a.system.callwait(ctx, addr: {target : $mail, ev : "recv", msg, res interface{}})

// 私聊
// a.system.call(ctx, addr : {target : $chat, ev : "ch_private", msg, interface{}})

// 组队
// a.system.call(ctx, addr : {target : $social, ev : "teamup", msg, res, interface{}})

// 退出公会
// a.system.call(ctx, addr : {target : $guild, ev :"exit", msg, res interface{}})

// 给客户端同步消息
// a.system.call(ctx, addr : {target : $client, ev : "push", msg, nil interface{}})

// 广播消息给客户端
// a.system.call(ctx, addr : {target : $client, ev : "bloadcast", msg, nil, interface{}})

//
// a.system.call(ctx, addr : {target : $login, "ev" : "login"})
// a.system.call(ctx, addr : {target : $game, "ev" : "login"})
// a.system.call(ctx, addr : {target : $entity, "ev" : "ev"})

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

	// Actor 的主循环，它在独立的 goroutine 中运行
	Update()

	Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	Exit()
}

type CreateFunc func(p *CreateActorParm) IActor

type ISystem interface {
	CreateActor() (IActor, error)
	Regist(ty string, opts ...CreateActorOption) // 注册 actor 到节点内
	Actors() []IActor

	FindActor(id string) (IActor, error)

	// 同步调用语义（实际实现是异步的，每个调用都是在独立的goroutine中）
	Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	// 异步调用语义，不阻塞当前的goroutine，用于耗时较长的rpc调用
	Send(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error
}

// chlidren ...
