package entitytest

import (
	"context"
	"reflect"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/core/actor"
)

type EntityWapper struct {
	cs       core.ICacheStrategy `bson:"-"`
	Bag      *EntityBagModule
	Airship  *EntityAirshipModule
	User     *EntityUserModule
	TimeInfo *EntityTimeInfoModule

	// Used to determine if it was read from cache
	isCache bool `bson:"-"`
}

var (
	userModuleType     = reflect.TypeOf(&EntityUserModule{})
	airshipModuleType  = reflect.TypeOf(&EntityAirshipModule{})
	bagModuleType      = reflect.TypeOf(&EntityBagModule{})
	timeInfoModuleType = reflect.TypeOf(&EntityTimeInfoModule{})
)

func NewEntityWapper(id string) *EntityWapper {
	return &EntityWapper{
		cs: actor.BuildEntityLoader(id, []reflect.Type{
			userModuleType,
			airshipModuleType,
			bagModuleType,
			timeInfoModuleType,
		}),
	}
}

func (e *EntityWapper) Load() error {
	err := e.cs.Load(context.TODO())
	if err != nil {
		return err
	}

	e.Bag = e.cs.GetModule(bagModuleType).(*EntityBagModule)
	e.Airship = e.cs.GetModule(airshipModuleType).(*EntityAirshipModule)
	e.User = e.cs.GetModule(userModuleType).(*EntityUserModule)
	e.TimeInfo = e.cs.GetModule(timeInfoModuleType).(*EntityTimeInfoModule)

	e.isCache = true

	return nil
}

func (e *EntityWapper) Sync() error {

	if !e.isCache {
		e.cs.SetModule(bagModuleType, e.Bag)
		e.cs.SetModule(airshipModuleType, e.Airship)
		e.cs.SetModule(userModuleType, e.User)
		e.cs.SetModule(timeInfoModuleType, e.TimeInfo)
	}

	return e.cs.Sync(context.TODO())
}
