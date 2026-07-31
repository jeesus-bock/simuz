// Package ai contains the AI runtime, script loading, and Lua-facing helpers for entities.
package ai

// Archetype is an enum-like type that classifies the default behavioral
// personality of an entity, influencing how it reacts to its surroundings.
type Archetype int

// Archetype constants define every possible behavioral archetype an entity
// can have. Each value represents a distinct AI behavior pattern.
const (
	// Passive entities ignore all other entities and follow their daily routine.
	Passive Archetype = iota
	// Aggressive entities attack any hostile entity within their perception range.
	Aggressive
	// Territorial entities attack any entity that enters a defined home zone.
	Territorial
	// Cowardly entities flee when their HP is low or when facing a stronger foe.
	Cowardly
	// Greedy entities prioritize looting items over engaging in combat.
	Greedy
	// Noble entities protect weaker entities and avoid attacking innocents.
	Noble
	// Curious entities investigate disturbances and unusual events in their area.
	Curious
	// Guarded entities warn others before resorting to attack.
	Guarded
	// Patrol entities walk a predefined path and report any intruders they spot.
	Patrol
	// Scripted entities use a Lua script to determine their behavior.
	Scripted
	// Dormant entities are inactive (typically deities) and do nothing until awakened.
	Dormant
	// Defensive entities stay near their home and only attack when directly threatened.
	Defensive
	// Hunting entities actively pursue prey when it is within range.
	Hunting
	// Gathering entities collect resources and avoid combat whenever possible.
	Gathering
	// Healing entities heal injured allies or creatures nearby.
	Healing
	// Scouting entities explore territory and flee when they feel threatened.
	Scouting
	// Guarding entities stay at a fixed post and attack any entity that approaches.
	Guarding
)

// String returns the human-readable name of the archetype as a lowercase string.
func (a Archetype) String() string {
	switch a {
	case Passive:
		return "passive"
	case Aggressive:
		return "aggressive"
	case Territorial:
		return "territorial"
	case Cowardly:
		return "cowardly"
	case Greedy:
		return "greedy"
	case Noble:
		return "noble"
	case Curious:
		return "curious"
	case Guarded:
		return "guarded"
	case Patrol:
		return "patrol"
	case Scripted:
		return "scripted"
	case Dormant:
		return "dormant"
	case Defensive:
		return "defensive"
	case Hunting:
		return "hunting"
	case Gathering:
		return "gathering"
	case Healing:
		return "healing"
	case Scouting:
		return "scouting"
	case Guarding:
		return "guarding"
	default:
		return "passive"
	}
}

// ParseArchetype converts a lowercase string into its corresponding Archetype
// value. If the string does not match any known archetype, it returns Passive.
func ParseArchetype(s string) Archetype {
	switch s {
	case "aggressive":
		return Aggressive
	case "territorial":
		return Territorial
	case "cowardly":
		return Cowardly
	case "greedy":
		return Greedy
	case "noble":
		return Noble
	case "curious":
		return Curious
	case "guarded":
		return Guarded
	case "patrol":
		return Patrol
	case "scripted":
		return Scripted
	case "dormant":
		return Dormant
	case "defensive":
		return Defensive
	case "hunting":
		return Hunting
	case "gathering":
		return Gathering
	case "healing":
		return Healing
	case "scouting":
		return Scouting
	case "guarding":
		return Guarding
	default:
		return Passive
	}
}

// ArchetypeInfo holds metadata about a single archetype, including its
// unique ID string, display name, and a short description of its behavior.
type ArchetypeInfo struct {
	// ID is the unique string identifier used for serialization and lookup.
	ID string `json:"id"`
	// Name is the human-readable display name of the archetype.
	Name string `json:"name"`
	// Description is a short summary of the archetype's behavior.
	Description string `json:"description"`
}

// Archetypes is the master list of all defined archetypes and their metadata.
// It is used for UI display, documentation, and runtime archetype lookup.
var Archetypes = []ArchetypeInfo{
	{ID: "passive", Name: "Passive", Description: "Ignores all entities, follows daily routine."},
	{ID: "aggressive", Name: "Aggressive", Description: "Attacks any hostile entity in perception range."},
	{ID: "territorial", Name: "Territorial", Description: "Attacks entities that enter a defined zone."},
	{ID: "cowardly", Name: "Cowardly", Description: "Flees when HP is low or facing stronger foe."},
	{ID: "greedy", Name: "Greedy", Description: "Prioritizes loot over combat."},
	{ID: "noble", Name: "Noble", Description: "Protects weaker entities."},
	{ID: "curious", Name: "Curious", Description: "Investigates disturbances."},
	{ID: "guarded", Name: "Guarded", Description: "Warns before attacking."},
	{ID: "patrol", Name: "Patrol", Description: "Walks a path, reports intruders."},
	{ID: "scripted", Name: "Scripted", Description: "Uses Lua script for behavior."},
	{ID: "dormant", Name: "Dormant", Description: "Inactive deity; does nothing until awakened."},
	{ID: "defensive", Name: "Defensive", Description: "Stays near home, attacks only when directly threatened."},
	{ID: "hunting", Name: "Hunting", Description: "Actively pursues prey when in range."},
	{ID: "gathering", Name: "Gathering", Description: "Collects resources, avoids combat."},
	{ID: "healing", Name: "Healing", Description: "Heals injured allies or creatures nearby."},
	{ID: "scouting", Name: "Scouting", Description: "Explores territory, flees when threatened."},
	{ID: "guarding", Name: "Guarding", Description: "Stays at a fixed post, attacks those who approach."},
}
