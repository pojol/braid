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

func TestReenter(t *testing.T) {

	nod := node.BuildProcessWithOption(
		core.NodeWithID("test-reenter-1"),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	// build
	var err error
	_, err = nod.System().Loader(mockaName).WithID(mockaName).Register(context.TODO())
	assert.Equal(t, err, nil)
	_, err = nod.System().Loader(reenterActorName).WithID(reenterActorName).Register(context.TODO())
	assert.Equal(t, err, nil)

	nod.Init()
	defer func() {
		wg := sync.WaitGroup{}
		nod.System().Exit(&wg)
		wg.Wait()
	}()

	t.Run("Normal Case", func(t *testing.T) {
		calcValue = 0
		err := nod.System().Call(reenterActorName, reenterActorName, "reenter",
			msg.NewBuilder(context.TODO()).Build())
		assert.Nil(t, err)
		time.Sleep(time.Second)
		assert.Equal(t, int32(8), calcValue) // (2 + 2) * 2
	})

	/*
		t.Run("Timeout Case", func(t *testing.T) {
			calcValue = 0

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel() // 确保资源被释放

			err := nod.System().Call(reenterActorName, reenterActorName, "timeout", msg.NewBuilder(ctx).Build())
			assert.NotNil(t, err)                                        // 应该返回actor不存在错误
			assert.Contains(t, err.Error(), "context deadline exceeded") // 验证是否是超时错误

		})
	*/
}
