package ai

import (
	"simuz/internal/entity"
	"simuz/internal/world"
)

type EntityLister interface {
	ByLocation(locID string) []*entity.Entity
	All() []*entity.Entity
}

type Context struct {
	Entity   *entity.Entity
	World    *world.World
	Entities EntityLister
	Time     *world.GameTime
}

type AISystem struct{}

func New() *AISystem {
	return &AISystem{}
}

func (a *AISystem) Tick(ctx *Context) {
	switch ctx.Entity.AI.Type {
	case "passive":
	case "aggressive":
		nearby := ctx.Entities.ByLocation(ctx.Entity.LocationID)
		for _, other := range nearby {
			if other.ID == ctx.Entity.ID || !other.Alive {
				continue
			}
			_ = other
		}
	}
}
