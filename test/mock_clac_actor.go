package nodeprocess

import (
	"context"

	"github.com/pojol/braid/core/workerthread"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router"
)

type clacActorProxy struct {
	*workerthread.BaseActor
}

func (a *clacActorProxy) Init() {
	a.BaseActor.Init()

	a.RegisterEventChain("clacA", &workerthread.DefaultChain{
		Before: []workerthread.MiddlewareHandler{},
		Handler: func(ctx context.Context, m *router.MsgWrapper) error {

			//entity := m.Entity.(*PlayerEntity)

			// Example: 向集群中任意actor发送消息，在本协程中阻塞等待返回
			a.Call(ctx, router.Target{
				ID: "mockentity",
				Ty: def.MockActorEntity,
				Ev: "print",
			}, &router.MsgWrapper{Req: &router.Message{Body: []byte("hello,entity!"), Header: &router.Header{}}})

			return nil
		},
	})

}
