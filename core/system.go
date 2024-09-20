package core

import (
	"context"

	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/router"
)

type CreateActorParm struct {
	ID      string
	Sys     ISystem
	Options map[string]interface{}
}

type CreateActorOption func(*CreateActorParm)

func CreateActorWithID(id string) CreateActorOption {
	return func(p *CreateActorParm) {
		p.ID = id
	}
}

func CreateActorWithOption(key string, value interface{}) CreateActorOption {
	return func(cap *CreateActorParm) {
		cap.Options[key] = value
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

	// Pub semantics for pubsub, used to publish messages to an actor's message cache queue
	Pub(ctx context.Context, topic string, msg *router.Message) error

	// Sub listens to messages in a channel within a specific topic
	//  opts can be used to set initial values on first listen, such as setting the TTL for messages in this topic
	Sub(topic string, channel string, opts ...pubsub.TopicOption) (*pubsub.Channel, error)

	Update()
	Exit()
}

// chlidren ...
