package mockdata

import (
	"context"
	fmt "fmt"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/router"
)

type DefaultActorLoader struct {
	factory core.IActorFactory
}

func BuildDefaultActorLoader(factory core.IActorFactory) core.IActorLoader {
	return &DefaultActorLoader{factory: factory}
}

func (al *DefaultActorLoader) Pick(builder core.IActorBuilder) error {

	customOptions := make(map[string]string)

	for key, value := range builder.GetOptions() {
		customOptions[key] = fmt.Sprint(value)
	}

	customOptions["actor_id"] = builder.GetID()
	customOptions["actor_ty"] = builder.GetType()

	go func() {
		err := builder.GetSystem().Call(router.Target{
			ID: def.SymbolWildcard,
			Ty: "MockDynamicPicker",
			Ev: "MockDynamicPick"},
			router.NewMsgWrap(context.TODO()).WithReqHeader(&router.Header{
				Custom: customOptions,
			}).Build(),
		)
		if err != nil {
			log.WarnF("[braid.actorLoader] call dynamic picker err %v", err.Error())
		}
	}()

	return nil
}

// Builder selects an actor from the factory and provides a builder
func (al *DefaultActorLoader) Builder(ty string, sys core.ISystem) core.IActorBuilder {
	ac := al.factory.Get(ty)
	if ac == nil {
		return nil
	}

	builder := &actor.ActorLoaderBuilder{
		CreateActorParm: actor.CreateActorParm{
			GenerationMode: actor.LocalGeneration, // Default to local option, can be modified using withpicker
			Options:        make(map[string]interface{}),
		},
		ISystem:          sys,
		ActorConstructor: *ac,
		IActorLoader:     al,
	}

	builder.WithType(ty)

	return builder
}
