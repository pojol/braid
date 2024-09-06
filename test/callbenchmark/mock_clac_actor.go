package callbenchmark

import (
	"context"

	"github.com/pojol/braid/core/workerthread"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router"
)

type mockEntityActor struct {
	*workerthread.BaseActor
}

func (a *mockEntityActor) Init() {
	a.BaseActor.Init()

	a.RegisterEvent("print", &workerthread.DefaultChain{
		Handler: func(ctx context.Context, m *router.MsgWrapper) error {

			a.Call(ctx, router.Target{
				ID: "mockentity",
				Ty: def.MockActorEntity,
				Ev: "print",
			}, &router.MsgWrapper{Req: &router.Message{Header: &router.Header{}}})

			return nil
		},
	})

}
