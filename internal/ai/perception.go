// Package ai contains the AI runtime, script loading, and Lua-facing helpers for entities.
package ai

type Perception struct {
	Range        float64
	FOV          float64
	HearingRange float64
	NightVision  float64
}

func NewPerception() Perception {
	return Perception{
		Range:        20,
		FOV:          120,
		HearingRange: 30,
		NightVision:  1.0,
	}
}
