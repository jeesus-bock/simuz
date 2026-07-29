// Package ai contains the AI runtime, script loading, and Lua-facing helpers for entities.
package ai

// Perception defines the sensory capabilities of an entity, determining how
// far and how well it can detect other entities and events in the world.
type Perception struct {
	// Range is the maximum distance (in game units) at which the entity can
	// visually detect other entities or objects.
	Range float64
	// FOV is the field of view in degrees, defining the angular width of the
	// entity's cone of vision.
	FOV float64
	// HearingRange is the maximum distance at which the entity can detect
	// sounds, such as footsteps or combat noises.
	HearingRange float64
	// NightVision is a multiplier that enhances the entity's visual range
	// in low-light or dark conditions. A value of 1.0 means no bonus.
	NightVision float64
}

// NewPerception returns a Perception struct with default sensory values
// suitable for a standard entity: 20 units of sight range, 120-degree FOV,
// 30 units of hearing range, and no night vision bonus.
func NewPerception() Perception {
	return Perception{
		Range:        20,
		FOV:          120,
		HearingRange: 30,
		NightVision:  1.0,
	}
}
