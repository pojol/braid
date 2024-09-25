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

func NewUserActor(p *core.ActorLoaderBuilder) core.IActor {
	return &MockUserActor{
		Runtime: &actor.Runtime{Id: p.ID, Ty: "MockUserActor"},
		State:   NewEntityWapper(p.ID),
	}
}

func (a *MockUserActor) Init() {
	a.Runtime.Init()
	err := a.State.Load(context.TODO())
	if err != nil {
		panic(fmt.Errorf("load user actor err %v", err.Error()))
	}

	// Implement events
	a.RegisterEvent("entity_test", func(actorCtx context.Context) core.IChain {
		return &actor.DefaultChain{
			Handler: func(ctx context.Context, m *router.MsgWrapper) error {

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
