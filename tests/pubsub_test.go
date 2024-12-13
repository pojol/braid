package tests

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/node"
	"github.com/pojol/braid/tests/mock"
	"github.com/stretchr/testify/assert"
)

func TestPubsub(t *testing.T) {
	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-pubsub-1"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	// build
	var err error
	_, err = nod.System().Loader("mocka").WithID("mocka").Register(context.TODO())
	assert.Equal(t, err, nil)

	nod.Init()
	defer func() {
		wg := sync.WaitGroup{}
		nod.System().Exit(&wg)
		wg.Wait()
	}()

	t.Run("normal", func(t *testing.T) {
		time.Sleep(time.Second * 1)

		err = nod.System().Pub("mocka", "offline_msg", []byte("offline msg"))
		assert.Equal(t, err, nil)

		time.Sleep(time.Second * 1)
	})
}

// go test -benchmem -run=^$ -bench ^BenchmarkPubsub$ github.com/pojol/braid/tests -v -benchtime=10s
func BenchmarkPubsub(b *testing.B) {
	atomic.StoreInt64(&mock.ReceivedMessageCount, 0)

	nod := node.BuildProcessWithOption(
		core.NodeWithID("benchmark-pubsub-1"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	nod.System().Loader("mocka").WithID("mocka").Register(context.TODO())

	nod.Init()
	defer func() {
		wg := sync.WaitGroup{}
		nod.System().Exit(&wg)
		wg.Wait()
	}()

	time.Sleep(time.Second)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nod.System().Pub("mocka", "offline_msg", []byte("offline msg"))
	}

	// 等待一小段时间确保消息都被处理
	time.Sleep(time.Second)
	b.Logf("Total messages received: %d", atomic.LoadInt64(&mock.ReceivedMessageCount))

}
