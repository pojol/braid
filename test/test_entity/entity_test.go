package testentity

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
	"github.com/pojol/braid/test/mockdata"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	slog, _ := log.NewServerLogger("test")
	log.SetSLog(slog)

	defer log.Sync()

	os.Exit(m.Run())
}

func mockEntity2DB(id string) {
	warp1 := mockdata.NewEntityWapper(id)
	warp1.Airship = &mockdata.EntityAirshipModule{
		ID: id,
	}
	warp1.Bag = &mockdata.EntityBagModule{
		ID:  id,
		Bag: make(map[string]*mockdata.ItemList),
	}
	warp1.TimeInfo = &mockdata.EntityTimeInfoModule{
		ID:         id,
		CreateTime: time.Now().Unix(),
	}
	warp1.User = &mockdata.EntityUserModule{
		ID:    id,
		Token: "111",
	}

	warp1.Bag.Bag["1001"] = &mockdata.ItemList{
		Items: []*mockdata.Item{
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
		"test-mock-entity",
		mockdata.BuildActorFactory(),
	)

	loader := actor.BuildDefaultActorLoader(sys, mockdata.BuildActorFactory())
	loader.Builder("MockUserActor").WithID(id).Build()

	for _, a := range sys.Actors() {
		a.Init(context.TODO())
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

	msg := router.NewMsgWrap(context.TODO()).Build()
	sys.Call(router.Target{ID: id, Ty: ty, Ev: "entity_test"}, msg)

	assert.Equal(t, msg.Res.Header.Custom["code"], "200")
	a, e := sys.FindActor(context.TODO(), id)
	assert.NoError(t, e, nil)

	userActor := a.(*mockdata.MockUserActor)
	assert.Equal(t, userActor.State.IsDirty(), true)

	userActor.State.Sync(context.TODO())
	assert.Equal(t, userActor.State.IsDirty(), false)

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

	warp1 := mockdata.NewEntityWapper(mockactorid)
	warp1.Airship = &mockdata.EntityAirshipModule{ID: mockactorid}
	warp1.Bag = &mockdata.EntityBagModule{ID: mockactorid, Bag: make(map[string]*mockdata.ItemList)}
	warp1.TimeInfo = &mockdata.EntityTimeInfoModule{ID: mockactorid, CreateTime: time.Now().Unix()}
	warp1.User = &mockdata.EntityUserModule{ID: mockactorid, Token: "111"}

	warp1.Bag.Bag["1001"] = &mockdata.ItemList{
		Items: []*mockdata.Item{
			{
				ID:     "1001",
				Num:    10,
				DictID: "1001",
			},
		},
	}

	mgo.Collection("braid-test", "entity_test").InsertOne(context.TODO(), warp1)

	warp2 := mockdata.NewEntityWapper(mockactorid)
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
