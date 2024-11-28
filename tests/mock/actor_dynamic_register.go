package mock

import (
	"context"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router/msg"
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

			Handler: func(mw *msg.Wrapper) error {

				actor_ty := msg.GetReqField[string](mw, def.KeyActorTy)
				actor_id := msg.GetReqField[string](mw, def.KeyActorID)

				builder := ctx.Loader(actor_ty)
				builder.WithID(actor_id)

				m, err := mw.GetReqCustomMap()
				if err != nil {
					return err
				}

				for k, v := range m {
					builder.WithOpt(k, v.(string))
				}

				_, err = builder.Register(mw.Ctx)
				if err != nil {
					return err
				}

				mw.Req.Header.PrevActorType = "MockDynamicRegister"
				return nil
			},
		}
	})
}
