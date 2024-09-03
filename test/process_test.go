package nodeprocess

import (
	"context"
	"testing"
	"time"

	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/core/workerthread"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router"

	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	var err error

	app := &ProcessNode{
		p: node.Parm{
			ID:   "001",
			Name: "test",
		},
	}
	workerthread.Init(
		workerthread.SystemService("", "001"),
		workerthread.SystemActorConstructor(
			[]workerthread.ActorConstructor{
				{Type: def.MockActorEntity, Constructor: func(p *workerthread.CreateActorParm) workerthread.IActor {
					return &userActorProxy{&workerthread.BaseActor{Id: "mockentity", Ty: def.MockActorEntity}}
				}},
				{Type: def.MockActorClac, Constructor: func(p *workerthread.CreateActorParm) workerthread.IActor {
					return &clacActorProxy{&workerthread.BaseActor{Id: "mockclac", Ty: def.MockActorClac}}
				}},
			},
		),
	)

	_, err = workerthread.Regist(def.MockActorClac, workerthread.CreateActorWithID("mockclac"))
	if err != nil {
		panic(err) // 创建非法的 actor
	}

	_, err = workerthread.Regist(def.MockActorEntity, workerthread.CreateActorWithID("mockentity"))
	if err != nil {
		panic(err)
	}

	err = app.Init()
	assert.Equal(t, err, nil)

	app.Update()
	workerthread.Call(context.TODO(), router.Target{
		ID: "mockclac",
		Ty: def.MockActorClac,
		Ev: "clacA",
	}, &router.MsgWrapper{Req: &router.Message{Header: &router.Header{}}})

	app.WaitClose()

	time.Sleep(time.Second * 2)
}

func BenchmarkCall(b *testing.B) {

}
