package actor

import (
	"context"
	"errors"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/router"
	"github.com/pojol/braid/router/msg"
)

type actorContext struct {
	ctx context.Context
}

func (ac *actorContext) Call(tar router.Target, mw *msg.Wrapper) error {
	actor, ok := ac.ctx.Value(actorKey{}).(core.IActor)
	if !ok {
		panic(errors.New("the actor instance does not exist in the ActorContext"))
	}

	return actor.Call(tar, mw)
}

func (ac *actorContext) CallBy(id string, ev string, mw *msg.Wrapper) error {
	actor, ok := ac.ctx.Value(actorKey{}).(core.IActor)
	if !ok {
		panic(errors.New("the actor instance does not exist in the ActorContext"))
	}

	if id == "" || ev == "" {
		panic(errors.New("callby parm err"))
	}

	return actor.Call(router.Target{ID: id, Ev: ev}, mw)
}

func (ac *actorContext) ID() string {
	actor, ok := ac.ctx.Value(actorKey{}).(core.IActor)
	if !ok {
		panic(errors.New("the actor instance does not exist in the ActorContext"))
	}

	return actor.ID()
}

func (ac *actorContext) Type() string {
	actor, ok := ac.ctx.Value(actorKey{}).(core.IActor)
	if !ok {
		panic(errors.New("the actor instance does not exist in the ActorContext"))
	}

	return actor.Type()
}

func (ac *actorContext) ReenterCall(ctx context.Context, tar router.Target, mw *msg.Wrapper) core.IFuture {
	actor, ok := ac.ctx.Value(actorKey{}).(core.IActor)
	if !ok {
		panic(errors.New("the actor instance does not exist in the ActorContext"))
	}

	return actor.ReenterCall(ctx, tar, mw)
}

func (ac *actorContext) Send(tar router.Target, mw *msg.Wrapper) error {
	sys, ok := ac.ctx.Value(systemKey{}).(core.ISystem)
	if !ok {
		panic(errors.New("the system instance does not exist in the ActorContext"))
	}

	return sys.Send(tar, mw)
}

func (ac *actorContext) Unregister(id, ty string) error {
	sys, ok := ac.ctx.Value(systemKey{}).(core.ISystem)
	if !ok {
		panic(errors.New("the system instance does not exist in the ActorContext"))
	}

	return sys.Unregister(id, ty)
}

func (ac *actorContext) Pub(topic string, msg *router.Message) error {
	sys, ok := ac.ctx.Value(systemKey{}).(core.ISystem)
	if !ok {
		panic(errors.New("the system instance does not exist in the ActorContext"))
	}

	return sys.Pub(topic, msg)
}

func (ac *actorContext) AddressBook() core.IAddressBook {
	sys, ok := ac.ctx.Value(systemKey{}).(core.ISystem)
	if !ok {
		panic(errors.New("the system instance does not exist in the ActorContext"))
	}

	return sys.AddressBook()
}

func (ac *actorContext) System() core.ISystem {
	sys, ok := ac.ctx.Value(systemKey{}).(core.ISystem)
	if !ok {
		panic(errors.New("the system instance does not exist in the ActorContext"))
	}

	return sys
}

func (ac *actorContext) Loader(actorType string) core.IActorBuilder {
	sys, ok := ac.ctx.Value(systemKey{}).(core.ISystem)
	if !ok {
		panic(errors.New("the system instance does not exist in the ActorContext"))
	}

	return sys.Loader(actorType)
}

func (ac *actorContext) WithValue(key, value interface{}) {
	ac.ctx = context.WithValue(ac.ctx, key, value)
}

func (ac *actorContext) GetValue(key interface{}) interface{} {
	return ac.ctx.Value(key)
}