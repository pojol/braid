package mockdata

import (
	"context"
	"fmt"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router"
)

type MockClacActor struct {
	*actor.Runtime
}

func NewClacActor(p core.IActorBuilder) core.IActor {
	return &MockClacActor{
		Runtime: &actor.Runtime{Id: p.GetType(), Ty: "MockClacActor"},
	}
}

func (a *MockClacActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	fmt.Println("init mock clac actor !")

	a.RegisterEvent("print", func(actorCtx context.Context) core.IChain {
		return &actor.DefaultChain{
			Handler: func(m *router.MsgWrapper) error {

				a.Call(router.Target{
					ID: "mockentity",
					Ty: def.MockActorEntity,
					Ev: "print",
				}, &router.MsgWrapper{Ctx: context.TODO(), Req: &router.Message{Header: &router.Header{}}})

				return nil
			},
		}
	})

}
