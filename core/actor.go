package core

import (
	"context"

	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/lib/timewheel"
	"github.com/pojol/braid/router"
)

type IChain interface {
	Execute(*router.MsgWrapper) error
}

type ActorContext interface {
	// Call 使用 actor 自身发起的 call 调用
	Call(tar router.Target, msg *router.MsgWrapper) error

	// ReenterCall 使用 actor 自身发起的 ReenterCall 调用
	ReenterCall(ctx context.Context, tar router.Target, msg *router.MsgWrapper) IFuture

	// Send sends an event to another actor
	// Asynchronous call semantics, does not block the current goroutine, used for long-running RPC calls
	Send(tar router.Target, msg *router.MsgWrapper) error

	// Pub semantics for pubsub, used to publish messages to an actor's message cache queue
	Pub(topic string, msg *router.Message) error

	// AddressBook 管理全局actor地址的对象，通常由 system 控制调用
	AddressBook() IAddressBook

	// Loader returns the actor loader
	Loader(string) IActorBuilder

	// Unregister unregisters an actor
	Unregister(id string) error

	ID() string
	Type() string

	// WithValue returns a new context with the given state.
	// It allows you to embed any state information into the context for later retrieval.
	//
	// Parameters:
	//   - key: The key for the type to be set in the context (can be defined using the form: type StateKey struct{})
	//   - value: The corresponding value
	WithValue(key, value interface{})

	// GetValue retrieves a value from the context based on the provided key.
	//
	// Parameters:
	//   - key: The key used to store the value in the context
	//
	// Returns:
	//   - The value associated with the key, or nil if not found
	//   - A boolean indicating whether the key was found in the context
	GetValue(key interface{}) interface{}
}

type IFuture interface {
	Complete(*router.MsgWrapper)
	IsCompleted() bool

	Then(func(*router.MsgWrapper)) IFuture
}

// IActor is an abstraction of threads (goroutines). In a Node (process),
// 1 to N actors execute specific business logic.
//
// Each actor object represents a logical computation unit that interacts
// with the outside world through a mailbox.
type IActor interface {
	Init(ctx context.Context)

	ID() string
	Type() string

	// Received pushes a message into the actor's mailbox
	Received(msg *router.MsgWrapper) error

	// RegisterEvent registers an event handling chain for the actor
	RegisterEvent(ev string, createChainF func(ActorContext) IChain) error

	// RegisterTimer registers a timer function for the actor (Note: all times used here are in milliseconds)
	//  dueTime: delay before execution, 0 for immediate execution
	//  interval: time between each tick
	//  f: callback function
	//  args: can be used to pass the actor entity to the timer callback
	RegisterTimer(dueTime int64, interval int64, f func(interface{}) error, args interface{}) *timewheel.Timer

	// SubscriptionEvent subscribes to a message
	//  If this is the first subscription to this topic, opts will take effect (you can set some options for the topic, such as ttl)
	//  topic: A subject that contains a group of channels (e.g., if topic = offline messages, channel = actorId, then each actor can get its own offline messages in this topic)
	//  channel: Represents different categories within a topic
	//  succ: Callback function for successful subscription
	SubscriptionEvent(topic string, channel string, succ func(), opts ...pubsub.TopicOption) error

	// Update is the main loop of the Actor, running in a separate goroutine
	Update()

	// Call sends an event to another actor
	Call(tar router.Target, msg *router.MsgWrapper) error

	ReenterCall(ctx context.Context, tar router.Target, msg *router.MsgWrapper) IFuture

	Context() ActorContext

	Exit()
}

type IActorLoader interface {

	// Builder selects an actor from the factory and provides a builder
	Builder(string, ISystem) IActorBuilder

	// Pick selects an appropriate node for the actor builder to register
	Pick(IActorBuilder) error

	AssignToNode(INode)
}

type IActorBuilder interface {
	GetID() string
	GetType() string
	GetGlobalQuantityLimit() int
	GetNodeUnique() bool
	GetWeight() int
	GetOpt(key string) string
	GetOptions() map[string]string

	GetSystem() ISystem
	GetLoader() IActorLoader
	GetConstructor() CreateFunc

	// ---
	WithID(string) IActorBuilder
	WithType(string) IActorBuilder
	WithOpt(string, string) IActorBuilder

	// ---
	Register() (IActor, error)
	Picker() error
}

type ActorConstructor struct {
	ID   string
	Name string

	// Weight occupied by the actor, weight algorithm reference: 2c4g (pod = 2 * 4 * 1000)
	Weight int

	Dynamic bool

	// Constructor function
	Constructor CreateFunc

	// NodeUnique indicates whether this actor is unique within the current node
	NodeUnique bool

	// Global quantity limit for the current actor type that can be registered
	GlobalQuantityLimit int

	Options map[string]string
}

type IActorFactory interface {
	Get(ty string) *ActorConstructor
	GetActors() []*ActorConstructor
}
