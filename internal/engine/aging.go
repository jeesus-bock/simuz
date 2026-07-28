// Package engine contains the simulation engine, tick processing, and related systems.
package engine

import (
	"log"
	"math/rand"

	"simuz/internal/entity"
	"simuz/internal/events"
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

	if species, ok := entity.GetSpeciesByID(ent.Species); ok && species.AutoFeed {
		ent.LastMealTick = int(sim.Tick)
		return
	}

	starvationCheck(ent, sim)
}

func starvationCheck(ent *entity.Entity, sim *Simulation) {
	// Determine starvation threshold from species if available
	threshold := ent.StarvationThreshold
	if threshold <= 0 {
		return
	}

	ticksSinceMeal := int(sim.Tick) - ent.LastMealTick
	if ticksSinceMeal > threshold {
		// Get faction for the entity to possibly adjust interval
		faction, ok := entity.GetFactionByID(ent.Faction)
		if ok {
			// Use faction-specific interval if defined, else default
			interval := entity.StarvationDamageInterval
			if (ticksSinceMeal - threshold) % interval == 0 {
				dmg := rand.Intn(entity.StarvationDamageMax() - entity.StarvationDamageMin() + 1) + entity.StarvationDamageMin()
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
}

func oldAge(ent *entity.Entity, sim *Simulation) {
	// Entity dies of old age
	ent.Alive = false
	log.Printf("%s died of old age at tick %d (age %d)", ent.Name, sim.Tick, ent.Age)
	// Optionally emit an event if needed; for now just log.
}
