package tests

import (
	"context"
	"sync"
	"testing"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/node"
	"github.com/pojol/braid/router/msg"
	"github.com/pojol/braid/tests/mock"
	"github.com/stretchr/testify/assert"
)

func TestCall(t *testing.T) {
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
	_, err = nod.System().Loader("mockc").WithID("mockc").Register(context.TODO())
	assert.Equal(t, err, nil)

	nod.Init()
	defer func() {
		wg := sync.WaitGroup{}
		nod.System().Exit(&wg)
		wg.Wait()
	}()

	t.Run("normal", func(t *testing.T) {
		m := msg.NewBuilder(context.TODO()).Build()
		err := nod.System().Call("mockc", "mockc", "ping", m)
		assert.Equal(t, err, nil)

		resval := msg.GetResField[string](m, "pong")
		assert.Equal(t, resval, "pong")
	})
}

func TestTCCSucc(t *testing.T) {
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

	err = nod.System().Call("mocka", "mocka", "tcc_succ", msg.NewBuilder(context.TODO()).Build())
	assert.Nil(t, err)

	assert.Equal(t, mock.MockBTccValue, 111)
	assert.Equal(t, mock.MockCTccValue, 222)
}
