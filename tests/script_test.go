package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/core/node"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router/msg"
	"github.com/pojol/braid/tests/mock"
)

type MockScriptActor struct {
	*actor.Runtime
}

func NewMockScriptActor(p core.IActorBuilder) core.IActor {
	return &MockScriptActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
	}
}

func (sa *MockScriptActor) Init(ctx context.Context) {
	sa.Runtime.Init(ctx)

	sa.RegisterEvent("test_script", func(ctx core.ActorContext) core.IChain {

		scriptHandler, err := actor.NewScriptHandler("mock")
		if err != nil {
			panic(fmt.Errorf("mock script actor registr script handler err %v", err.Error()))
		}

		return &actor.DefaultChain{
			Script: scriptHandler,
		}
	})
}

func TestScript(t *testing.T) {
	redis.BuildClientWithOption(redis.WithAddr("redis://127.0.0.1:6379/0"))
	redis.FlushAll(context.TODO()) // clean cache

	factory := mock.BuildActorFactory()
	factory.Constructors["MockScriptActor"] = &core.ActorConstructor{
		ID:                  "MockScriptActor",
		Name:                "MockScriptActor",
		Weight:              20,
		Constructor:         NewMockScriptActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 1,
		Dynamic:             false,
		Options:             make(map[string]string),
	}
	loader := mock.BuildDefaultActorLoader(factory)

	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-script-1"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	nod.Init()

	time.Sleep(time.Second)

	nod.System().Call(def.SymbolLocalFirst, "MockScriptActor", "test_script",
		msg.NewBuilder(context.Background()).WithReqCustomFields(msg.Attr{Key: "test", Value: "hello braid script"}).Build(),
	)

	time.Sleep(time.Second)
}
