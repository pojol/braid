package mockdata

import (
	"context"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/router"
	"github.com/pojol/braid/router/msg"
)

type DefaultActorLoader struct {
	factory core.IActorFactory
}

func BuildDefaultActorLoader(factory core.IActorFactory) core.IActorLoader {
	return &DefaultActorLoader{factory: factory}
}

func (al *DefaultActorLoader) Pick(builder core.IActorBuilder) error {

	msgbuild := msg.NewBuilder(context.TODO())

	for key, value := range builder.GetOptions() {
		msgbuild.WithReqCustomFields(msg.Attr{Key: key, Value: value})
	}

	msgbuild.WithReqCustomFields(def.ActorID(builder.GetID()))
	msgbuild.WithReqCustomFields(def.ActorTy(builder.GetType()))

	go func() {
		err := builder.GetSystem().Call(router.Target{
			ID: def.SymbolWildcard,
			Ty: "MockDynamicPicker",
			Ev: "MockDynamicPick"},
			msgbuild.Build(),
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
		ISystem:          sys,
		ActorConstructor: *ac,
		IActorLoader:     al,
	}

	return builder
}

func (al *DefaultActorLoader) AssignToNode(node core.INode) {
	actors := al.factory.GetActors()

	for _, actor := range actors {

		if actor.Dynamic {
			continue
		}

		builder := al.Builder(actor.Name, node.System())
		if actor.ID == "" {
			actor.ID = actor.Name
		}

		builder.WithID(node.ID() + "_" + actor.ID)

		_, err := builder.Register()
		if err != nil {
			log.InfoF("assign to node build actor %s err %v", actor.Name, err)
		}
	}
}
