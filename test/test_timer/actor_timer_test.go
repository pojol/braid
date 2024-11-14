package testtimer

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/test/mockdata"
)

func TestMain(m *testing.M) {
	slog, _ := log.NewServerLogger("test")
	log.SetSLog(slog)

	defer log.Sync()

	os.Exit(m.Run())
}

type MockTimerActor struct {
	*actor.Runtime
}

func NewMockTimerActor(p core.IActorBuilder) core.IActor {
	return &MockTimerActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
	}
}

func (ta *MockTimerActor) Init(ctx context.Context) {
	ta.Runtime.Init(ctx)

	ta.RegisterTimer(0, 1000, func(i interface{}) error {
		fmt.Println("mock timer handler tick")
		return nil
	}, nil)
}

func TestActorTimer(t *testing.T) {
	redis.BuildClientWithOption(redis.WithAddr("redis://127.0.0.1:6379/0"))
	redis.FlushAll(context.TODO()) // clean cache

	factory := mockdata.BuildActorFactory()
	factory.Constructors["MockTimerActor"] = &core.ActorConstructor{
		ID:                  "MockTimerActor",
		Name:                "MockTimerActor",
		Weight:              20,
		Constructor:         NewMockTimerActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 1,
		Dynamic:             false,
		Options:             make(map[string]string),
	}
	loader := mockdata.BuildDefaultActorLoader(factory)

	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-timer-1"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	nod.Init()
	nod.Update()

	time.Sleep(time.Second * 10)
}
