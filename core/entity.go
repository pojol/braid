package core

import (
	"context"
	"errors"
	"reflect"
	"sync"
)

type IEntity interface {
	GetID() string
	SetModule(typ reflect.Type, module interface{})
	GetModule(typ reflect.Type) interface{}
}

type ICacheStrategy interface {
	// Load - Loads data, prioritizing retrieval from the cache layer. If not found in cache, it pulls from the database, stores in cache, and then returns.
	Load(ctx context.Context) error

	// Sync - Synchronizes memory data back to the cache
	Sync(ctx context.Context) error

	// Store - Stores data to the database and clears the cache
	Store(ctx context.Context) error

	// IsDirty - Checks if the data is dirty
	IsDirty() bool

	// GetModule - Retrieves the corresponding module from the loader by type
	GetModule(typ reflect.Type) interface{}

	// SetModule - Sets the module to the loader (usually during entity initialization)
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
	Func    func(entity interface{}) error
}
