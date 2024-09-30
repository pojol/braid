package actor

import (
	"context"
	"fmt"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/router"
)

type DefaultActorLoader struct {
	sys     core.ISystem
	factory core.IActorFactory
}

func BuildDefaultActorLoader(sys core.ISystem, factory core.IActorFactory) core.IActorLoader {
	return &DefaultActorLoader{sys: sys, factory: factory}
}

func (al *DefaultActorLoader) Pick(builder core.IActorBuilder) error {

	customOptions := make(map[string]string)

	for key, value := range builder.GetOptions() {
		customOptions[key] = fmt.Sprint(value)
	}

	customOptions["actor_id"] = builder.GetID()
	customOptions["actor_ty"] = builder.GetType()

	go func() {
		err := al.sys.Call(router.Target{
			ID: def.SymbolWildcard,
			Ty: def.ActorDynamicPicker,
			Ev: def.EvDynamicPick},
			router.NewMsgWrap(context.TODO()).WithReqHeader(&router.Header{
				Custom: customOptions,
			}).Build(),
		)
		if err != nil {
			log.Warn("[braid.actorLoader] call synamic picker err %v", err.Error())
		}
	}()

	return nil
}

// Builder selects an actor from the factory and provides a builder
func (al *DefaultActorLoader) Builder(ty string) core.IActorBuilder {
	ac := al.factory.Get(ty)
	if ac == nil {
		return nil
	}

	builder := &ActorLoaderBuilder{
		CreateActorParm: CreateActorParm{
			GenerationMode: LocalGeneration, // Default to local option, can be modified using withpicker
			Options:        make(map[string]interface{}),
		},
		ISystem:          al.sys,
		ActorConstructor: *ac,
		IActorLoader:     al,
	}

	builder.WithType(ty)

	return builder
}
