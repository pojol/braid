package entitytest

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/mgo"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/router"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	slog, _ := log.NewServerLogger("test")
	log.SetSLog(slog)

	defer log.Sync()

	os.Exit(m.Run())
}

func mockEntity2DB(id string) {
	warp1 := NewEntityWapper(id)
	warp1.Airship = &EntityAirshipModule{
		ID: id,
	}
	warp1.Bag = &EntityBagModule{
		ID:  id,
		Bag: make(map[string]*ItemList),
	}
	warp1.TimeInfo = &EntityTimeInfoModule{
		ID:         id,
		CreateTime: time.Now().Unix(),
	}
	warp1.User = &EntityUserModule{
		ID:    id,
		Token: "111",
	}

	warp1.Bag.Bag["1001"] = &ItemList{
		Items: []*Item{
			{
				ID:     "1001",
				Num:    10,
				DictID: "1001",
			},
		},
	}

	mgo.Collection("braid-test", "entity_test").InsertOne(context.TODO(), warp1)
}

func mockEntity(id string) core.ISystem {
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

	sys.Register(context.TODO(), "mockUserActor", core.CreateActorWithID(id))

	for _, a := range sys.Actors() {
		a.Init()
		go a.Update()
	}

	return sys
}

func TestEntityLoad(t *testing.T) {

	// use mock redis
	redis.BuildClientWithOption(redis.WithAddr("redis://127.0.0.1:6379/0"))
	mgo.Build(mgo.AppendConn(mgo.ConnInfo{
		Name: "braid-test",
		Addr: "mongodb://127.0.0.1:27017",
	}))

	id := "test.actor.1"
	ty := "mockUserActor"

	mockEntity2DB(id)

	redis.FlushAll(context.TODO()) // clean cache

	//////////////////////////////////////////

	// load entity with db and sync to redis
	sys := mockEntity(id)

	msg := router.NewMsgWrap().Build()
	sys.Call(context.TODO(), router.Target{ID: id, Ty: ty, Ev: "entity_test"}, msg)

	assert.Equal(t, msg.Res.Header.Custom["code"], "200")
	a, e := sys.FindActor(context.TODO(), id)
	assert.NoError(t, e, nil)

	userActor := a.(*mockUserActor)
	assert.Equal(t, userActor.entity.IsDirty(), true)

	userActor.entity.Sync(context.TODO())
	assert.Equal(t, userActor.entity.IsDirty(), false)

	time.Sleep(time.Second * 2)

}

func TestEntitySync(t *testing.T) {

}

func TestEntityStore(t *testing.T) {

}

func TestEntityDB(t *testing.T) {
	redis.BuildClientWithOption(redis.WithAddr("redis://127.0.0.1:6379/0"))
	mgo.Build(mgo.AppendConn(mgo.ConnInfo{
		Name: "braid-test",
		Addr: "mongodb://127.0.0.1:27017",
	}))

	mockactorid := "test.actor.1"

	warp1 := NewEntityWapper(mockactorid)
	warp1.Airship = &EntityAirshipModule{ID: mockactorid}
	warp1.Bag = &EntityBagModule{ID: mockactorid, Bag: make(map[string]*ItemList)}
	warp1.TimeInfo = &EntityTimeInfoModule{ID: mockactorid, CreateTime: time.Now().Unix()}
	warp1.User = &EntityUserModule{ID: mockactorid, Token: "111"}

	warp1.Bag.Bag["1001"] = &ItemList{
		Items: []*Item{
			{
				ID:     "1001",
				Num:    10,
				DictID: "1001",
			},
		},
	}

	mgo.Collection("braid-test", "entity_test").InsertOne(context.TODO(), warp1)

	warp2 := NewEntityWapper(mockactorid)
	err := warp2.Load(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, warp1.Airship.ID, warp2.Airship.ID)

	warp2.User.Token = "222"
	warp2.Sync(context.TODO())
	warp2.Store(context.TODO())

	//warp3 := NewEntityWapper(mockactorid)
	//err = warp3.Load(context.TODO())
	//assert.NoError(t, err)
	//assert.Equal(t, warp2.User.Token, warp3.User.Token)
}
