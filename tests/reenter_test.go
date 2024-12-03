package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/node"
	"github.com/pojol/braid/router/msg"
	"github.com/pojol/braid/tests/mock"
	"github.com/stretchr/testify/assert"
)

func TestReenter(t *testing.T) {

	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-reenter-1"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	// build
	var err error
	_, err = nod.System().Loader("mocka").WithID("mocka").Register(context.TODO())
	assert.Equal(t, err, nil)
	_, err = nod.System().Loader("mockb").WithID("mockb").Register(context.TODO())
	assert.Equal(t, err, nil)

	nod.Init()
	defer func() {
		wg := sync.WaitGroup{}
		nod.System().Exit(&wg)
		wg.Wait()
	}()

	time.Sleep(time.Second)

	t.Run("Normal Case", func(t *testing.T) {
		mock.RecenterCalcValue = 0
		err := nod.System().Call("mocka", "mocka", "reenter",
			msg.NewBuilder(context.TODO()).Build())
		assert.Nil(t, err)
		time.Sleep(time.Second)
		assert.Equal(t, int32(8), mock.RecenterCalcValue) // (2 + 2) * 2
	})

	t.Run("Timeout Case", func(t *testing.T) {

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		m := msg.NewBuilder(ctx).Build()

		nod.System().Call("mocka", "mocka", "timeout", m)
		time.Sleep(time.Second * 4)
		assert.NotNil(t, m.Err) // 应该返回actor不存在错误
	})

	t.Run("Timeout chain", func(t *testing.T) {
		mock.RecenterCalcValue = 0

		err := nod.System().Call("mocka", "mocka", "chain", msg.NewBuilder(context.TODO()).Build())
		assert.Nil(t, err)

		assert.Nil(t, err)
		time.Sleep(time.Second)
		assert.Equal(t, int32(18), mock.RecenterCalcValue) // ((2 + 2) * 2) + 10
	})
}
