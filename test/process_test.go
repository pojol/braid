package nodeprocess

import (
	"braid/core/actor"
	"braid/core/cluster/node"
	"braid/def"
	"braid/router"
	"context"
	"testing"
	"time"

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
	actor.Init(
		actor.SystemService("", "001"),
		actor.SystemActorConstructor(
			[]actor.ActorConstructor{
				{Type: def.MockActorEntity, Constructor: func(p *actor.CreateActorParm) actor.IActor {
					return &userActorProxy{&actor.BaseActor{Id: "mockentity", Ty: def.MockActorEntity}}
				}},
				{Type: def.MockActorClac, Constructor: func(p *actor.CreateActorParm) actor.IActor {
					return &clacActorProxy{&actor.BaseActor{Id: "mockclac", Ty: def.MockActorClac}}
				}},
			},
		),
	)

	_, err = actor.Regist(def.MockActorClac, actor.CreateActorWithID("mockclac"))
	if err != nil {
		panic(err) // 创建非法的 actor
	}

	_, err = actor.Regist(def.MockActorEntity, actor.CreateActorWithID("mockentity"))
	if err != nil {
		panic(err)
	}

	err = app.Init()
	assert.Equal(t, err, nil)

	app.Update()
	actor.Call(context.TODO(), router.Target{
		ID: "mockclac",
		Ty: def.MockActorClac,
		Ev: "clacA",
	}, &router.MsgWrapper{Req: &router.Message{Header: &router.Header{}}})

	app.WaitClose()

	time.Sleep(time.Second * 2)
}

func BenchmarkCall(b *testing.B) {

}
