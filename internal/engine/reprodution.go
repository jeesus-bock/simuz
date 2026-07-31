package engine

import (
	"fmt"
	"math/rand"
	"simuz/internal/entity"
	"simuz/internal/world"
)

const (
	GestationDurationTicks = 1200 // Adjust to match your game time scale
)

// UpdatePregnancy tracks gestational milestones, handles random miscarriages,
// and evaluates historical childbed/labor risks once the duration concludes.
// UpdatePregnancy tracks gestational milestones, handles random miscarriages,
// and evaluates historical childbed/labor risks once the duration concludes.
func (s *Simulation) UpdatePregnancy(mother *entity.Entity, tm *world.GameTime, rng *rand.Rand, em *entity.EntityManager) {
	ticksPregnant := tm.Tick - mother.Reproduction.PregnantSinceTick
	if ticksPregnant < GestationDurationTicks {
		return // not yet full term
	}
	// ~1.5% resolution chance per pass past full term
	if rng.Intn(1000) < 15 {
		mother.Reproduction.Pregnant = false
		roll := rng.Intn(100)

		if roll < 25 {
			// Outcome A: Medieval Stillbirth right at labor (25% chance)
			mother.Reproduction.History = append(mother.Reproduction.History, entity.PregnancyHistory{
				TickCompleted: tm.Tick,
				FatherID:      mother.Reproduction.FatherID,
				Outcome:       entity.OutcomeStillbirth,
				Notes:         "Infant failed to draw breath",
			})
			mother.Mood = "depressed"
			mother.HP -= 30
			fmt.Printf("[TRAGEDY] %s delivered a stillborn infant after full term gestation.\n", mother.Name)

		} else if roll < 35 {
			// Outcome B: Fatal Childbed Fever / Maternal Mortality (10% chance)
			mother.Alive = false
			mother.HP = 0
			mother.TimeOfDeath = tm.Tick
			mother.Mood = "dead"
			fmt.Printf("[MORTALITY] %s tragically succumbed to childbed fever during labor.\n", mother.Name)

		} else {
			// Outcome C: Successful Live Birth! (65% chance)
			mother.Reproduction.History = append(mother.Reproduction.History, entity.PregnancyHistory{
				TickCompleted: tm.Tick,
				FatherID:      mother.Reproduction.FatherID,
				Outcome:       entity.OutcomeLiveBirth,
			})
			mother.Mood = "happy"
			fmt.Printf("[BIRTH] %s successfully gave birth to a healthy child after %d ticks!\n", mother.Name, ticksPregnant)
		}
	}
}
