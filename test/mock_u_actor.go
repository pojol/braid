package nodeprocess

import (
	"context"
	"fmt"

	"github.com/pojol/braid/core/workerthread"
	"github.com/pojol/braid/router"
)

type userActorProxy struct {
	*workerthread.BaseActor
}

func (a *userActorProxy) Init() {
	a.BaseActor.Init()

	a.RegisterEventChain("print", &workerthread.DefaultChain{
		Handler: func(ctx context.Context, m *router.MsgWrapper) error {

			fmt.Println("entity actor recved:", string(m.Req.Body))

			return nil
		},
	})
}
