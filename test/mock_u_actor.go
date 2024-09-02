package nodeprocess

import (
	"braid/core/actor"
	"braid/router"
	"context"
	"fmt"
)

type userActorProxy struct {
	*actor.BaseActor
}

func (a *userActorProxy) Init() {
	a.BaseActor.Init()

	a.RegisterEventChain("print", &actor.DefaultChain{
		Handler: func(ctx context.Context, m *router.MsgWrapper) error {

			fmt.Println("entity actor recved:", string(m.Req.Body))

			return nil
		},
	})
}
