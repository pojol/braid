package actors

import (
	"context"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router"
)

type dynamicRegisterActor struct {
	*actor.Runtime
	loader core.IActorLoader
}

type loadKey struct{}

func NewDynamicRegisterActor(p *core.ActorLoaderBuilder) core.IActor {
	return &dynamicRegisterActor{
		Runtime: &actor.Runtime{Id: p.ID, Ty: def.ActorDynamicRegister, Sys: p.ISystem},
		loader:  p.IActorLoader,
	}
}

func (a *dynamicRegisterActor) Init() {
	a.Runtime.Init()
	a.SetContext(loadKey{}, a.loader)

	a.RegisterEvent(def.EvDynamicRegister, MakeDynamicRegister)
}

func MakeDynamicRegister(actorCtx context.Context) core.IChain {
	return &actor.DefaultChain{

		Handler: func(ctx context.Context, mw *router.MsgWrapper) error {

			loader := actorCtx.Value(loadKey{}).(core.IActorLoader)

			actor_ty := mw.Req.Header.Custom["actor_ty"]
			actor_id := mw.Req.Header.Custom["actor_id"]

			builder := loader.Builder(actor_ty)
			builder.WithID(actor_id)

			for k, v := range mw.Req.Header.Custom {
				builder.WithOpt(k, v)
			}

			actor, err := builder.RegisterLocally()
			if err != nil {
				return err
			}

			actor.Init()
			go actor.Update()

			return nil
		},
	}
}
