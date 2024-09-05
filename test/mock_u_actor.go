package nodeprocess

import (
	"context"
	"fmt"

	"github.com/pojol/braid/core/workerthread"
	"github.com/pojol/braid/router"
)

type mockEntityActor struct {
	*workerthread.BaseActor
}

func (a *mockEntityActor) Init() {
	a.BaseActor.Init()

	a.RegisterEvent("print", &workerthread.DefaultChain{
		Handler: func(ctx context.Context, m *router.MsgWrapper) error {

			fmt.Println("entity actor recved:", string(m.Req.Body))

			return nil
		},
	})
}
