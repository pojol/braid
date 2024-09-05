package nodeprocess

import (
	"context"
	"strconv"
	"testing"

	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router"
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

	for i := 0; i < NodeNum; i++ {
		node := &ProcessNode{p: node.Parm{ID: strconv.Itoa(i), Name: "node_" + strconv.Itoa(i)}}

		sys := NewSystem(
			WithServiceInfo(),
			WithActorConstructor(),
		)

		for k := 0; k < JumpNum; k++ {
			// register jump1
			for j := 0; j < ActorNum/JumpNum; j++ {
				sys.Register(context.TODO(), def.MockActorClac, withid("actor_"+strconv.Itoa(k)+"_"+"uuid"))
			}
		}

		node := NewProcessNode(
			WithServiceInfo(),
			WithSystem(sys),
		)

		node.Init()
		node.Update()
		defer node.Exit()

		nodes = append(nodes, node)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 随机找一个第一跳的 actor 发送 event
		nodes[0].Call(context.TODO(), 
			router.Target{ID: "actor_0_"+map["randid"], Ty: def.MockActorClac}, 
			&router.MsgWrapper{Req: &router.Message{Header: &router.Header{}}})
	}

}

//////////////////////  SymbolWildcard //////////////////////
