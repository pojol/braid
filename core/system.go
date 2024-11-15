package core

import (
	"context"
	"sync"

	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/router"
)

type CreateFunc func(IActorBuilder) IActor

type ISystem interface {
	Register(IActorBuilder) (IActor, error)
	Unregister(id, ty string) error

	Actors() []IActor

	FindActor(ctx context.Context, id string) (IActor, error)

	// Call sends an event to another actor
	// Synchronous call semantics (actual implementation is asynchronous, each call is in a separate goroutine)
	Call(tar router.Target, msg *router.MsgWrapper) error

	// Send sends an event to another actor
	// Asynchronous call semantics, does not block the current goroutine, used for long-running RPC calls
	Send(tar router.Target, msg *router.MsgWrapper) error

	// Pub semantics for pubsub, used to publish messages to an actor's message cache queue
	Pub(topic string, msg *router.Message) error

	// Sub listens to messages in a channel within a specific topic
	//  opts can be used to set initial values on first listen, such as setting the TTL for messages in this topic
	Sub(topic string, channel string, opts ...pubsub.TopicOption) (*pubsub.Channel, error)

	// Loader returns the actor loader
	Loader(string) IActorBuilder

	AddressBook() IAddressBook

	Update()
	Exit(*sync.WaitGroup)
}

// chlidren ...
