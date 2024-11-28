package tests

import (
	"context"
	"sync"
	"testing"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/node"
	"github.com/pojol/braid/router/msg"
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
	_, err = nod.System().Loader(mockaName).WithID(mockaName).Register(context.TODO())
	assert.Equal(t, err, nil)
	_, err = nod.System().Loader(mockbName).WithID(mockbName).Register(context.TODO())
	assert.Equal(t, err, nil)
	_, err = nod.System().Loader(reenterActorName).WithID(reenterActorName).Register(context.TODO())
	assert.Equal(t, err, nil)

	nod.Init()
	defer func() {
		wg := sync.WaitGroup{}
		nod.System().Exit(&wg)
		wg.Wait()
	}()

	t.Run("normal", func(t *testing.T) {
		m := msg.NewBuilder(context.TODO()).Build()
		err := nod.System().Call(mockbName, mockbName, "ping", m)
		assert.Equal(t, err, nil)

		resval := msg.GetResField[string](m, "pong")
		assert.Equal(t, resval, "pong")
	})
}
