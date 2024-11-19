package testreentercall

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/router"
	"github.com/pojol/braid/router/msg"
	"github.com/pojol/braid/test/mockdata"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	slog, _ := log.NewServerLogger("test")
	log.SetSLog(slog)

	defer log.Sync()

	os.Exit(m.Run())
}

func TestReenterCall(t *testing.T) {
	// use mock redis
	redis.BuildClientWithOption(redis.WithAddr("redis://127.0.0.1:6379/0"))
	redis.FlushAll(context.TODO()) // clean cache

	factory := mockdata.BuildActorFactory()
	loader := mockdata.BuildDefaultActorLoader(factory)

	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-reenter-call-1"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	// build
	var err error
	_, err = nod.System().Loader("MockClacActor").WithID("clac-1").Register()
	assert.Equal(t, err, nil)
	_, err = nod.System().Loader("MockClacActor").WithID("clac-2").Register()
	assert.Equal(t, err, nil)

	nod.Init()
	nod.Update()

	time.Sleep(time.Second)

	err = nod.System().Call(router.Target{ID: "clac-1", Ty: "MockClacActor", Ev: "mockreenter"}, msg.NewBuilder(context.TODO()).Build())
	assert.Equal(t, err, nil)

	time.Sleep(time.Second * 2)

	wg := sync.WaitGroup{}
	nod.System().Exit(&wg)
	wg.Wait()
}
