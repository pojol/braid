package tests

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/core/node"
	"github.com/pojol/braid/tests/mock"
	"github.com/stretchr/testify/assert"
)

type mockTimerActor struct {
	*actor.Runtime
}

func newMockTimerActor(p core.IActorBuilder) core.IActor {
	return &mockTimerActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
	}
}

var tick1 int32
var tick2 int32
var tick3 int32
var tick4 int32
var tick5 int32

func (ta *mockTimerActor) Init(ctx context.Context) {
	ta.Runtime.Init(ctx)

	ta.RegisterTimer(0, 1000, func(i interface{}) error {
		atomic.AddInt32(&tick1, 1)
		return nil
	}, nil)

	ta.RegisterTimer(500, 500, func(i interface{}) error {
		atomic.AddInt32(&tick2, 1)
		return nil
	}, nil)

	ta.RegisterTimer(0, 100, func(i interface{}) error {
		atomic.AddInt32(&tick3, 1)
		return nil
	}, nil)

	ta.RegisterTimer(1000, 0, func(i interface{}) error {
		atomic.AddInt32(&tick4, 1)
		return nil
	}, nil)

	var t core.ITimer
	t = ta.RegisterTimer(0, 200, func(i interface{}) error {
		atomic.AddInt32(&tick5, 1)

		if atomic.LoadInt32(&tick5) == 5 {
			ta.RemoveTimer(t)
		}

		return nil
	}, nil)
}

func TestActorTimer1(t *testing.T) {

	factory := mock.BuildActorFactory()
	factory.Constructors["MockTimerActor"] = &core.ActorConstructor{
		ID:                  "MockTimerActor",
		Name:                "MockTimerActor",
		Weight:              20,
		Constructor:         newMockTimerActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 1,
		Dynamic:             false,
		Options:             make(map[string]string),
	}
	loader := mock.BuildDefaultActorLoader(factory)

	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-timer-1"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	nod.Init()

	t.Run("tick1", func(t *testing.T) {
		time.Sleep(time.Second * 5)
		tickcnt := atomic.LoadInt32(&tick1)
		assert.True(t, tickcnt >= int32(4) && tickcnt <= int32(6))
	})
}

func TestActorTimer2(t *testing.T) {

	factory := mock.BuildActorFactory()
	factory.Constructors["MockTimerActor"] = &core.ActorConstructor{
		ID:                  "MockTimerActor",
		Name:                "MockTimerActor",
		Weight:              20,
		Constructor:         newMockTimerActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 1,
		Dynamic:             false,
		Options:             make(map[string]string),
	}
	loader := mock.BuildDefaultActorLoader(factory)

	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-timer-2"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	nod.Init()

	t.Run("tick2", func(t *testing.T) {
		time.Sleep(time.Second * 5)
		tickcnt := atomic.LoadInt32(&tick2)
		targetcnt := int32(5*(1000/500) - 1)
		assert.True(t, tickcnt >= int32(targetcnt-1) && tickcnt <= int32(targetcnt+1))
	})
}

func TestActorTimer3(t *testing.T) {

	factory := mock.BuildActorFactory()
	factory.Constructors["MockTimerActor"] = &core.ActorConstructor{
		ID:                  "MockTimerActor",
		Name:                "MockTimerActor",
		Weight:              20,
		Constructor:         newMockTimerActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 1,
		Dynamic:             false,
		Options:             make(map[string]string),
	}
	loader := mock.BuildDefaultActorLoader(factory)

	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-timer-3"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	nod.Init()

	t.Run("tick3", func(t *testing.T) {
		time.Sleep(time.Second * 5)
		tickcnt := atomic.LoadInt32(&tick3)
		targetcnt := int32(5 * (1000 / 100))
		fmt.Println(tickcnt, targetcnt)
		assert.True(t, tickcnt >= int32(targetcnt-1) && tickcnt <= int32(targetcnt+1))
	})
}

func TestActorTimer4(t *testing.T) {

	factory := mock.BuildActorFactory()
	factory.Constructors["MockTimerActor"] = &core.ActorConstructor{
		ID:                  "MockTimerActor",
		Name:                "MockTimerActor",
		Weight:              20,
		Constructor:         newMockTimerActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 1,
		Dynamic:             false,
		Options:             make(map[string]string),
	}
	loader := mock.BuildDefaultActorLoader(factory)

	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-timer-4"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	nod.Init()

	t.Run("tick4", func(t *testing.T) {
		time.Sleep(time.Second * 3)
		assert.Equal(t, atomic.LoadInt32(&tick4), int32(1))
	})
}

func TestActorTimer5(t *testing.T) {

	factory := mock.BuildActorFactory()
	factory.Constructors["MockTimerActor"] = &core.ActorConstructor{
		ID:                  "MockTimerActor",
		Name:                "MockTimerActor",
		Weight:              20,
		Constructor:         newMockTimerActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 1,
		Dynamic:             false,
		Options:             make(map[string]string),
	}
	loader := mock.BuildDefaultActorLoader(factory)

	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-timer-5"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	nod.Init()

	t.Run("tick5", func(t *testing.T) {
		time.Sleep(time.Second * 3)
		assert.Equal(t, atomic.LoadInt32(&tick5), int32(5))
	})
}
