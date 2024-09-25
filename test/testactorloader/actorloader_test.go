package testactorloader

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/mgo"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/test/mockdata"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	slog, _ := log.NewServerLogger("test")
	log.SetSLog(slog)

	defer log.Sync()

	os.Exit(m.Run())
}

func TestActorLoader(t *testing.T) {
	// use mock redis
	redis.BuildClientWithOption(redis.WithAddr("redis://127.0.0.1:6379/0"))
	mgo.Build(mgo.AppendConn(mgo.ConnInfo{
		Name: "braid-test",
		Addr: "mongodb://127.0.0.1:27017",
	}))

	redis.FlushAll(context.TODO()) // clean cache

	sys := node.BuildSystemWithOption(mockdata.BuildActorFactory())

	node := &mockdata.ProcessNode{
		P:   core.NodeParm{ID: "test-actor-loader-1"},
		Sys: sys,
	}

	var err error

	_, err = sys.Loader().Builder(def.ActorDynamicPicker).WithID("nodeid-picker").RegisterLocally()
	assert.Equal(t, err, nil)

	_, err = sys.Loader().Builder(def.ActorDynamicRegister).WithID("nodeid-register").RegisterLocally()
	assert.Equal(t, err, nil)

	node.Init()
	node.Update()

	err = sys.Loader().Builder("MockClacActor").WithID("001").RegisterDynamically(context.Background())
	assert.Equal(t, err, nil)

	//node.WaitClose()
	select {
	case <-time.After(3 * time.Second):
		// 3 seconds have passed
	}
}
