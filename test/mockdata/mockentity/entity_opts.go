package mockentity

import "github.com/pojol/braid/core"

const (
	EntityTimeoutField   = "dataPlane.entity.ttl"
	EntityDatabaseName   = "dptest"
	EntityCollectionName = "entity"
)

type EntityParam struct {

	// 记录 entity ttl的field， 默认为 TimeoutField 用户也可以自定义修改
	timeOutField string

	// 数据库名
	databaseName string

	// collection
	collectionName string

	// 缓存加载器
	cacheLoadStrategy core.ICacheStrategy

	// entity id 的 前缀id（自定义字段，有时候会在前缀上补一些渠道标识
	prefixID string

	// 同步间隔时间 (秒
	syncInterval int

	// 数据过期时间（过期后会同步给数据库）
	ttl int
}

type EntityOption func(*EntityParam)

func WithCacheLoadStrategy(cs core.ICacheStrategy) EntityOption {
	return func(param *EntityParam) {
		param.cacheLoadStrategy = cs
	}
}

func WithTimeoutField(field string) EntityOption {
	return func(param *EntityParam) {
		param.timeOutField = field
	}
}

func WithTTL(ttl int) EntityOption {
	return func(param *EntityParam) {
		param.ttl = ttl
	}
}

func WithDatabaseName(name string) EntityOption {
	return func(param *EntityParam) {
		param.databaseName = name
	}
}

func WithCollectionName(name string) EntityOption {
	return func(param *EntityParam) {
		param.collectionName = name
	}
}
