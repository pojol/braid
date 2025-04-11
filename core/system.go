package core

import (
	"context"
	"errors"
	"sync"

	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/router/msg"
)

type CreateFunc func(IActorBuilder) IActor

var ErrActorRegisterRepeat = errors.New("[braid.system] register actor repeat")

type ISystem interface {
	Register(context.Context, IActorBuilder) (IActor, error)
	Unregister(id, ty string) error

	Actors() []IActor

	FindActor(ctx context.Context, id string) (IActor, error)

	// Call sends an event to another actor
	// Synchronous call semantics (actual implementation is asynchronous, each call is in a separate goroutine)
	Call(idOrSymbol, actorType, event string, mw *msg.Wrapper) error

	// Send sends an event to another actor
	// Asynchronous call semantics, does not block the current goroutine, used for long-running RPC calls
	Send(idOrSymbol, actorType, event string, mw *msg.Wrapper) error

	// Pub semantics for pubsub, used to publish messages to an actor's message cache queue
	Pub(topic string, event string, body []byte) error

	// Sub listens to messages in a channel within a specific topic
	//  opts can be used to set initial values on first listen, such as setting the TTL for messages in this topic
	Sub(topic string, channel string, opts ...pubsub.TopicOption) (*pubsub.Channel, error)

	// Loader returns the actor loader
	Loader(string) IActorBuilder

	AddressBook() IAddressBook

	Exit(*sync.WaitGroup)
}

// chlidren ...
