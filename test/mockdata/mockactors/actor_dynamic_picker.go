package mockactors

import (
	"context"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/router"
)

type dynamicPickerActor struct {
	*actor.Runtime
}

func NewDynamicPickerActor(p core.IActorBuilder) core.IActor {
	return &dynamicPickerActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: "MockDynamicPicker", Sys: p.GetSystem()},
	}
}

func (a *dynamicPickerActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	a.RegisterEvent("MockDynamicPick", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{

			Handler: func(mw *router.MsgWrapper) error {

				actor_ty := mw.Req.Header.Custom["actor_ty"]

				// Select a node with low weight and relatively fewer registered actors of this type
				nodeaddr, err := ctx.AddressBook().GetLowWeightNodeForActor(mw.Ctx, actor_ty)
				if err != nil {
					return err
				}

				// dispatcher to picker node
				return ctx.Call(router.Target{ID: nodeaddr.Node + "_" + "register", Ty: "MockDynamicRegister", Ev: "MockDynamicRegister"}, mw)
			},
		}
	})
}
