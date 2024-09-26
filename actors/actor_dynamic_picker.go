package actors

import (
	"context"
	"fmt"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router"
)

type dynamicPickerActor struct {
	*actor.Runtime
}

type addressbookTy struct{}

func NewDynamicPickerActor(p *core.ActorLoaderBuilder) core.IActor {
	return &dynamicPickerActor{
		Runtime: &actor.Runtime{Id: p.ID, Ty: def.ActorDynamicRegister, Sys: p.ISystem},
	}
}

func (a *dynamicPickerActor) Init() {
	a.Runtime.Init()
	a.RegisterEvent(def.EvDynamicPick, MakeDynamicPick)
}

func MakeDynamicPick(actorCtx context.Context) core.IChain {
	return &actor.DefaultChain{

		Handler: func(ctx context.Context, mw *router.MsgWrapper) error {

			sys := core.GetSystem(actorCtx)

			actor_ty := mw.Req.Header.Custom["actor_ty"]
			fmt.Println("recv pick event", actor_ty)

			// Select a node with low weight and relatively fewer registered actors of this type
			nodeaddr, err := sys.AddressBook().GetLowWeightNodeForActor(ctx, actor_ty)
			if err != nil {
				return err
			}

			// dispatcher to picker node
			fmt.Println("dynamic picker", actor_ty, "=>", nodeaddr.Node+"_"+"register")
			return sys.Call(ctx, router.Target{ID: nodeaddr.Node + "_" + "register", Ty: def.ActorDynamicRegister, Ev: def.EvDynamicRegister}, mw)
		},
	}
}
