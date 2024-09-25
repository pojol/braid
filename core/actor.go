package core

import (
	"context"

	"github.com/pojol/braid/lib/pubsub"
	"github.com/pojol/braid/lib/timewheel"
	"github.com/pojol/braid/router"
)

type IChain interface {
	Execute(context.Context, *router.MsgWrapper) error
}

// Users can define custom keys to pass required structures into the context

// StateKey is a custom type for the context key
type StateKey struct{}

// SystemKey is a custom type for the context key
type SystemKey struct{}

// GetState retrieves the state from the given context.
// If no state was set in the context, it returns nil.
//
// Parameters:
//   - ctx: The context.Context to retrieve the state from.
//
// Returns:
//   - The state stored in the context as an interface{}, or nil if not found.
//   - The returned value should be type-asserted to its original type before use.
func GetState(ctx context.Context) interface{} {
	if ctx == nil {
		return nil
	}
	return ctx.Value(StateKey{})
}

// GetSystem retrieves the ISystem from the given context.
// If no ISystem was set in the context, it returns nil.
//
// Parameters:
//   - ctx: The context.Context to retrieve the ISystem from.
//
// Returns:
//   - The ISystem stored in the context, or nil if not found.
func GetSystem(ctx context.Context) ISystem {
	if ctx == nil {
		return nil
	}
	return ctx.Value(SystemKey{}).(ISystem)
}

// IActor is an abstraction of threads (goroutines). In a Node (process),
// 1 to N actors execute specific business logic.
//
// Each actor object represents a logical computation unit that interacts
// with the outside world through a mailbox.
type IActor interface {
	Init()

	ID() string
	Type() string

	// 向 actor 的 mailbox 压入一条消息
	Received(msg *router.MsgWrapper) error

	// RegisterEvent registers an event handling chain for the actor
	RegisterEvent(ev string, createChainF func(context.Context) IChain) error

	// RegisterTimer registers a timer function for the actor (Note: all times used here are in milliseconds)
	//  dueTime: delay before execution, 0 for immediate execution
	//  interval: time between each tick
	//  f: callback function
	//  args: can be used to pass the actor entity to the timer callback
	RegisterTimer(dueTime int64, interval int64, f func() error, args interface{}) *timewheel.Timer

	// SubscriptionEvent subscribes to a message
	//  If this is the first subscription to this topic, opts will take effect (you can set some options for the topic, such as ttl)
	//  topic: A subject that contains a group of channels (e.g., if topic = offline messages, channel = actorId, then each actor can get its own offline messages in this topic)
	//  channel: Represents different categories within a topic
	//  succ: Callback function for successful subscription
	SubscriptionEvent(topic string, channel string, succ func(), opts ...pubsub.TopicOption) error

	// Actor 的主循环，它在独立的 goroutine 中运行
	Update()

	// Call 发送一个事件给另外一个 actor
	Call(ctx context.Context, tar router.Target, msg *router.MsgWrapper) error

	// SetContext returns a new context with the given state.
	// It allows you to embed any state information into the context for later retrieval.
	//
	// Parameters:
	//   - ctx: The parent context.Context to derive from.
	//   - state: The state information to store in the new context. Can be of any type.
	//
	// Returns:
	//   - A new context.Context that includes the provided state.
	SetContext(key, value interface{})

	Exit()
}

// ActorLoaderBuilder used to build ActorLoader
type ActorLoaderBuilder struct {
	CreateActorParm
	ISystem
	ActorConstructor
	IActorLoader
}

type IActorLoader interface {

	// Builder selects an actor from the factory and provides a builder
	Builder(string) *ActorLoaderBuilder

	// Pick selects an appropriate node for the actor builder to register
	Pick(*ActorLoaderBuilder) error
}

const (
	// Register the actor to the current node
	//  Note: It is still subject to the GlobalQuantityLimit restriction. If a limit is set and the number of registrations
	//  on other nodes has already exceeded this limit, the current registration will be skipped
	ActorRegisteraionType_Static = "Static"

	// The registration of dynamic nodes is executed through the `pick` function of the actor loader,
	// which analyzes the weight information of nodes in the cluster and the distribution of actors of the same type,
	// to select an appropriate node for registration

	// Dynamic and node-unique registration type
	ActorRegisteraionType_DynamicUnique = "DynamicUnique"
	// Dynamic and random node registration type
	ActorRegisteraionType_DynamicRandom = "DynamicRandom"
)

type ActorConstructor struct {
	// Weight occupied by the actor, weight algorithm reference: 2c4g (pod = 2 * 4 * 1000)
	Weight int
	// Constructor function
	Constructor CreateFunc

	// Registration types (globally unique, random node, must be current node)
	RegisteraionType string

	// Global quantity limit for the current actor type that can be registered
	GlobalQuantityLimit int
}

type IActorFactory interface {
	Get(ty string) *ActorConstructor
}
