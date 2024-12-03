package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/node"
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
