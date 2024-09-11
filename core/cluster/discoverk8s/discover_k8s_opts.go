package discoverk8s

import "time"

type ServicePortPair struct {
	Name string
	Port int
}

type Parm struct {
	// 同步节点信息间隔
	SyncServicesInterval time.Duration

	Namespace string
	Tag       string

	ServicePortLst []ServicePortPair

	Blacklist []string
}

// Option consul discover config wrapper
type Option func(*Parm)

// WithSyncServiceInterval 修改config中的interval
func WithSyncServiceInterval(interval time.Duration) Option {
	return func(c *Parm) {
		c.SyncServicesInterval = interval
	}
}

func WithNamespace(ns string) Option {
	return func(c *Parm) {
		c.Namespace = ns
	}
}

// 用于描述服务开放的端口号
func WithServicePortPairs(pairs []ServicePortPair) Option {
	return func(c *Parm) {
		c.ServicePortLst = pairs
	}
}

func (p *Parm) getPortWithServiceName(name string) int {
	for _, v := range p.ServicePortLst {
		if v.Name == name {
			return v.Port
		}
	}
	return 0
}

func WithSelectorTag(tag string) Option {
	return func(c *Parm) {
		c.Tag = tag
	}
}

func WithBlacklist(lst []string) Option {
	return func(c *Parm) {
		c.Blacklist = lst
	}
}
