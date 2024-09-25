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
	core.IAddressBook
}

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

			// actor_ty := mw.Req.Header.Custom["actor_ty"]

			// 检查是否超过 limit 的限制
			// 要挑选一个权重低，且 actorty 相对注册少的节点

			fmt.Println("recv picker msg")

			// 指派到该节点进行注册
			sys.Call(ctx, router.Target{ID: "nodeid-register", Ty: def.ActorDynamicRegister, Ev: def.EvDynamicRegister}, mw)

			return nil
		},
	}
}
