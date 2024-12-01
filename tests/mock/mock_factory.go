package mock

import (
	"github.com/pojol/braid/core"
)

// MockActorFactory is a factory for creating actors
type MockActorFactory struct {
	Constructors map[string]*core.ActorConstructor
}

// NewActorFactory create new actor factory
func BuildActorFactory() *MockActorFactory {
	factory := &MockActorFactory{
		Constructors: make(map[string]*core.ActorConstructor),
	}

	factory.Constructors["MockDynamicPicker"] = &core.ActorConstructor{
		ID:                  "MockDynamicPicker",
		Name:                "MockDynamicPicker",
		Weight:              100,
		Constructor:         NewDynamicPickerActor,
		NodeUnique:          true,
		GlobalQuantityLimit: 10,
		Options:             make(map[string]string),
	}

	factory.Constructors["MockDynamicRegister"] = &core.ActorConstructor{
		ID:                  "MockDynamicRegister",
		Name:                "MockDynamicRegister",
		Weight:              100,
		Constructor:         NewDynamicRegisterActor,
		NodeUnique:          true,
		GlobalQuantityLimit: 0,
		Options:             make(map[string]string),
	}

	factory.Constructors["MockActorControl"] = &core.ActorConstructor{
		ID:                  "MockActorControl",
		Name:                "MockActorControl",
		Weight:              100,
		Constructor:         NewControlActor,
		NodeUnique:          true,
		GlobalQuantityLimit: 0,
		Options:             make(map[string]string),
	}

	factory.Constructors["mocka"] = &core.ActorConstructor{
		ID:          "mocka",
		Name:        "mocka",
		Weight:      100,
		Constructor: newMockA,
		NodeUnique:  true,
		Dynamic:     true,
	}

	factory.Constructors["mockb"] = &core.ActorConstructor{
		ID:          "mockb",
		Name:        "mockb",
		Weight:      100,
		Constructor: newMockB,
		NodeUnique:  true,
		Dynamic:     true,
	}

	factory.Constructors["mockc"] = &core.ActorConstructor{
		ID:          "mockc",
		Name:        "mockc",
		Weight:      100,
		Constructor: newMockC,
		NodeUnique:  true,
		Dynamic:     true,
	}

	return factory
}

func (factory *MockActorFactory) Get(actorType string) *core.ActorConstructor {
	if _, ok := factory.Constructors[actorType]; ok {
		return factory.Constructors[actorType]
	}

	return nil
}

func (factory *MockActorFactory) GetActors() []*core.ActorConstructor {
	actors := []*core.ActorConstructor{}
	for _, v := range factory.Constructors {
		actors = append(actors, v)
	}
	return actors
}
