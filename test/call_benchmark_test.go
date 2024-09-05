package nodeprocess

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/core/workerthread"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router"
	"golang.org/x/exp/rand"
)

// 2节点	1w个actor	2跳
// 4节点	2w个actor	2跳
// 6节点	4w个actor	2跳
// 8节点	8w个actor	2跳
// 16节点	16w个actor	2跳

// 2节点	1w个actor	4跳
// 4节点	2w个actor	4跳
// 6节点	4w个actor	4跳
// 8节点	8w个actor	4跳
// 16节点	16w个actor	4跳

func TestMain(m *testing.M) {

}

func Benchmark2Node1wActor2Jump(b *testing.B) {

	var nodes []*ProcessNode
	NodeNum := 2
	ActorNum := 10000
	JumpNum := 2

	actorjump1arr := []string{}

	for i := 0; i < NodeNum; i++ {

		sys := workerthread.BuildSystemWithOption(
			workerthread.SystemActorConstructor(
				[]workerthread.ActorConstructor{
					{Type: def.MockActorEntity, Constructor: func(p *workerthread.CreateActorParm) workerthread.IActor {
						return &mockEntityActor{
							&workerthread.BaseActor{Id: "mockentity", Ty: def.MockActorEntity, Sys: p.Sys},
						}
					}},
				},
			),
			workerthread.SystemService("service_"+strconv.Itoa(i), "node_"+strconv.Itoa(i)),
			workerthread.SystemWithAcceptor(1000+i),
		)

		node := &ProcessNode{
			p:   node.Parm{ID: strconv.Itoa(i)},
			sys: sys,
		}

		for k := 0; k < JumpNum; k++ {
			// register jump
			for j := 0; j < ActorNum/JumpNum; j++ {
				uid := uuid.NewString()
				aid := "actor_" + strconv.Itoa(k) + "_" + uid

				if k == 0 { // jump 1
					actorjump1arr = append(actorjump1arr, aid)
				}

				sys.Register(context.TODO(),
					def.MockActorEntity,
					workerthread.CreateActorWithID(aid))
			}
		}

		node.Init()
		node.Update()
		defer node.WaitClose()

		nodes = append(nodes, node)
	}

	rand.Seed(uint64(time.Now().UnixNano()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		randomIndex := rand.Intn(len(actorjump1arr))

		// 随机找一个第一跳的 actor 发送 event
		nodes[0].System().Call(context.TODO(),
			router.Target{ID: actorjump1arr[randomIndex], Ty: def.MockActorEntity},
			&router.MsgWrapper{Req: &router.Message{Header: &router.Header{}}})
	}

}

//////////////////////  SymbolWildcard //////////////////////
