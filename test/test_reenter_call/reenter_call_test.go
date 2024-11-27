package testreentercall

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/node"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/router/msg"
	"github.com/pojol/braid/test/mockdata"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	slog, _ := log.NewServerLogger("test")
	log.SetSLog(slog)

	defer log.Sync()

	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer mr.Close()
	redis.BuildClientWithOption(redis.WithAddr(fmt.Sprintf("redis://%s", mr.Addr())))

	os.Exit(m.Run())
}

func TestReenterCall(t *testing.T) {
	factory := mockdata.BuildActorFactory()
	loader := mockdata.BuildDefaultActorLoader(factory)

	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-reenter-call-1"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	// build
	var err error
	_, err = nod.System().Loader("MockClacActor").WithID("clac-1").Register(context.TODO())
	assert.Equal(t, err, nil)
	_, err = nod.System().Loader("MockClacActor").WithID("clac-2").Register(context.TODO())
	assert.Equal(t, err, nil)

	nod.Init()

	time.Sleep(time.Second)

	nod.System().Call("clac-1", "mockactor", "mockreenter", msg.NewBuilder(context.TODO()).Build())
	err = nod.System().Call("clac-1", "MockClacActor", "mockreenter", msg.NewBuilder(context.TODO()).Build())
	assert.Equal(t, err, nil)

	time.Sleep(time.Second * 2)

	wg := sync.WaitGroup{}
	nod.System().Exit(&wg)
	wg.Wait()
}
