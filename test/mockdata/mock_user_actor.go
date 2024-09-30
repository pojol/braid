package mockdata

import (
	"context"
	"fmt"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/router"
)

type MockUserActor struct {
	*actor.Runtime
	State *EntityWapper
}

func NewUserActor(p core.IActorBuilder) core.IActor {
	return &MockUserActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: "MockUserActor"},
		State:   NewEntityWapper(p.GetID()),
	}
}

func (a *MockUserActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)
	err := a.State.Load(context.TODO())
	if err != nil {
		panic(fmt.Errorf("load user actor err %v", err.Error()))
	}

	// Implement events
	a.RegisterEvent("entity_test", func(actorCtx context.Context) core.IChain {
		return &actor.DefaultChain{
			Handler: func(m *router.MsgWrapper) error {

				if a.State.Bag.EnoughItem("1001", 10) {
					a.State.Bag.ConsumeItem("1001", 5, "test", "")

					// mark success
					fmt.Println("entity_test consume item success")
					m.Res.Header.Custom["code"] = "200"
				}

				return nil
			},
		}
	})

	// one minute try sync to cache
	a.RegisterTimer(0, 1000*60, func() error {
		a.State.Sync(context.TODO())

		return nil
	}, nil)
}
