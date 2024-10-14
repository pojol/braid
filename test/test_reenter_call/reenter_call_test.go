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

	loader := mockdata.BuildDefaultActorLoader(mockdata.BuildActorFactory())

	sys := node.BuildSystemWithOption("test-reenter-call-1", loader)

	node := &mockdata.ProcessNode{
		P:   core.NodeParm{ID: "st-reenter-call-1"},
		Sys: sys,
	}

	// build
	var err error
	_, err = sys.Loader("MockClacActor").WithID("clac-1").Build()
	assert.Equal(t, err, nil)
	_, err = sys.Loader("MockClacActor").WithID("clac-2").Build()
	assert.Equal(t, err, nil)

	node.Init()
	node.Update()

	time.Sleep(time.Second)

	err = sys.Call(router.Target{ID: "clac-1", Ty: "MockClacActor", Ev: "mockreenter"}, router.NewMsgWrap(context.TODO()).Build())
	assert.Equal(t, err, nil)

	time.Sleep(time.Second * 2)

	wg := sync.WaitGroup{}
	sys.Exit(&wg)
	wg.Wait()
}
