package mockactors

import (
	"context"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/router"
)

type dynamicRegisterActor struct {
	*actor.Runtime
	loader core.IActorLoader
}

func NewDynamicRegisterActor(p core.IActorBuilder) core.IActor {
	return &dynamicRegisterActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: "MockDynamicRegister", Sys: p.GetSystem()},
		loader:  p.GetLoader(),
	}
}

func (a *dynamicRegisterActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	a.RegisterEvent("MockDynamicRegister", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{

			Handler: func(mw *router.MsgWrapper) error {

				actor_ty := mw.Req.Header.Custom["actor_ty"]
				actor_id := mw.Req.Header.Custom["actor_id"]

				builder := ctx.Loader(actor_ty)
				builder.WithID(actor_id)

				for k, v := range mw.Req.Header.Custom {
					builder.WithOpt(k, v)
				}

				actor, err := builder.Register()
				if err != nil {
					return err
				}

				mw.Req.Header.PrevActorType = "MockDynamicRegister"

				actor.Init(mw.Ctx)
				go actor.Update()

				return nil
			},
		}
	})
}
