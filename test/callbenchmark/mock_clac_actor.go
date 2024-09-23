package callbenchmark

import (
	"context"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router"
)

type mockEntityActor struct {
	*actor.Runtime
}

func (a *mockEntityActor) Init() {
	a.Runtime.Init()

	a.RegisterEvent("print", func(actorCtx context.Context) core.IChain {
		return &actor.DefaultChain{
			Handler: func(ctx context.Context, m *router.MsgWrapper) error {

				a.Call(ctx, router.Target{
					ID: "mockentity",
					Ty: def.MockActorEntity,
					Ev: "print",
				}, &router.MsgWrapper{Req: &router.Message{Header: &router.Header{}}})

				return nil
			},
		}
	})

}
