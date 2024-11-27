package testentity

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pojol/braid/3rd/mgo"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/node"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/router/msg"
	"github.com/pojol/braid/test/mockdata"
	"github.com/pojol/braid/test/mockdata/mockactors"
	"github.com/pojol/braid/test/mockdata/mockentity"
	"github.com/stretchr/testify/assert"
	"github.com/tryvium-travels/memongo"
)

func TestMain(m *testing.M) {
	slog, _ := log.NewServerLogger("test")
	log.SetSLog(slog)

	mongoServer, err := memongo.Start("4.0.5") // 指定 MongoDB 版本
	if err != nil {
		panic(err)
	}
	defer mongoServer.Stop()

	// 获取连接地址
	mongoURI := mongoServer.URIWithRandomDB()

	// 初始化你的 MongoDB 客户端
	mgo.Build(mgo.AppendConn(mgo.ConnInfo{
		Name: "braid-test",
		Addr: mongoURI,
	}))

	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer mr.Close()
	redis.BuildClientWithOption(redis.WithAddr(fmt.Sprintf("redis://%s", mr.Addr())))

	defer log.Sync()

	os.Exit(m.Run())
}

func mockEntity2DB(id string) {
	warp1 := mockentity.NewEntityWapper(id)
	warp1.Airship = &mockentity.EntityAirshipModule{
		ID: id,
	}
	warp1.Bag = &mockentity.EntityBagModule{
		ID:  id,
		Bag: make(map[string]*mockentity.ItemList),
	}
	warp1.TimeInfo = &mockentity.EntityTimeInfoModule{
		ID:         id,
		CreateTime: time.Now().Unix(),
	}
	warp1.User = &mockentity.EntityUserModule{
		ID:    id,
		Token: "111",
	}

	warp1.Bag.Bag["1001"] = &mockentity.ItemList{
		Items: []*mockentity.Item{
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

	factory := mockdata.BuildActorFactory()
	loader := mockdata.BuildDefaultActorLoader(factory)

	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-mock-entity"),
		core.NodeWithFactory(factory),
		core.NodeWithLoader(loader),
	)

	loader.Builder("MockUserActor", nod.System()).WithID(id).Register(context.TODO())

	for _, a := range nod.System().Actors() {
		a.Init(context.TODO())
	}

	return nod.System()
}

func TestEntityLoad(t *testing.T) {

	id := "test.actor.1"
	ty := "mockUserActor"

	mockEntity2DB(id)
	//////////////////////////////////////////

	// load entity with db and sync to redis
	sys := mockEntity(id)

	m := msg.NewBuilder(context.TODO()).Build()
	sys.Call(id, ty, "entity_test", m)

	assert.Equal(t, m.Err, nil)
	assert.Equal(t, msg.GetResField[string](m, "code"), "200")
	a, e := sys.FindActor(context.TODO(), id)
	assert.NoError(t, e, nil)

	userActor := a.(*mockactors.MockUserActor)
	assert.Equal(t, userActor.State.IsDirty(), true)

	userActor.State.Sync(context.TODO(), false)
	assert.Equal(t, userActor.State.IsDirty(), false)

	time.Sleep(time.Second * 2)

}

func TestEntitySync(t *testing.T) {

}

func TestEntityStore(t *testing.T) {

}

func TestEntityDB(t *testing.T) {
	mockactorid := "test.actor.1"

	warp1 := mockentity.NewEntityWapper(mockactorid)
	warp1.Airship = &mockentity.EntityAirshipModule{ID: mockactorid}
	warp1.Bag = &mockentity.EntityBagModule{ID: mockactorid, Bag: make(map[string]*mockentity.ItemList)}
	warp1.TimeInfo = &mockentity.EntityTimeInfoModule{ID: mockactorid, CreateTime: time.Now().Unix()}
	warp1.User = &mockentity.EntityUserModule{ID: mockactorid, Token: "111"}

	warp1.Bag.Bag["1001"] = &mockentity.ItemList{
		Items: []*mockentity.Item{
			{
				ID:     "1001",
				Num:    10,
				DictID: "1001",
			},
		},
	}

	mgo.Collection("braid-test", "entity_test").InsertOne(context.TODO(), warp1)

	warp2 := mockentity.NewEntityWapper(mockactorid)
	err := warp2.Load(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, warp1.Airship.ID, warp2.Airship.ID)

	warp2.User.Token = "222"
	warp2.Sync(context.TODO(), false)
	warp2.Store(context.TODO())

	//warp3 := NewEntityWapper(mockactorid)
	//err = warp3.Load(context.TODO())
	//assert.NoError(t, err)
	//assert.Equal(t, warp2.User.Token, warp3.User.Token)
}
