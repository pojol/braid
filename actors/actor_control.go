package actors

import (
	"context"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/router"
)

type controlActor struct {
	*actor.Runtime
}

func NewControlActor(p core.IActorBuilder) core.IActor {
	return &controlActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: def.ActorControl, Sys: p.GetSystem()},
	}
}

func (a *controlActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	a.RegisterEvent(def.EvUnregister, MakeUnregister)
}

func MakeUnregister(ctx core.ActorContext) core.IChain {
	return &actor.DefaultChain{
		Handler: func(mw *router.MsgWrapper) error {

			actor_id := mw.Req.Header.Custom["actor_id"]

			err := ctx.Unregister(actor_id)
			if err != nil {
				log.WarnF("[braid.actor_control] unregister actor %v err %v", actor_id, err)
			}

			return nil
		},
	}
}
