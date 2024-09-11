package discoverk8s

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/pojol/braid/3rd/k8s"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/router"
	"github.com/pojol/braid/utils/algorithms"
)

type k8sDiscover struct {
	discoverTicker *time.Ticker
	info           def.ServiceInfo
	parm           Parm

	cli *k8s.Client
	sys core.ISystem

	// service id : service nod
	nodemap map[string]*def.Node

	sync.Mutex
}

func (k *k8sDiscover) Init() error {
	return nil
}

func BuildWithOption(info def.ServiceInfo, cli *k8s.Client, opts ...Option) core.IDiscover {

	p := Parm{
		SyncServicesInterval: time.Second * 2,
		Namespace:            "default",
		Tag:                  "braid",
	}

	for _, opt := range opts {
		opt(&p)
	}

	return &k8sDiscover{
		info:    info,
		cli:     cli,
		parm:    p,
		nodemap: make(map[string]*def.Node),
	}

}

// 后面使用 k8s 自带的 watch 机制
func (k *k8sDiscover) discoverImpl() {

	k.Lock()
	defer k.Unlock()

	servicesnodes := make(map[string]bool)
	updateflag := false

	services, err := k.cli.ListServices(context.TODO(), k.parm.Namespace)
	if err != nil {
		log.Warn("[braid.discover] err %v", err.Error())
		return
	}

	for _, v := range services {
		if v.Info.Name == "" || len(v.Nodes) == 0 {
			log.Warn("[braid.discover] service %s has no node", v.Info.Name)
			continue
		}

		if !algorithms.ContainsInSlice(v.Tags, k.parm.Tag) {
			log.Info("[braid.discover] rule out with service tag %v, self tag %v", v.Tags, k.parm.Tag)
			continue
		}

		if v.Info.Name == k.info.Name {
			log.Info("[braid.discover] rule out with self")
			continue
		}

		if algorithms.ContainsInSlice(k.parm.Blacklist, v.Info.Name) {
			log.Info("[braid.discover] rule out with black list %v", v.Info.Name)
			continue // 排除黑名单节点
		}

		// 添加节点
		for _, nod := range v.Nodes {

			servicesnodes[nod.ID] = true

			if _, ok := k.nodemap[nod.ID]; !ok {

				sn := def.Node{
					Name:    v.Info.Name,
					ID:      nod.ID,
					Address: nod.Address + ":" + strconv.Itoa(k.parm.getPortWithServiceName(v.Info.Name)),
				}
				log.Info("[braid.discover] new service %s node %s addr %s", v.Info.Name, nod.ID, sn.Address)
				k.nodemap[nod.ID] = &sn

				k.sys.Call(context.TODO(), router.Target{}, router.NewMsg().Build())
				/*
					k.pubsub.GetTopic(meta.TopicDiscoverServiceUpdate).Pub(context.TODO(), meta.EncodeUpdateMsg(
						meta.TopicDiscoverServiceNodeAdd,
						sn,
					))
				*/

				updateflag = true

			}

		}
	}

	// 排除节点
	for nodek := range k.nodemap {

		if _, ok := servicesnodes[nodek]; !ok {
			log.Info("[braid.discover] remove service %s node %s", k.nodemap[nodek].Name, k.nodemap[nodek].ID)

			/*
				k.pubsub.GetTopic(meta.TopicDiscoverServiceUpdate).Pub(context.TODO(), meta.EncodeUpdateMsg(
					meta.TopicDiscoverServiceNodeRmv,
					*k.nodemap[nodek],
				))
			*/
			delete(k.nodemap, nodek)
			updateflag = true
		}

	}

	// 同步节点信息
	if updateflag {

	}

}

func (k *k8sDiscover) discover() {
	syncService := func() {
		defer func() {
			if err := recover(); err != nil {
				log.Info("[braid.discover] syncService err %v", err)
			}
		}()
		// todo ..
		k.discoverImpl()
	}

	k.discoverTicker = time.NewTicker(k.parm.SyncServicesInterval)

	k.discoverImpl()

	for {
		<-k.discoverTicker.C
		syncService()
	}
}

func (k *k8sDiscover) Run() {

	log.Info("[braid.discover] running ...")

	go func() {
		k.discover()
	}()
}

func (k *k8sDiscover) Close() {

}
