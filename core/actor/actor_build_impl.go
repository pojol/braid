package actor

import (
	"context"

	"github.com/pojol/braid/core"
)

// ActorLoaderBuilder used to build ActorLoader
type ActorLoaderBuilder struct {
	core.ISystem
	core.ActorConstructor
	core.IActorLoader
}

func (p *ActorLoaderBuilder) WithID(id string) core.IActorBuilder {
	if id == "" {
		panic("[braid.actor] id is empty")
	}
	p.ID = id
	return p
}

func (p *ActorLoaderBuilder) WithType(ty string) core.IActorBuilder {
	p.Name = ty
	return p
}

func (p *ActorLoaderBuilder) WithOpt(key string, value string) core.IActorBuilder {
	p.Options[key] = value
	return p
}

func (p *ActorLoaderBuilder) GetID() string {
	return p.ID
}

func (p *ActorLoaderBuilder) GetType() string {
	return p.Name
}

func (p *ActorLoaderBuilder) GetWeight() int {
	return p.Weight
}

func (p *ActorLoaderBuilder) GetGlobalQuantityLimit() int {
	return p.GlobalQuantityLimit
}

func (p *ActorLoaderBuilder) GetNodeUnique() bool {
	return p.NodeUnique
}

func (p *ActorLoaderBuilder) GetOptions() map[string]string {
	return p.Options
}

func (p *ActorLoaderBuilder) GetOpt(key string) string {
	return p.Options[key]
}

func (p *ActorLoaderBuilder) GetSystem() core.ISystem {
	return p.ISystem
}

func (p *ActorLoaderBuilder) GetLoader() core.IActorLoader {
	return p.IActorLoader
}

func (p *ActorLoaderBuilder) GetConstructor() core.CreateFunc {
	return p.Constructor
}

func (p *ActorLoaderBuilder) Register(ctx context.Context) (core.IActor, error) {
	return p.ISystem.Register(ctx, p)
}

func (p *ActorLoaderBuilder) Picker(ctx context.Context) error {
	return p.IActorLoader.Pick(ctx, p) // Note: This method is asynchronous
}
