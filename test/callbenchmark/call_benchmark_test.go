package callbenchmark

import (
	"context"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/core/workerthread"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router"
	"github.com/pojol/braid/test"
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

var (
	nodes         []*test.ProcessNode
	actorjump1arr []string
	actorjump2arr []string
	initOnce      sync.Once
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func setupBenchmark() {
	NodeNum := 2
	ActorNum := 10
	JumpNum := 2

	redis.BuildClientWithOption(redis.WithAddr("redis://127.0.0.1:6379/0"))

	for i := 0; i < NodeNum; i++ {
		port, err := getFreePort()
		if err != nil {
			panic(err)
		}

		sys := workerthread.BuildSystemWithOption(
			workerthread.SystemActorConstructor(
				[]workerthread.ActorConstructor{
					{Type: def.MockActorEntity, Constructor: func(p *workerthread.CreateActorParm) workerthread.IActor {
						return &mockEntityActor{
							&workerthread.BaseActor{Ty: def.MockActorEntity, Sys: p.Sys},
						}
					}},
				},
			),
			workerthread.SystemService("service_"+strconv.Itoa(i), "node_"+strconv.Itoa(i)),
			workerthread.SystemWithAcceptor(port),
		)

		node := &test.ProcessNode{
			P:   node.Parm{ID: strconv.Itoa(i)},
			Sys: sys,
		}

		for k := 0; k < JumpNum; k++ {
			for j := 0; j < ActorNum/JumpNum; j++ {
				uid := uuid.NewString()
				aid := "actor_j" + strconv.Itoa(k) + "_" + uid

				if k == 0 {
					actorjump1arr = append(actorjump1arr, aid)
				} else if k == 1 {
					actorjump2arr = append(actorjump2arr, aid)
				}

				sys.Register(context.TODO(), def.MockActorEntity, workerthread.CreateActorWithID(aid))
			}
		}

		node.Init()
		node.Update()

		nodes = append(nodes, node)
	}

	time.Sleep(time.Second * 2)
	rand.Seed(uint64(time.Now().UnixNano()))
}

func TestMain(m *testing.M) {
	setupBenchmark()
	m.Run()
	// 清理资源
	for _, node := range nodes {
		node.System().Exit()
	}
}

func Benchmark2Node1wActor2Jump(b *testing.B) {
	initOnce.Do(setupBenchmark)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		randomIndex := rand.Intn(len(actorjump1arr))

		nodes[0].System().Call(context.TODO(),
			router.Target{
				ID: actorjump1arr[randomIndex],
				Ty: def.MockActorEntity,
				Ev: "print",
			},
			&router.MsgWrapper{Req: &router.Message{Header: &router.Header{
				Custom: map[string]string{
					"next": actorjump2arr[randomIndex],
				},
			}}})
	}
}

//////////////////////  SymbolWildcard //////////////////////
