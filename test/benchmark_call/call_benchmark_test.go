package benchmarkcall

import (
	"context"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/node"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/router/msg"
	"github.com/pojol/braid/test/mockdata"
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
	nodes         []core.INode
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

		factory := mockdata.BuildActorFactory()
		loader := mockdata.BuildDefaultActorLoader(factory)

		node := node.BuildProcessWithOption(
			core.NodeWithID(strconv.Itoa(i)),
			core.NodeWithLoader(loader),
			core.NodeWithFactory(factory),
			core.NodeWithPort(port),
		)

		for k := 0; k < JumpNum; k++ {
			for j := 0; j < ActorNum/JumpNum; j++ {
				uid := uuid.NewString()
				aid := "actor_j" + strconv.Itoa(k) + "_" + uid

				if k == 0 {
					actorjump1arr = append(actorjump1arr, aid)
				} else if k == 1 {
					actorjump2arr = append(actorjump2arr, aid)
				}

				node.System().Loader("MockClacActor").WithID(aid).Register(context.TODO())
			}
		}

		node.Init()

		nodes = append(nodes, node)
	}

	time.Sleep(time.Second * 2)
	rand.Seed(uint64(time.Now().UnixNano()))
}

func TestMain(m *testing.M) {
	setupBenchmark()
	m.Run()
	// 清理资源
	wg := sync.WaitGroup{}
	for _, node := range nodes {
		node.System().Exit(&wg)
	}
}

func Benchmark2Node1wActor2Jump(b *testing.B) {
	initOnce.Do(setupBenchmark)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		randomIndex := rand.Intn(len(actorjump1arr))

		m := msg.NewBuilder(context.TODO())
		m.WithReqCustomFields(msg.Attr{Key: "next", Value: actorjump2arr[randomIndex]})

		nodes[0].System().Call(
			actorjump1arr[randomIndex],
			def.MockActorEntity,
			"print", m.Build())
	}
}

//////////////////////  SymbolWildcard //////////////////////
