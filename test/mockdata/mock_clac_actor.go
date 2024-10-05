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
		Runtime: &actor.Runtime{Id: p.GetType(), Ty: "MockClacActor", Sys: p.GetSystem()},
	}
}

func (a *MockClacActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	fmt.Println("init", a.Id, "actor succ!")

	a.RegisterEvent("print", func(actorCtx core.ActorContext) core.IChain {
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
	a.RegisterEvent("clac", func(actorCtx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(mw *router.MsgWrapper) error {

				// 2.
				fmt.Println(actorCtx.GetID(), "recv clac event")
				return nil
			},
		}
	})

	a.RegisterEvent("mockreenter", MakeEvReenter)
}

func MakeEvReenter(actorCtx core.ActorContext) core.IChain {
	return &actor.DefaultChain{
		Handler: func(mw *router.MsgWrapper) error {

			// Initiate an asynchronous re-entrant call
			// 发起一次异步可重入调用
			future := actorCtx.ReenterCall(mw.Ctx, router.Target{ID: "clac-2", Ty: def.MockActorEntity, Ev: "clac"}, mw)

			// Register callback functions to handle the result after the asynchronous call completes. Note:
			//  1. The Then method itself is a synchronous call, returning immediately.
			//  2. The asynchronous operation represented by the future is executed in parallel.
			//  3. Callback functions are called synchronously in sequence after the future completes.
			// 注册回调函数，在异步调用完成后处理结果， 注意：
			//  1. Then 方法本身是同步调用，立即返回。
			//  2. future 代表的异步操作是并行执行的。
			//  3. 回调函数会在 future 完成后被依次同步调用
			future.Then(func(ret *router.MsgWrapper) {

				// 3.
				fmt.Println(actorCtx.GetID(), "call clac event callback!", ret.Err)

			}).Then(func(ret *router.MsgWrapper) {
				// Chained call
				// 链式调用
			})

			// 1.
			fmt.Println(actorCtx.GetID(), "call clac event completed! but not callback")
			// Note: This returns immediately, not waiting for the asynchronous operation to complete
			// 注意：这里立即返回，不等待异步操作完成
			return nil
		},
	}
}
