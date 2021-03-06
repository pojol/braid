package linkerredis

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/pojol/braid-go/3rd/redis"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/modules/electorconsul"
	"github.com/pojol/braid-go/modules/moduleparm"
	"github.com/pojol/braid-go/modules/pubsubnsq"
	"github.com/pojol/braid-go/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	mock.Init()

	m.Run()
}

func TestLinkerTarget(t *testing.T) {
	var tmu sync.Mutex
	tmu.Lock()
	// 用于生成测试用例使用的key前缀
	LinkerRedisPrefix = "TestLinkerTarget-"
	tmu.Unlock()

	log := module.GetBuilder(zaplogger.Name).Build("TestLinkerTarget").(logger.ILogger)

	mbb := module.GetBuilder(pubsubnsq.Name)
	mbb.AddModuleOption(pubsubnsq.WithLookupAddr([]string{}))
	mbb.AddModuleOption(pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("TestLinkerTarget", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	eb := module.GetBuilder(electorconsul.Name)
	eb.AddModuleOption(electorconsul.WithConsulAddr(mock.ConsulAddr))
	e := eb.Build("TestLinkerTarget",
		moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb)).(elector.IElector)
	defer e.Close()

	b := module.GetBuilder(Name)
	b.AddModuleOption(WithRedisAddr(mock.RedisAddr))
	b.AddModuleOption(WithRedisMaxIdle(8))
	b.AddModuleOption(WithRedisMaxActive(16))
	b.AddModuleOption(WithSyncTick(100))

	lc := b.Build("gate", moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb)).(linkcache.ILinkCache)

	// clean
	rclient := redis.New()
	rclient.Init(redis.Config{
		Address:        mock.RedisAddr,
		ReadTimeOut:    5 * time.Second,
		WriteTimeOut:   5 * time.Second,
		ConnectTimeOut: 2 * time.Second,
		MaxIdle:        16,
		MaxActive:      128,
		IdleTimeout:    0,
	})
	rclient.Del(LinkerRedisPrefix + "*")

	lc.Init()
	lc.Run()
	defer lc.Close()

	mb.GetTopic(elector.ChangeState).Pub(elector.EncodeStateChangeMsg(elector.EMaster))

	nods := []discover.Node{
		{
			ID:      "a001",
			Name:    "base",
			Address: "127.0.0.1:12001",
		},
		{
			ID:      "a002",
			Name:    "login",
			Address: "127.0.0.1:13001",
		},
	}

	err := lc.Link("token01", nods[0])
	assert.Equal(t, err, nil)

	err = lc.Link("token01", nods[1])
	assert.Equal(t, err, nil)

	err = lc.Link("token02", nods[0])
	assert.Equal(t, err, nil)

	addr, err := lc.Target("token01", "base")
	assert.Equal(t, err, nil)
	assert.Equal(t, addr, "127.0.0.1:12001")

	_, err = lc.Target("unknowtoken", "base")
	assert.NotEqual(t, err, nil)

	mb.GetTopic(linkcache.TokenUnlink).Pub(&pubsub.Message{Body: []byte("token01")})
	mb.GetTopic(linkcache.TokenUnlink).Pub(&pubsub.Message{Body: []byte("token02")})

	time.Sleep(time.Millisecond * 500)

	for _, v := range nods {
		mb.GetTopic(discover.ServiceUpdate).Pub(discover.EncodeUpdateMsg(discover.EventRemoveService, v))
	}

	time.Sleep(time.Millisecond * 100)
}

