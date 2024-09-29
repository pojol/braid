package actors

import (
	"context"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router"
)

type dynamicPickerActor struct {
	*actor.Runtime
}

func NewDynamicPickerActor(p *core.ActorLoaderBuilder) core.IActor {
	return &dynamicPickerActor{
		Runtime: &actor.Runtime{Id: p.ID, Ty: def.ActorDynamicPicker, Sys: p.ISystem},
	}
}

func (a *dynamicPickerActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)
	a.RegisterEvent(def.EvDynamicPick, MakeDynamicPick)
}

func MakeDynamicPick(actorCtx context.Context) core.IChain {
	return &actor.DefaultChain{

		Handler: func(mw *router.MsgWrapper) error {

			sys := core.GetSystem(actorCtx)
			actor := core.GetActor(actorCtx)

			actor_ty := mw.Req.Header.Custom["actor_ty"]

			// Select a node with low weight and relatively fewer registered actors of this type
			nodeaddr, err := sys.AddressBook().GetLowWeightNodeForActor(mw.Ctx, actor_ty)
			if err != nil {
				return err
			}

			// dispatcher to picker node
			return actor.Call(router.Target{ID: nodeaddr.Node + "_" + "register", Ty: def.ActorDynamicRegister, Ev: def.EvDynamicRegister}, mw)
		},
	}
}
