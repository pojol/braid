package core

import (
	"context"

	"github.com/pojol/braid/router"
)

type CreateActorParm struct {
	ID  string
	Sys ISystem
}

type CreateActorOption func(*CreateActorParm)

func CreateActorWithID(id string) CreateActorOption {
	return func(p *CreateActorParm) {
		p.ID = id
	}
}

type CreateFunc func(p *CreateActorParm) IActor

type ISystem interface {
	Register(ctx context.Context, ty string, opts ...CreateActorOption) (IActor, error)
	Actors() []IActor

	FindActor(ctx context.Context, id string) (IActor, error)

	// 同步调用语义（实际实现是异步的，每个调用都是在独立的goroutine中）
	Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	// 异步调用语义，不阻塞当前的goroutine，用于耗时较长的rpc调用
	Send(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	// pubsub 的pub语义，用于将消息发布到某个 actor 的消息缓存队列中
	Pub(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	Update()
	Exit()
}

// chlidren ...
