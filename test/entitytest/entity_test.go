package entitytest

import (
	"context"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/router"
	"github.com/stretchr/testify/assert"
)

func TestEntity(t *testing.T) {

	// use mock redis
	redis.BuildClientWithOption(redis.WithAddr("redis://127.0.0.1:6379/0"))

	warp1 := NewEntityWapper("test.actor.1")
	warp1.Airship = &EntityAirshipModule{
		ID: "test.actor.1",
	}
	warp1.Bag = &EntityBagModule{
		Bag: make(map[int32]*ItemList),
	}
	warp1.TimeInfo = &EntityTimeInfoModule{
		CreateTime: time.Now().Unix(),
	}
	warp1.User = &EntityUserModule{
		GameID: "test.actor.1",
		Token:  "111",
	}

	warp1.Bag.Bag[1001] = &ItemList{
		Items: []*Item{
			{
				ID:     "1001",
				Num:    10,
				DictID: 1001,
			},
		},
	}

	warp1.Sync()

	//////////////////////////////////////////
	sys := node.BuildSystemWithOption(
		node.SystemActorConstructor(
			[]node.ActorConstructor{
				{Type: "mockUserActor", Constructor: func(p *core.CreateActorParm) core.IActor {
					return &mockUserActor{
						Runtime: &actor.Runtime{Ty: "mockUserActor", Sys: p.Sys},
						entity:  NewEntityWapper(p.ID),
					}
				}},
			},
		),
		node.SystemService("service_1", "node_1"),
	)

	sys.Register(context.TODO(), "mockUserActor", core.CreateActorWithID("test.actor.1"))

	for _, a := range sys.Actors() {
		a.Init()
		go a.Update()
	}

	time.Sleep(time.Second * 1)

	msg := router.GetMsgWithPool()
	sys.Call(context.TODO(), router.Target{ID: "test.actor.1", Ty: "mockUserActor", Ev: "entity_test"}, msg)

	assert.Equal(t, msg.Res.Header.Custom["code"], "200")
	router.PutMsg(msg)

	time.Sleep(time.Second * 2)

}
