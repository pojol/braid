package mockactors

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router"
	"github.com/pojol/braid/router/msg"
)

type MockClacActor struct {
	*actor.Runtime
}

func NewClacActor(p core.IActorBuilder) core.IActor {
	return &MockClacActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
	}
}

var GlobalCreateCnt = int32(0)

func (a *MockClacActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	atomic.AddInt32(&GlobalCreateCnt, 1)

	a.RegisterEvent("print", func(actorCtx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(mw *msg.Wrapper) error {

				a.Call("mockentity", def.MockActorEntity, "print",
					&msg.Wrapper{Ctx: context.TODO(), Req: &router.Message{Header: &router.Header{}}})

				return nil
			},
		}
	})
	a.RegisterEvent("clac", func(actorCtx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(mw *msg.Wrapper) error {

				// 2.
				fmt.Println(actorCtx.ID(), "recv clac event")
				return nil
			},
		}
	})

	a.RegisterEvent("mockreenter", MakeEvReenter)
}

func MakeEvReenter(actorCtx core.ActorContext) core.IChain {
	return &actor.DefaultChain{
		Handler: func(mw *msg.Wrapper) error {

			// Initiate an asynchronous re-entrant call
			// 发起一次异步可重入调用
			future := actorCtx.ReenterCall(mw.Ctx, "clac-2", def.MockActorEntity, "clac", mw)

			// Register callback functions to handle the result after the asynchronous call completes. Note:
			//  1. The Then method itself is a synchronous call, returning immediately.
			//  2. The asynchronous operation represented by the future is executed in parallel.
			//  3. Callback functions are called synchronously in sequence after the future completes.
			// 注册回调函数，在异步调用完成后处理结果， 注意：
			//  1. Then 方法本身是同步调用，立即返回。
			//  2. future 代表的异步操作是并行执行的。
			//  3. 回调函数会在 future 完成后被依次同步调用
			future.Then(func(ret *msg.Wrapper) {

				// 3.
				fmt.Println(actorCtx.ID(), "call clac event callback!", ret.Err)

			}).Then(func(ret *msg.Wrapper) {
				// Chained call
				// 链式调用
			})

			// 1.
			fmt.Println(actorCtx.ID(), "call clac event completed! but not callback")
			// Note: This returns immediately, not waiting for the asynchronous operation to complete
			// 注意：这里立即返回，不等待异步操作完成
			return nil
		},
	}
}
