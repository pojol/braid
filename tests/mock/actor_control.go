package mock

import (
	"context"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/router/msg"
)

type controlActor struct {
	*actor.Runtime
}

func NewControlActor(p core.IActorBuilder) core.IActor {
	return &controlActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: "MockActorControl", Sys: p.GetSystem()},
	}
}

func (a *controlActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	a.OnEvent("MockUnregister", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(mw *msg.Wrapper) error {

				actor_id := msg.GetReqCustomField[string](mw, "actor_id")
				actor_ty := msg.GetReqCustomField[string](mw, "actor_ty")

				err := ctx.Unregister(actor_id, actor_ty)
				if err != nil {
					log.WarnF("[braid.actor_control] unregister actor %v err %v", actor_id, err)
				}

				return nil
			},
		}
	})
}
