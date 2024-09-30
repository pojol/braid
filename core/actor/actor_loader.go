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

func (al *DefaultActorLoader) Pick(builder *core.ActorLoaderBuilder) error {

	customOptions := make(map[string]string)

	for key, value := range builder.Options {
		customOptions[key] = fmt.Sprint(value)
	}

	customOptions["actor_id"] = builder.ID
	customOptions["actor_ty"] = builder.ActorTy

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
func (al *DefaultActorLoader) Builder(ty string) *core.ActorLoaderBuilder {
	ac := al.factory.Get(ty)
	if ac == nil {
		return nil
	}

	builder := &core.ActorLoaderBuilder{
		CreateActorParm: core.CreateActorParm{
			Options: make(map[string]interface{}),
		},
		ISystem:          al.sys,
		ActorConstructor: *ac,
		IActorLoader:     al,
	}

	builder.WithType(ty)

	return builder
}