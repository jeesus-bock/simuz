// Package species is the best
package species

import (
	"math/rand"
)

// Species defines the base data for a creature species in the simulation.
// It is the single source of truth for all species-related information.
//
// Time units in this struct:
//   - MaxAge, AdultAge: game-days (converted to ticks at comparison time using game speed)
//   - GestationTicks: simulation ticks (1 tick = 1 real second at 1 Hz)
//   - StarvationThreshold: simulation ticks
//
// At the default speed of 24, one game-day = 1440/24 = 60 ticks.
type Species struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	MaxAge              int        `json:"max_age"`     // years: natural lifespan
	AdultAge            int        `json:"adult_age"`   // years: age at which entity can reproduce
	CanLevelUp          bool       `json:"can_level_up"`
	CanReproduce        bool       `json:"can_reproduce"`
	IsCaveman           bool       `json:"is_caveman"`
	IsImmortal          bool       `json:"is_immortal"`
	GestationTicks      int        `json:"gestation_ticks"` // simulation ticks: pregnancy duration (1 tick = 1 sec at speed 1)
	DefaultScripts      []string   `json:"default_scripts,omitempty"`
	DefaultSleepCycle   string     `json:"default_sleep_cycle"` // "diurnal", "nocturnal", "none"
	AutoFeed            bool       `json:"auto_feed"`
	StarvationThreshold int        `json:"starvation_threshold"` // simulation ticks before starvation damage; 0 = immune
	MaleNames           []string   `json:"male_names,omitempty"`
	FemaleNames         []string   `json:"female_names,omitempty"`
	BaseAttrs           Attributes `json:"base_attrs"`

	// --- Optimized Grand Design Fields ---
	FluidBiology    bool     `json:"fluid_biology"`              // True for Fey/Planar entities to alternate reproductive states over time
	PreferredBiomes []string `json:"preferred_biomes,omitempty"` // For spatial placement filters
}

// GetRandomName handles procedural linguistic choices using a fallback protection check.
func (s *Species) GetRandomName(gender string, rng *rand.Rand) string {
	if gender == "male" && len(s.MaleNames) > 0 {
		return s.MaleNames[rng.Intn(len(s.MaleNames))]
	}
	if gender == "female" && len(s.FemaleNames) > 0 {
		return s.FemaleNames[rng.Intn(len(s.FemaleNames))]
	}

	// Non-binary or fallback strategy: Combine or draw randomly from the total pool
	combinedPool := append(s.MaleNames, s.FemaleNames...)
	if len(combinedPool) > 0 {
		return combinedPool[rng.Intn(len(combinedPool))]
	}
	return "Nameless " + s.Name
}

// GetByID returns the Species definition for a given ID.
func GetByID(id string) (Species, bool) {
	s, exists := Registry[id]
	return s, exists
}

var StarvationDamageInterval = 10
var StarvationDamageMin = 10
var StarvationDamageMax = 20
