package core

import (
	"context"
	"errors"
	"reflect"
	"sync"
)

type EntityBlock string

type IBlock interface {
	// Unload - 卸载 将 entity 编码成字节流 bytes
	Unload() ([]byte, error)

	// Load - 装载 将 bytes 加载到本地并解析成 entity
	Load([]byte) error
}

// IEntity 实体接口
//
//	游戏角色实体，包含了角色的各种数据，他们以json的数据格式存储
//	因此，entity 不支持如 map，time.Time，pointer 等复杂类型
type IEntity interface {

	// GetID 获取 entity id
	GetID() string

	GetToken() string
	SetToken(token string)

	// GetDataPtr 获取 entity 数据指针
	GetDataPtr() interface{}

	// CheckVersion 检查 entity 的版本号，低版本的 entity 需要进行升级行为
	//  依赖分布式锁
	CheckVersion() error
}

type ICacheStrategy interface {
	// Load - 尝试从 cache 中拉取entity的数据
	Load(ctx context.Context) error

	// Unload - 将内存数据同步回cache
	Unload(ctx context.Context) error

	// Sync - 将内存数据同步回cache
	Sync(ctx context.Context) error

	// Clean - 将内存数据清除（请务必确保在保存完数据后进行
	Clean(ctx context.Context) error

	// GetModule - 获取entity的module
	GetModule(typ reflect.Type) interface{}

	// SetModule - 设置entity的module
	SetModule(typ reflect.Type, module interface{})
}

var verList = []VerStrategy{}
var verRMu sync.RWMutex

// RegisterVersionStrategy 注册entity版本变更执行的策略函数
func RegisterVersionStrategy(s VerStrategy) {
	verList = append(verList, s)
}

var ErrNotfoundVersion = errors.New("not found version")

func GetVerStrategy(ver int) (VerStrategy, error) {
	verRMu.RLock()
	defer verRMu.RUnlock()

	for _, v := range verList {
		if v.Version == ver {
			return v, nil
		}
	}
	return VerStrategy{}, ErrNotfoundVersion
}

func GetNextVersion(ver int) int {
	verRMu.RLock()
	defer verRMu.RUnlock()

	for _, v := range verList {
		if v.Version > ver {
			return v.Version
		}
	}

	// 保持不变
	return ver
}

type VerStrategy struct {
	Version int
	Reason  string
	Func    func(entity IEntity) error
}
