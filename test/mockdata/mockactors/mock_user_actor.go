package mockactors

import (
	"context"
	"fmt"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/router/msg"
	"github.com/pojol/braid/test/mockdata/mockentity"
)

type MockUserActor struct {
	*actor.Runtime
	State *mockentity.EntityWapper
}

func NewUserActor(p core.IActorBuilder) core.IActor {
	return &MockUserActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: "MockUserActor"},
		State:   mockentity.NewEntityWapper(p.GetID()),
	}
}

func (a *MockUserActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)
	err := a.State.Load(context.TODO())
	if err != nil {
		panic(fmt.Errorf("load user actor err %v", err.Error()))
	}

	// Implement events
	a.RegisterEvent("entity_test", func(actorCtx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(m *msg.Wrapper) error {

				if a.State.Bag.EnoughItem("1001", 10) {
					a.State.Bag.ConsumeItem("1001", 5, "test", "")

					// mark success
					fmt.Println("entity_test consume item success")
					m.ToBuilder().WithResCustomFields(msg.Attr{Key: "code", Value: "200"})
					fmt.Println("build msg err", m.Err, msg.GetResField[string](m, "code"))
				}

				return nil
			},
		}
	})

	// one minute try sync to cache
	a.RegisterTimer(0, 1000*60, func(args interface{}) error {
		a.State.Sync(context.TODO(), false)

		return nil
	}, nil)
}
