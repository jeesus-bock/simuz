// Package entity defines the simulation entities, their attributes, and related behaviors.
package entity

// SpeciesMaxAge returns the maximum age for a species in ticks.
// Returns 0 for immortal species or unknown species.
func SpeciesMaxAge(species string) int {
	return SpeciesMaxAge(species)
}

// SpeciesStarvationThreshold returns the tick threshold after which a species starts taking starvation damage.
func SpeciesStarvationThreshold(species string) int {
	return SpeciesStarvationThreshold(species)
}

// ShouldAutoFeed returns whether a species requires automatic feeding.
func ShouldAutoFeed(species string) bool {
	return ShouldAutoFeed(species)
}

// StarvationDamageInterval returns the tick interval at which starvation damage is applied.
func StarvationDamageInterval() int {
	return StarvationDamageInterval()
}

// CanLevelUp returns whether a species can gain levels.
func CanLevelUp(species string) bool {
	return CanLevelUp(species)
}

// CanReproduce returns whether a species can reproduce naturally.
func CanReproduce(species string) bool {
	return CanReproduce(species)
}

// IsCavemanSpecies returns whether a species reproduces without forming mate bonds.
func IsCavemanSpecies(species string) bool {
	return IsCavemanSpecies(species)
}

// GestationTicksForSpecies returns the gestation period in ticks for a species.
// Falls back to 200 ticks if the species has no entry.
func GestationTicksForSpecies(species string) int {
	return GestationTicksForSpecies(species)
}

// SpeciesDefaultScripts returns the default AI script IDs for a species.
func SpeciesDefaultScripts(species string) []string {
	return SpeciesDefaultScripts(species)
}

// SpeciesDefaultSleepCycle returns the default sleep cycle for a species.
func SpeciesDefaultSleepCycle(species string) string {
	return SpeciesDefaultSleepCycle(species)
}

// SpeciesNames returns the name pool for a species.
func SpeciesNames(species string) []string {
	return SpeciesNames(species)
}

// SpeciesBaseAttrs returns the base attributes for a species.
func SpeciesBaseAttrs(species string) Attributes {
	return SpeciesBaseAttrs(species)
}
