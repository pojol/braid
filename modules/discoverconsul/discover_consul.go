// 实现文件 基于 consul 实现的服务发现
package discoverconsul

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/pojol/braid-go/3rd/consul"
	"github.com/pojol/braid-go/internal/utils"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/modules/moduleparm"
)

const (
	// Name 发现器名称
	Name = "ConsulDiscover"

	// DiscoverTag 用于docker发现的tag， 所有希望被discover服务发现的节点，
	// 都应该在Dockerfile中设置 ENV SERVICE_TAGS=braid
	DiscoverTag = "braid"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")

	// 权重预设值，可以约等于节点支持的最大连接数
	// 在开启linker的情况下，节点的连接数越多权重值就越低，直到降到最低的 1权重
	defaultWeight = 1024
)

type consulDiscoverBuilder struct {
	opts []interface{}
}

func newConsulDiscover() module.IBuilder {
	return &consulDiscoverBuilder{}
}

func (b *consulDiscoverBuilder) Name() string {
	return Name
}

func (b *consulDiscoverBuilder) Type() module.ModuleType {
	return module.Discover
}

func (b *consulDiscoverBuilder) AddModuleOption(opt interface{}) {
	b.opts = append(b.opts, opt)
}

func (b *consulDiscoverBuilder) Build(name string, buildOpts ...interface{}) interface{} {

	bp := moduleparm.BuildParm{}
	for _, opt := range buildOpts {
		opt.(moduleparm.Option)(&bp)
	}

	p := Parm{
		Tag:                       "braid",
		Name:                      name,
		SyncServicesInterval:      time.Second * 2,
		SyncServiceWeightInterval: time.Second * 10,
		Address:                   "http://127.0.0.1:8500",
	}

	for _, opt := range b.opts {
		opt.(Option)(&p)
	}

	e := &consulDiscover{
		parm:       p,
		ps:         bp.PS,
		logger:     bp.Logger,
		passingMap: make(map[string]*syncNode),
	}

	e.ps.RegistTopic(discover.ServiceUpdate, pubsub.ScopeProc)

	return e
}

func (dc *consulDiscover) Init() error {

	// check address
	_, err := consul.GetConsulLeader(dc.parm.Address)
	if err != nil {
		return fmt.Errorf("%v Dependency check error %v [%v]", dc.parm.Name, "consul", dc.parm.Address)
	}

	ip, err := utils.GetLocalIP()
	if err != nil {
		return fmt.Errorf("%v GetLocalIP err %v", dc.parm.Name, err.Error())
	}

	linkC := dc.ps.GetTopic(linkcache.ServiceLinkNum).Sub(Name + "-" + ip)
	linkC.Arrived(func(msg *pubsub.Message) {
		lninfo := linkcache.DecodeLinkNumMsg(msg)
		dc.lock.Lock()
		defer dc.lock.Unlock()

		if _, ok := dc.passingMap[lninfo.ID]; ok {
			dc.passingMap[lninfo.ID].linknum = lninfo.Num
		}
	})

	return nil
}

// Discover 发现管理braid相关的节点
type consulDiscover struct {
	discoverTicker   *time.Ticker
	syncWeightTicker *time.Ticker

	// parm
	parm   Parm
	ps     pubsub.IPubsub
	logger logger.ILogger

	// service id : service nod
	passingMap map[string]*syncNode

	lock sync.Mutex
}

type syncNode struct {
	service string
	id      string
	address string

	linknum int

	dyncWeight int
	physWeight int
}

func (dc *consulDiscover) InBlacklist(name string) bool {

	for _, v := range dc.parm.Blacklist {
		if v == name {
			return true
		}
	}

	return false
}

func (dc *consulDiscover) discoverImpl() {

	dc.lock.Lock()
	defer dc.lock.Unlock()

	services, err := consul.GetCatalogServices(dc.parm.Address, dc.parm.Tag)
	if err != nil {
		return
	}

	for _, service := range services {
		if service.ServiceName == dc.parm.Name {
			continue
		}

		if dc.InBlacklist(service.ServiceName) {
			continue
		}

		if service.ServiceName == "" || service.ServiceID == "" {
			continue
		}

		if _, ok := dc.passingMap[service.ServiceID]; !ok { // new nod
			sn := syncNode{
				service:    service.ServiceName,
				id:         service.ServiceID,
				address:    service.ServiceAddress + ":" + strconv.Itoa(service.ServicePort),
				dyncWeight: 0,
				physWeight: defaultWeight,
			}
			dc.logger.Infof("new service %s addr %s", service.ServiceName, sn.address)
			dc.passingMap[service.ServiceID] = &sn

			dc.ps.GetTopic(discover.ServiceUpdate).Pub(discover.EncodeUpdateMsg(
				discover.EventAddService,
				discover.Node{
					ID:      sn.id,
					Name:    sn.service,
					Address: sn.address,
					Weight:  sn.physWeight,
				},
			))
		}
	}

	for k := range dc.passingMap {
		if _, ok := services[k]; !ok { // rmv nod
			dc.logger.Infof("remove service %s id %s", dc.passingMap[k].service, dc.passingMap[k].id)

			dc.ps.GetTopic(discover.ServiceUpdate).Pub(discover.EncodeUpdateMsg(
				discover.EventRemoveService,
				discover.Node{
					ID:      dc.passingMap[k].id,
					Name:    dc.passingMap[k].service,
					Address: dc.passingMap[k].address,
				},
			))

			delete(dc.passingMap, k)
		}
	}
}

func (dc *consulDiscover) syncWeight() {
	dc.lock.Lock()
	defer dc.lock.Unlock()

	for k, v := range dc.passingMap {
		if v.linknum == 0 {
			continue
		}

		if v.linknum == v.dyncWeight {
			continue
		}

		dc.passingMap[k].dyncWeight = v.linknum
		nweight := 0
		if dc.passingMap[k].physWeight-v.linknum > 0 {
			nweight = dc.passingMap[k].physWeight - v.linknum
		} else {
			nweight = 1
		}

		dc.ps.GetTopic(discover.ServiceUpdate).Pub(discover.EncodeUpdateMsg(
			discover.EventUpdateService,
			discover.Node{
				ID:     v.id,
				Name:   v.service,
				Weight: nweight,
			},
		))
	}
}

func (dc *consulDiscover) discover() {
	syncService := func() {
		defer func() {
			if err := recover(); err != nil {
				dc.logger.Errorf("consul discover syncService err %v", err)
			}
		}()
		// todo ..
		dc.discoverImpl()
	}

	dc.discoverTicker = time.NewTicker(dc.parm.SyncServicesInterval)

	dc.discoverImpl()

	for {
		select {
		case <-dc.discoverTicker.C:
			syncService()
		}
	}
}

func (dc *consulDiscover) weight() {
	syncWeight := func() {
		defer func() {
			if err := recover(); err != nil {
				dc.logger.Errorf("consul discover syncWeight err %v", err)
			}
		}()

		dc.syncWeight()
	}

	dc.syncWeightTicker = time.NewTicker(dc.parm.SyncServiceWeightInterval)

	for {
		select {
		case <-dc.syncWeightTicker.C:
			syncWeight()
		}
	}
}

// Discover 运行管理器
func (dc *consulDiscover) Run() {
	go func() {
		dc.discover()
	}()

	go func() {
		dc.weight()
	}()
}

// Close close
func (dc *consulDiscover) Close() {

}

func init() {
	module.Register(newConsulDiscover())
}
