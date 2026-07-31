// Package engine contains the simulation engine, tick processing, and related systems.
package engine

import (
	"log"
	"math/rand"

	"simuz/internal/entity"
	"simuz/internal/events"
	"simuz/internal/species"
)

func processAging(ent *entity.Entity, sim *Simulation) {
	if !ent.Alive || ent.Immortal {
		return
	}

	ent.Age++
	if ent.MaxAge > 0 && ent.Age >= ent.MaxAge {
		oldAge(ent, sim)
		return
	}

	if ent.LastMealTick <= 0 {
		ent.LastMealTick = int(sim.Tick)
		return
	}

	if sp, ok := species.GetByID(ent.Species); ok && sp.AutoFeed {
		ent.LastMealTick = int(sim.Tick)
		return
	}

	starvationCheck(ent, sim)
}

func starvationCheck(ent *entity.Entity, sim *Simulation) {
	sp, ok := species.GetByID(ent.Species)
	if !ok || sp.StarvationThreshold <= 0 {
		return
	}

	ticksSinceMeal := int(sim.Tick) - ent.LastMealTick
	if ticksSinceMeal > sp.StarvationThreshold {
		interval := species.StarvationDamageInterval
		if (ticksSinceMeal-int(sp.StarvationThreshold))%interval == 0 {
			dmg := rand.Intn(species.StarvationDamageMax-species.StarvationDamageMin+1) + species.StarvationDamageMin
			ent.TakeDamage(dmg)
			sim.Emit(events.SimEvent{
				Type:   events.EventTypeStarvation,
				Source: ent.ID,
				Data: map[string]interface{}{
					"damage": dmg,
				},
			})
			if !ent.Alive {
				log.Printf("%s starved to death at tick %d (age %d)", ent.Name, sim.Tick, ent.Age)
			}
		}
	}
}

func oldAge(ent *entity.Entity, sim *Simulation) (lastWords string) {
	ent.Alive = false
	ent.TimeOfDeath = sim.Tick
	return lastWords
}
