package ai

type Archetype int

const (
	Passive Archetype = iota
	Aggressive
	Territorial
	Cowardly
	Greedy
	Noble
	Curious
	Guarded
	Patrol
	Scripted
	Dormant
	// New archetypes
	Defensive
	Hunting
	Gathering
	Healing
	Scouting
	Guarding
)

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

type ArchetypeInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

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
