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

// RegisterLocally registers the actor to the current node
func (p *ActorLoaderBuilder) RegisterLocally() (IActor, error) {
	return p.ISystem.Register(p)
}

// RegisterDynamically registers the actor dynamically to the cluster (by selecting an appropriate node through load balancing)
func (p *ActorLoaderBuilder) RegisterDynamically() error {
	return p.IActorLoader.Pick(p)
}

type CreateFunc func(p *ActorLoaderBuilder) IActor

type ISystem interface {
	Register(*ActorLoaderBuilder) (IActor, error)
	Actors() []IActor

	FindActor(ctx context.Context, id string) (IActor, error)

	// Call sends an event to another actor
	// Synchronous call semantics (actual implementation is asynchronous, each call is in a separate goroutine)
	Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	// Send sends an event to another actor
	// Asynchronous call semantics, does not block the current goroutine, used for long-running RPC calls
	Send(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	// Pub semantics for pubsub, used to publish messages to an actor's message cache queue
	Pub(ctx context.Context, topic string, msg *router.Message) error

	// Sub listens to messages in a channel within a specific topic
	//  opts can be used to set initial values on first listen, such as setting the TTL for messages in this topic
	Sub(topic string, channel string, opts ...pubsub.TopicOption) (*pubsub.Channel, error)

	// Loader returns the actor loader
	Loader() IActorLoader

	AddressBook() IAddressBook

	Update()
	Exit()
}

// chlidren ...
