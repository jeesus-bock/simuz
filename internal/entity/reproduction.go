package entity

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

// MaintainPregnancy updates the pregnancy state for civilized species,
// ensuring proper gestation progression and handling complications.
// This function should be called each simulation tick for pregnant entities
// of civilized species including kobolds, hobbits, gnolls, and others.
// Returns true if the pregnancy has resulted in a live birth.
func (r *Reproduction) MaintainPregnancy(currentTick uint64, gestationTicks int) (birthOccurred bool) {
	if !r.Pregnant {
		return false
	}

	gestationProgress := currentTick - r.PregnantSinceTick
	if gestationProgress >= uint64(gestationTicks) {
		// Pregnancy has reached full term
		r.Pregnant = false
		r.History = append(r.History, PregnancyHistory{
			TickCompleted: currentTick,
			Outcome:       OutcomeLiveBirth,
		})
		return true
	}

	// Pregnancy is still in progress - civilized species have reduced
	// complication rates due to structured care and support systems.
	return false
}
