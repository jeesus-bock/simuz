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
	if species, ok := entity.GetSpeciesByID(ent.Species); ok {
		if threshold := species.StarvationThreshold; threshold <= 0 {
			return
		}
	}

	ticksSinceMeal := int(sim.Tick) - ent.LastMealTick
	if ticksSinceMeal > ent.StarvationThreshold {
		interval, ok := entity.GetFaction(key[i])
		if ok && ent.Faction == key[i] {
			interval = entity.StarvationDamageInterval
			if (ticksSinceMeal-int(ent.StarvationThreshold))%interval == 0 {
				dmg := rand.Intn(entity.StarvationDamageMax()-entity.StarvationDamageMin()+1) + entity.StarvationDamageMin()
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

func oldAge{ent *entity.Entityn, sim *Simulation((lastWord string){

}