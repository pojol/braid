package tests

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/router/msg"
	"github.com/pojol/braid/tests/mock"
)

const reenterActorName = "MockReenterActor"
const mockaName = "mocka"
const mockbName = "mockb"

var factory *mock.MockActorFactory
var loader core.IActorLoader

func TestMain(m *testing.M) {
	slog, _ := log.NewServerLogger("test")
	log.SetSLog(slog)

	defer log.Sync()

	factory = mock.BuildActorFactory()
	loader = mock.BuildDefaultActorLoader(factory)
	factory.Constructors[reenterActorName] = &core.ActorConstructor{
		ID:          reenterActorName,
		Name:        reenterActorName,
		Weight:      80,
		NodeUnique:  true,
		Constructor: newMockReenterActor,
	}
	factory.Constructors[mockaName] = &core.ActorConstructor{
		ID:          mockaName,
		Name:        mockaName,
		Weight:      80,
		NodeUnique:  true,
		Constructor: newMockA,
	}
	factory.Constructors[mockbName] = &core.ActorConstructor{
		ID:          mockbName,
		Name:        mockbName,
		Weight:      80,
		NodeUnique:  true,
		Constructor: newMockB,
	}

	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer mr.Close()
	redis.BuildClientWithOption(redis.WithAddr(fmt.Sprintf("redis://%s", mr.Addr())))

	os.Exit(m.Run())
}

type mockReenterActor struct {
	*actor.Runtime
}

var calcValue int32

func newMockReenterActor(p core.IActorBuilder) core.IActor {
	return &mockReenterActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
	}
}

func (ra *mockReenterActor) Init(ctx context.Context) {
	ra.Runtime.Init(ctx)

	ra.RegisterEvent("reenter", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				calculateVal := 2
				w.ToBuilder().WithReqCustomFields(msg.Attr{Key: "calculateVal", Value: calculateVal})

				future := ctx.ReenterCall(w.Ctx, mockaName, mockaName, "clac", w)
				future.Then(func(w *msg.Wrapper) {
					val := msg.GetResField[int](w, "calculateVal")
					atomic.CompareAndSwapInt32(&calcValue, 0, int32(val*2))
				})

				return nil
			},
		}
	})

	ra.RegisterEvent("timeout", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				future := ctx.ReenterCall(w.Ctx, mockaName, mockaName, "timeout", w)
				future.Then(func(w *msg.Wrapper) {
					fmt.Println("timeout then", w.Err)
				})
				return nil
			},
		}
	})
}

type mockActorA struct {
	*actor.Runtime
}

func newMockA(p core.IActorBuilder) core.IActor {
	return &mockActorA{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
	}
}

func (a *mockActorA) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	a.RegisterEvent("clac", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				val := msg.GetReqField[int](w, "calculateVal")
				w.ToBuilder().WithResCustomFields(msg.Attr{Key: "calculateVal", Value: val + 2})

				return nil
			},
		}
	})

	a.RegisterEvent("timeout", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				time.Sleep(time.Second * 5)
				return nil
			},
		}
	})
}

type mockActorB struct {
	*actor.Runtime
}

func newMockB(p core.IActorBuilder) core.IActor {
	return &mockActorB{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
	}
}

func (a *mockActorB) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	a.RegisterEvent("ping", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				w.ToBuilder().WithResCustomFields(msg.Attr{Key: "pong", Value: "pong"})
				return nil
			},
		}
	})
}
