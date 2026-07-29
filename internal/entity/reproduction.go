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
