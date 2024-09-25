package core

import (
	"context"

	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/router"
)

type CreateActorParm struct {
	ID      string
	ActorTy string
	Options map[string]interface{}
}

func (p *ActorLoaderBuilder) WithID(id string) *ActorLoaderBuilder {
	p.ID = id
	return p
}

func (p *ActorLoaderBuilder) WithType(ty string) *ActorLoaderBuilder {
	p.ActorTy = ty
	return p
}

func (p *ActorLoaderBuilder) WithOpt(key string, value interface{}) *ActorLoaderBuilder {
	p.Options[key] = value
	return p
}

func (p *ActorLoaderBuilder) RegisterLocally() (IActor, error) {
	return p.ISystem.Register(p)
}

func (p *ActorLoaderBuilder) RegisterDynamically(ctx context.Context) error {
	return p.IActorLoader.Pick(p)
}

type CreateFunc func(p *ActorLoaderBuilder) IActor

type ISystem interface {
	Register(*ActorLoaderBuilder) (IActor, error)
	Actors() []IActor

	FindActor(ctx context.Context, id string) (IActor, error)

	// 同步调用语义（实际实现是异步的，每个调用都是在独立的goroutine中）
	Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	// 异步调用语义，不阻塞当前的goroutine，用于耗时较长的rpc调用
	Send(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	// Pub semantics for pubsub, used to publish messages to an actor's message cache queue
	Pub(ctx context.Context, topic string, msg *router.Message) error

	// Sub listens to messages in a channel within a specific topic
	//  opts can be used to set initial values on first listen, such as setting the TTL for messages in this topic
	Sub(topic string, channel string, opts ...pubsub.TopicOption) (*pubsub.Channel, error)

	Loader() IActorLoader

	Update()
	Exit()
}

// chlidren ...
