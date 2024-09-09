package entitytest

import (
	"context"
	"fmt"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/router"
)

type mockUserActor struct {
	*actor.Runtime
	entity *EntityWapper
}

func NewEntity(p *core.CreateActorParm) core.IActor {
	return &mockUserActor{
		Runtime: &actor.Runtime{Id: p.ID, Ty: "mockUserActor"},
		entity:  NewEntityWapper(p.ID),
	}
}

func (a *mockUserActor) Init() {
	a.Runtime.Init()
	err := a.entity.Load()
	if err != nil {
		panic(fmt.Errorf("load user actor err %v", err.Error()))
	}

	// 实现各种事件
	a.RegisterEvent("entity_test", &actor.DefaultChain{
		Handler: func(ctx context.Context, m *router.MsgWrapper) error {

			if a.entity.Bag.EnoughItem(1001, 10) {
				a.entity.Bag.ConsumeItem(1001, 5, "test", "")

				// 标记成功
				m.Res.Header.Custom["code"] = "200"
			}

			return nil
		},
	})

	// 1分钟尝试一次将脏
	a.RegisterTimer(0, 1000*60, func() error {
		a.entity.Sync()

		return nil
	}, nil)
}
