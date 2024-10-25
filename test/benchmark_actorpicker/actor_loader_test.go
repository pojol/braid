package benchmarkactorloader

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	trdredis "github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/cluster/node"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/test/mockdata"
	"github.com/redis/go-redis/v9"
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

func TestMain(m *testing.M) {
	// 设置代码，只运行一次
	slog, _ := log.NewServerLogger("BenchmarkPicker")
	log.SetSLog(slog)
	defer log.Sync()

	trdredis.BuildClientWithOption(
		trdredis.WithAddr("redis://127.0.0.1:6379/0"),
		trdredis.WithPoolSize(1024),
		trdredis.WithPoolTimeout(time.Second*10),
		trdredis.WithMinIdle(100),
	)

	// clean cache
	trdredis.FlushAll(context.Background())

	for i := 0; i < 20; i++ {
		i := i // 创建一个新的变量来捕获循环变量
		go func() {
			factory := mockdata.BuildActorFactory()
			loader := mockdata.BuildDefaultActorLoader(factory)

			nodid := "1000_" + strconv.Itoa(i)
			p, _ := getFreePort()

			nod := node.BuildProcessWithOption(
				core.NodeWithID(nodid),
				core.NodeWithWeight(10000),
				core.NodeWithLoader(loader),
				core.NodeWithFactory(factory),
				core.NodeWithPort(p),
			)

			err := nod.Init()
			if err != nil {
				panic(fmt.Errorf("node init err %v", err.Error()))
			}

			nod.Update()
		}()
	}
	time.Sleep(time.Second)

	m.Run()
	// 清理资源
}

func makeNodeKey(nodid string) string {
	return fmt.Sprintf("{node:%s}", nodid)
}

func printWeight() error {
	// Get all node infos from the set
	nodeInfoMap, err := trdredis.HGetAll(context.Background(), def.RedisAddressbookNodesField).Result()
	if err != nil {
		return fmt.Errorf("failed to get node infos: %v", err)
	}

	if len(nodeInfoMap) == 0 {
		return fmt.Errorf("no nodes found")
	}

	pipe := trdredis.Pipeline()

	// Prepare pipeline commands to get weights for all nodes
	for nodeID := range nodeInfoMap {
		pipe.HGet(context.Background(), makeNodeKey(nodeID), "total_weight")
	}

	// Execute pipeline
	cmders, err := pipe.Exec(context.Background())
	if err != nil {
		return fmt.Errorf("pipeline execution failed: %v", err)
	}

	// Process results
	i := 0
	for nodeID, nodeInfoJSON := range nodeInfoMap {
		if i >= len(cmders) {
			break
		}

		var nodeInfo core.AddressInfo
		if err := json.Unmarshal([]byte(nodeInfoJSON), &nodeInfo); err != nil {
			log.WarnF("unable to unmarshal node info: %v", err)
			i++
			continue
		}

		weightStr, err := cmders[i].(*redis.StringCmd).Result()
		if err != nil {
			log.WarnF("unable to get weight for node %s: %v", nodeID, err)
			i++
			continue
		}

		weight, _ := strconv.Atoi(weightStr)
		fmt.Println("node", nodeInfo.Node, "cur weight", weight)

		i++
	}

	return nil
}

func TestPicker(b *testing.T) {

	factory := mockdata.BuildActorFactory()
	loader := mockdata.BuildDefaultActorLoader(factory)

	nodid := "1000_x"
	p, _ := getFreePort()

	nod := node.BuildProcessWithOption(
		core.NodeWithID(nodid),
		core.NodeWithWeight(10000),
		core.NodeWithLoader(loader),
		core.NodeWithPort(p),
	)

	err := nod.Init()
	if err != nil {
		panic(fmt.Errorf("node init err %v", err.Error()))
	}

	nod.Update()

	time.Sleep(time.Second)

	for i := 0; i < 5000; i++ {
		err = nod.System().Loader("MockClacActor").Picker()
		if err != nil {
			b.Logf("picker err %v", err.Error())
		}
	}

	time.Sleep(time.Second * 20)
	b.Logf("total register count %v", atomic.LoadInt32(&mockdata.GlobalCreateCnt))

	// 再看下分布情况
	printWeight()
}
