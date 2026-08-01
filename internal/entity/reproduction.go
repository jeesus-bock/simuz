package entity

import (
	"math/rand"

	"simuz/internal/event"
)

type PregnancyOutcome string

const (
	OutcomeLiveBirth   PregnancyOutcome = "live_birth"
	OutcomeMiscarriage PregnancyOutcome = "miscarriage"
	OutcomeStillbirth  PregnancyOutcome = "stillbirth"
)

// PregnancyHistory logs what happened during past gestations.
type PregnancyHistory struct {
	TickCompleted uint64           `json:"tick_completed"`
	FatherID      string           `json:"father_id"`
	Outcome       PregnancyOutcome `json:"outcome"`
	Notes         string           `json:"notes,omitempty"` // e.g., "Died of childbed fever"
}

type Reproduction struct {
	Pregnant          bool   `json:"pregnant,omitempty"`
	PregnantSinceTick uint64 `json:"pregnant_since_tick,omitempty"`
	FatherID          string `json:"father_id,omitempty"`

	// Trackers for historical metrics and medical status
	ConceptionsCount int                `json:"conceptions_count"`
	History          []PregnancyHistory `json:"history,omitempty"`
}

// StartPregnancy initiates a pregnancy for the entity, recording the father
// and incrementing the conception counter. Returns a SimEvent for the pregnancy start.
func (r *Reproduction) StartPregnancy(fatherID string, currentTick uint64, motherID string) []event.SimEvent {
	r.Pregnant = true
	r.PregnantSinceTick = currentTick
	r.FatherID = fatherID
	r.ConceptionsCount++

	return []event.SimEvent{
		{
			Tick:   currentTick,
			Type:   "pregnancy_start",
			Entity: motherID,
			Data: map[string]interface{}{
				"father_id": fatherID,
			},
		},
	}
}

// MaintainPregnancy updates the pregnancy state each tick, handling gestation
// progression, complications (miscarriage, stillbirth), and outcomes.
// Call this each simulation tick for pregnant entities.
// Returns a slice of SimEvents for any events that occurred this tick.
func (r *Reproduction) MaintainPregnancy(currentTick uint64, gestationTicks int, e *Entity) []event.SimEvent {
	if !r.Pregnant {
		return nil
	}

	// If gestationTicks is invalid, treat as immediate birth.
	if gestationTicks <= 0 {
		r.Pregnant = false
		r.History = append(r.History, PregnancyHistory{
			TickCompleted: currentTick,
			FatherID:      r.FatherID,
			Outcome:       OutcomeLiveBirth,
			Notes:         "Instant birth (zero gestation)",
		})
		return []event.SimEvent{
			{
				Tick:   currentTick,
				Type:   "birth",
				Entity: e.ID,
				Data: map[string]interface{}{
					"father_id": r.FatherID,
					"outcome":   OutcomeLiveBirth,
				},
			},
		}
	}

	// If the mother has died during pregnancy, the pregnancy ends in stillbirth.
	if !e.Alive {
		r.Pregnant = false
		r.History = append(r.History, PregnancyHistory{
			TickCompleted: currentTick,
			FatherID:      r.FatherID,
			Outcome:       OutcomeStillbirth,
			Notes:         "Mother died during pregnancy",
		})
		return []event.SimEvent{
			{
				Tick:   currentTick,
				Type:   "stillbirth",
				Entity: e.ID,
				Data: map[string]interface{}{
					"father_id": r.FatherID,
					"outcome":   OutcomeStillbirth,
					"notes":     "Mother died during pregnancy",
				},
			},
		}
	}

	gestationProgress := currentTick - r.PregnantSinceTick
	if gestationProgress >= uint64(gestationTicks) {
		// Pregnancy has reached full term — live birth.
		r.Pregnant = false
		r.History = append(r.History, PregnancyHistory{
			TickCompleted: currentTick,
			FatherID:      r.FatherID,
			Outcome:       OutcomeLiveBirth,
		})
		return []event.SimEvent{
			{
				Tick:   currentTick,
				Type:   "birth",
				Entity: e.ID,
				Data: map[string]interface{}{
					"father_id": r.FatherID,
					"outcome":   OutcomeLiveBirth,
				},
			},
		}
	}

	// During pregnancy there is a small chance of complications.
	// Roll 1d100: ≤5% miscarriage, 6–8% stillbirth, otherwise normal progression.
	roll := rand.Intn(100)
	if roll < 5 {
		r.Pregnant = false
		r.History = append(r.History, PregnancyHistory{
			TickCompleted: currentTick,
			FatherID:      r.FatherID,
			Outcome:       OutcomeMiscarriage,
			Notes:         "Spontaneous miscarriage",
		})
		return []event.SimEvent{
			{
				Tick:   currentTick,
				Type:   "miscarriage",
				Entity: e.ID,
				Data: map[string]interface{}{
					"father_id": r.FatherID,
					"outcome":   OutcomeMiscarriage,
					"notes":     "Spontaneous miscarriage",
				},
			},
		}
	}
	if roll < 8 {
		r.Pregnant = false
		r.History = append(r.History, PregnancyHistory{
			TickCompleted: currentTick,
			FatherID:      r.FatherID,
			Outcome:       OutcomeStillbirth,
			Notes:         "Late-term complication",
		})
		return []event.SimEvent{
			{
				Tick:   currentTick,
				Type:   "stillbirth",
				Entity: e.ID,
				Data: map[string]interface{}{
					"father_id": r.FatherID,
					"outcome":   OutcomeStillbirth,
					"notes":     "Late-term complication",
				},
			},
		}
	}

	// Pregnancy is still in progress — no event this tick.
	return nil
}
