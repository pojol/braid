package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/node"
	"github.com/pojol/braid/router/msg"
	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-send-1"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	// build
	var err error
	_, err = nod.System().Loader("mockb").WithID("mockb").Register(context.TODO())
	assert.Equal(t, err, nil)

	nod.Init()
	defer func() {
		wg := sync.WaitGroup{}
		nod.System().Exit(&wg)
		wg.Wait()
	}()

	t.Run("normal", func(t *testing.T) {
		m := msg.NewBuilder(context.TODO()).Build()
		timenow := time.Now()
		err := nod.System().Send("mockb", "mockb", "timeout", m)

		assert.Equal(t, true, time.Since(timenow) < time.Second) // 添加时间差检查
		assert.Equal(t, err, nil)
	})
}
