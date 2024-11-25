package mockentity

import (
	"context"
	"reflect"

	"github.com/pojol/braid/core"
)

type EntityWapper struct {
	ID       string                `bson:"_id"`
	cs       core.ICacheStrategy   `bson:"-"`
	Bag      *EntityBagModule      `bson:"bag"`
	Airship  *EntityAirshipModule  `bson:"airship"`
	User     *EntityUserModule     `bson:"user"`
	TimeInfo *EntityTimeInfoModule `bson:"time_info"`

	// Used to determine if it was read from cache
	isCache bool `bson:"-"`
}

func (e *EntityWapper) GetID() string {
	return e.ID
}

func (e *EntityWapper) SetModule(moduleType reflect.Type, module interface{}) {
	switch moduleType {
	case reflect.TypeOf(&EntityBagModule{}):
		e.Bag = module.(*EntityBagModule)
	case reflect.TypeOf(&EntityAirshipModule{}):
		e.Airship = module.(*EntityAirshipModule)
	case reflect.TypeOf(&EntityUserModule{}):
		e.User = module.(*EntityUserModule)
	case reflect.TypeOf(&EntityTimeInfoModule{}):
		e.TimeInfo = module.(*EntityTimeInfoModule)
	}
}

func (e *EntityWapper) GetModule(moduleType reflect.Type) interface{} {
	switch moduleType {
	case reflect.TypeOf(&EntityBagModule{}):
		return e.Bag
	case reflect.TypeOf(&EntityAirshipModule{}):
		return e.Airship
	case reflect.TypeOf(&EntityUserModule{}):
		return e.User
	case reflect.TypeOf(&EntityTimeInfoModule{}):
		return e.TimeInfo
	}
	return nil
}

func NewEntityWapper(id string) *EntityWapper {
	e := &EntityWapper{
		ID: id,
	}
	e.cs = BuildEntityLoader("braid-test", "entity_test", e)
	return e
}

func (e *EntityWapper) Load(ctx context.Context) error {
	err := e.cs.Load(ctx)
	if err != nil {
		return err
	}

	e.isCache = true

	return nil
}

func (e *EntityWapper) Sync(ctx context.Context, forceUpdate bool) error {
	return e.cs.Sync(ctx, forceUpdate)
}

func (e *EntityWapper) Store(ctx context.Context) error {
	return e.cs.Store(ctx)
}

func (e *EntityWapper) IsDirty() bool {
	return e.cs.IsDirty()
}
