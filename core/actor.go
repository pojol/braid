package core

import (
	"context"
	"time"

	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/router/msg"
)

type IChain interface {
	Execute(*msg.Wrapper) error
}

type ActorContext interface {
	// Call performs a blocking call to target actor
	//
	// Parameters:
	//   - idOrSymbol: target actorID, or routing rule symbol to target actor
	//   - actorType: type of actor, obtained from actor template
	//   - event: event name to be handled
	//   - mw: message wrapper for routing
	Call(idOrSymbol, actorType, event string, mw *msg.Wrapper) error

	// ReenterCall performs a reentrant(asynchronous) call
	//
	// Parameters:
	//   - idOrSymbol: target actorID, or routing rule symbol to target actor
	//   - actorType: type of actor, obtained from actor template
	//   - event: event name to be handled
	//   - mw: message wrapper for routing
	ReenterCall(idOrSymbol, actorType, event string, mw *msg.Wrapper) IFuture

	// Send performs an asynchronous call
	//
	// Parameters:
	//   - idOrSymbol: target actorID, or routing rule symbol to target actor
	//   - actorType: type of actor, obtained from actor template
	//   - event: event name to be handled
	//   - mw: message wrapper for routing
	Send(idOrSymbol, actorType, event string, mw *msg.Wrapper) error

	// Pub semantics for pubsub, used to publish messages to an actor's message cache queue
	Pub(topic string, event string, body []byte) error

	// AddressBook actor 地址管理对象
	AddressBook() IAddressBook

	//
	System() ISystem

	// Loader returns the actor loader
	Loader(string) IActorBuilder

	// Unregister unregisters an actor
	Unregister(id, ty string) error

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
	Complete(*msg.Wrapper)
	IsCompleted() bool

	Then(func(*msg.Wrapper)) IFuture
}

// ITimer interface for timer operations
type ITimer interface {
	// Stop stops the timer
	// Returns false if the timer has already been triggered or stopped
	Stop() bool

	// Reset resets the timer
	// interval: new interval duration (if 0, uses the existing interval)
	// Returns whether the reset was successful
	Reset(interval time.Duration) bool

	// IsActive checks if the timer is active
	IsActive() bool

	// Interval gets the current interval duration
	Interval() time.Duration

	// NextTrigger gets the next trigger time
	NextTrigger() time.Time

	// Execute executes the timer callback
	Execute() error
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
	Received(mw *msg.Wrapper) error

	// OnEvent registers an event handling chain for the actor
	OnEvent(ev string, createChainF func(ActorContext) IChain) error

	// OnTimer registers a timer function for the actor (Note: all times used here are in milliseconds)
	//  dueTime: delay before execution, 0 for immediate execution
	//  interval: time between each tick
	//  f: callback function
	//  args: can be used to pass the actor entity to the timer callback
	OnTimer(dueTime int64, interval int64, f func(interface{}) error, args interface{}) ITimer

	// CancelTimer cancels a timer
	CancelTimer(t ITimer)

	// SubscriptionEvent subscribes to a message
	//  If this is the first subscription to this topic, opts will take effect (you can set some options for the topic, such as ttl)
	//  topic: A subject that contains a group of channels (e.g., if topic = offline messages, channel = actorId, then each actor can get its own offline messages in this topic)
	//  channel: Represents different categories within a topic
	//  createChainF: Callback function for successful subscription
	Sub(topic string, channel string, createChainF func(ActorContext) IChain, opts ...pubsub.TopicOption) error

	// Call sends an event to another actor
	Call(idOrSymbol, actorType, event string, mw *msg.Wrapper) error

	ReenterCall(idOrSymbol, actorType, event string, mw *msg.Wrapper) IFuture

	Context() ActorContext

	Exit()
}

type IActorLoader interface {

	// Builder selects an actor from the factory and provides a builder
	Builder(string, ISystem) IActorBuilder

	// Pick selects an appropriate node for the actor builder to register
	Pick(context.Context, IActorBuilder) error

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
	Register(context.Context) (IActor, error)
	Picker(context.Context) error
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
