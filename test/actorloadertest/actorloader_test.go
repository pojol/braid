package actorloader

import (
	"os"
	"testing"

	"github.com/pojol/braid/3rd/mgo"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/test"
)

func TestMain(m *testing.M) {
	slog, _ := log.NewServerLogger("test")
	log.SetSLog(slog)

	defer log.Sync()

	os.Exit(m.Run())
}

func ActorLoaderTest(t *testing.T) {
	// use mock redis
	redis.BuildClientWithOption(redis.WithAddr("redis://127.0.0.1:6379/0"))
	mgo.Build(mgo.AppendConn(mgo.ConnInfo{
		Name: "braid-test",
		Addr: "mongodb://127.0.0.1:27017",
	}))

	sys := node.BuildSystemWithOption()

	node := &test.ProcessNode{
		P:           core.NodeParm{ID: "test-actor-loader-1"},
		Sys:         sys,
		ActorLoader: actor.BuildDefaultActorLoader(sys, test.BuildActorFactory()),
	}

	node.Loader().Pick("UserActor").WithID("001").Register()
}
