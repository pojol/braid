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

func NewDynamicRegisterActor(p core.IActorBuilder) core.IActor {
	return &dynamicRegisterActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: def.ActorDynamicRegister, Sys: p.GetSystem()},
		loader:  p.GetLoader(),
	}
}

func (a *dynamicRegisterActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	a.RegisterEvent(def.EvDynamicRegister, MakeDynamicRegister)
}

func MakeDynamicRegister(ctx core.ActorContext) core.IChain {
	return &actor.DefaultChain{

		Handler: func(mw *router.MsgWrapper) error {

			actor_ty := mw.Req.Header.Custom["actor_ty"]
			actor_id := mw.Req.Header.Custom["actor_id"]

			builder := ctx.Loader(actor_ty)
			builder.WithID(actor_id)

			for k, v := range mw.Req.Header.Custom {
				builder.WithOpt(k, v)
			}

			actor, err := builder.Build()
			if err != nil {
				return err
			}

			mw.Req.Header.PrevActorType = def.ActorDynamicRegister

			actor.Init(mw.Ctx)
			go actor.Update()

			return nil
		},
	}
}
