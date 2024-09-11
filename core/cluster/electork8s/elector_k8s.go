package electork8s

import (
	"context"
	"time"

	"github.com/pojol/braid/3rd/k8s"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/log"
)

type k8selector struct {
	info   def.ServiceInfo
	p      Parm
	locked bool

	watchTicker   *time.Ticker
	refreshTicker *time.Ticker

	cli *k8s.Client
}

func BuildWithOption(info def.ServiceInfo, k8scli *k8s.Client, opts ...Option) core.IElector {

	p := Parm{
		WatchTick:   time.Second * 5,
		RefreshTick: time.Second * 10,
		Namespace:   "default",
		Name:        "leader-election",
	}

	for _, opt := range opts {
		opt(&p)
	}

	p.Name = info.Name + "-" + p.Name

	return &k8selector{
		p:    p,
		info: info,
		cli:  k8scli,
	}

}

func (e *k8selector) Init() error {

	// 尝试创建资源
	identity, err := e.cli.CreateLeases(context.TODO(), e.p.Namespace, e.p.Name, e.info.ID)
	errmsg := ""
	if err != nil {
		errmsg = err.Error()
	}

	log.Info("[braid.elector] create leases %s %s %s %s %s", e.p.Namespace, e.p.Name, e.info.ID, identity, errmsg)

	return nil
}

// 监听&获取资源
func (e *k8selector) watch() {
	watchLock := func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error("[braid.elector] watchLock err %v", err)
			}
		}()

		if !e.locked {
			tag, err := e.cli.GetLeases(context.TODO(), e.p.Namespace, e.p.Name)
			if err != nil {
				log.Warn("[braid.elector] acquire lock service %s err %v", e.p.Name, err.Error())
			}

			if tag == e.info.ID {
				e.locked = true
				/*
					e.ps.GetTopic(meta.TopicElectionChangeState).Pub(context.TODO(),
						meta.EncodeStateChangeMsg(meta.EMaster, e.info.ID))
				*/
				log.Info("[braid.elector] acquire lock service %s", e.p.Name)
			} else {
				/*
					e.ps.GetTopic(meta.TopicElectionChangeState).Pub(context.TODO(),
						meta.EncodeStateChangeMsg(meta.ESlave, e.info.ID))
				*/
			}
		}
	}

	// time.Millisecond * 2000
	e.watchTicker = time.NewTicker(e.p.WatchTick)

	for {
		<-e.watchTicker.C
		watchLock()
	}
}

// 续租
func (e *k8selector) refresh() {
	refushSession := func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error("[braid.elector] refresh err %v", err)
			}
		}()

		err := e.cli.RenewLeases(context.TODO(), e.p.Namespace, e.p.Name)
		if err != nil {
			log.Warn("[braid.elector] refresh session err %v", err.Error())
		}
	}

	// time.Millisecond * 1000 * 5
	e.refreshTicker = time.NewTicker(e.p.RefreshTick)

	for {
		<-e.refreshTicker.C

		if e.locked {
			refushSession()
		}
	}
}

func (e *k8selector) Run() {
	go func() {
		e.refresh()
	}()

	go func() {
		e.watch()
	}()
}

func (e *k8selector) Close() {
	err := e.cli.RmvLeases(context.TODO(), e.p.Namespace, e.p.Name)
	if err != nil {
		log.Warn("[braid.elector] remove leases err %s", err.Error())
	}
}