func TestLocalTarget(t *testing.T) {
	var tmu sync.Mutex
	tmu.Lock()
	// 用于生成测试用例使用的key前缀
	LinkerRedisPrefix = "TestLocalTarget-"
	tmu.Unlock()

	log := module.GetBuilder(zaplogger.Name).Build("TestLocalTarget").(logger.ILogger)

	mbb := module.GetBuilder(pubsubnsq.Name)
	mbb.AddModuleOption(pubsubnsq.WithLookupAddr([]string{}))
	mbb.AddModuleOption(pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("TestLocalTarget", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	eb := module.GetBuilder(electorconsul.Name)
	eb.AddModuleOption(electorconsul.WithConsulAddr(mock.ConsulAddr))
	e := eb.Build("TestLocalTarget",
		moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb)).(elector.IElector)
	defer e.Close()

	b := module.GetBuilder(Name)
	b.AddModuleOption(WithRedisAddr(mock.RedisAddr))
	b.AddModuleOption(WithMode(LinkerRedisModeLocal))

	lc := b.Build("localgate", moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb)).(linkcache.ILinkCache)

	// clean
	rclient := redis.New()
	rclient.Init(redis.Config{
		Address:        mock.RedisAddr,
		ReadTimeOut:    5 * time.Second,
		WriteTimeOut:   5 * time.Second,
		ConnectTimeOut: 2 * time.Second,
		MaxIdle:        16,
		MaxActive:      128,
		IdleTimeout:    0,
	})
	rclient.Del(LinkerRedisPrefix + "*")

	lc.Init()
	lc.Run()
	defer lc.Close()

	mb.GetTopic(elector.ChangeState).Pub(elector.EncodeStateChangeMsg(elector.EMaster))

	nods := []discover.Node{
		{
			ID:      "local001",
			Name:    "localbase",
			Address: "127.0.0.1:12001",
		},
		{
			ID:      "local002",
			Name:    "locallogin",
			Address: "127.0.0.1:13001",
		},
	}

	err := lc.Link("localtoken01", nods[0])
	assert.Equal(t, err, nil)

	err = lc.Link("localtoken01", nods[1])
	assert.Equal(t, err, nil)

	err = lc.Link("localtoken02", nods[0])
	assert.Equal(t, err, nil)

	addr, err := lc.Target("localtoken01", "localbase")
	assert.Equal(t, err, nil)
	assert.Equal(t, addr, "127.0.0.1:12001")

	_, err = lc.Target("unknowtoken", "localbase")
	assert.NotEqual(t, err, nil)

	lc.Unlink("localtoken01")

	for _, v := range nods {
		lc.Down(v)
	}

	time.Sleep(time.Millisecond * 500)
}

func TestDown(t *testing.T) {
	LinkerRedisPrefix = "TestLinkerRedisDown-"

}

func BenchmarkLink(b *testing.B) {
	LinkerRedisPrefix = "benchmarklink"

	log := module.GetBuilder(zaplogger.Name).Build("BenchmarkLink").(logger.ILogger)

	mbb := module.GetBuilder(pubsubnsq.Name)
	mbb.AddModuleOption(pubsubnsq.WithLookupAddr([]string{}))
	mbb.AddModuleOption(pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("BenchmarkLink", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	eb := module.GetBuilder(electorconsul.Name)
	eb.AddModuleOption(electorconsul.WithConsulAddr(mock.ConsulAddr))
	e := eb.Build("TestLinkerTarget",
		moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb)).(elector.IElector)
	defer e.Close()

	lb := module.GetBuilder(Name)
	lb.AddModuleOption(WithRedisAddr(mock.RedisAddr))

	lc := lb.Build("gate", moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb)).(linkcache.ILinkCache)
	rand.Seed(time.Now().UnixNano())

	lc.Init()
	lc.Run()
	defer lc.Close()

	baseTargets := []discover.Node{
		{
			ID:      "a001",
			Name:    "base",
			Address: "127.0.0.1:12001",
		},
		{
			ID:      "a002",
			Name:    "base",
			Address: "127.0.0.1:12002",
		},
		{
			ID:      "a003",
			Name:    "base",
			Address: "127.0.0.1:12003",
		},
	}

	loginTargets := []discover.Node{
		{
			ID:      "b001",
			Name:    "login",
			Address: "127.0.0.1:13001",
		},
		{
			ID:      "b002",
			Name:    "login",
			Address: "127.0.0.1:13001",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Link("token"+strconv.Itoa(i), baseTargets[rand.Intn(len(baseTargets))])
		lc.Link("token"+strconv.Itoa(i), loginTargets[rand.Intn(len(loginTargets))])
	}
}
