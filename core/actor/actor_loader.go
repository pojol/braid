package actor

import "github.com/pojol/braid/core"

type DefaultActorLoader struct {
	sys     core.ISystem
	factory core.IActorFactory
}

func BuildDefaultActorLoader(sys core.ISystem, factory core.IActorFactory) core.IActorLoader {
	return &DefaultActorLoader{sys: sys, factory: factory}
}

func (al *DefaultActorLoader) Pick(ty string) *core.ActorLoaderBuilder {

	ac := al.factory.Get(ty)
	if ac == nil {
		return nil
	}

	builder := &core.ActorLoaderBuilder{
		CreateActorParm:  core.CreateActorParm{},
		ISystem:          al.sys,
		ActorConstructor: *ac,
	}

	builder.WithType(ty)

	return builder
}

func (al *DefaultActorLoader) AutoRegisterWithWeight() {

}
