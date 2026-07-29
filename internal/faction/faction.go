// Package faction contains all the faction related code
package faction

import (
	"simuz/internal/relation"
	"sync"
)

type Faction struct {
	// --- Existing Core Fields ---
	ID                string `json:"id"`
	Name              string `json:"name"`
	relation.Relation `json:"Relation"`

	// --- 1. Economic & Resource Pools ---
	// Allows groups like the Smog-Iron Cartel or Bread-Weavers to hoard materials
	VaultGold int            `json:"vaultGold"`
	Stockpile map[string]int `json:"stockpile"` // e.g., {"raw_grain": 450, "iron_sword": 12}

	// --- 2. Territory & Spatial Influence ---
	// Tracks where this gang is strongest and what checkpoints they control
	HQLocationID    string             `json:"hqLocationId"`
	ControlledZones map[string]float64 `json:"controlledZones"` // LocationID -> Influence % (0.0 to 100.0)

	// --- 3. Demographic Caps & Tracking ---
	// Essential for looping through members or evaluating political leverage
	LeaderEntityID string          `json:"leaderEntityId"`
	MemberIDs      map[string]bool `json:"memberIds"`   // Fast lookup set of active member IDs
	MaxCapacity    int             `json:"maxCapacity"` // Soft cap before they stop spawning units

	// --- 4. Dynamic Strategic Agendas ---
	// Drives what the Lua AI scripts do behind the scenes
	CurrentState     string `json:"currentState"`     // "peaceful", "striking", "war_mobilization", "starving"
	PrimaryObjective string `json:"primaryObjective"` // e.g., "extort_wealth", "harvest_spores"
	WealthTier       int    `json:"wealthTier"`       // 1 = Slum beggars, 5 = High Scribe Nobility

	// --- 5. Thread Safety (Crucial for Game Loops) ---
	// Prevents concurrent map write panics when multiple entities read/write to the faction simultaneously
	mu sync.RWMutex
}

func (f *Faction) GetID() string {
	return f.ID
}
func GetFactionByID(id string) (*Faction, bool) {
	faction, exists := FactionRegistry[id]
	return faction, exists
}

// FactionRegistry is the one source of truth for all faction data in simuz.
// Every faction used in the simulation must have an entry here.
var FactionRegistry = map[string]*Faction{
	"smog_iron_cartel": {
		Relation: Relation{
			SpeciesRelation: SpeciesRelation{
				"orc":   3,  // Mostly run by Orcs
				"dwarf": 5,  // Heavy dwarven engineering bias
				"elf":   -6, // Dislikes delicate elven craft
			},
			FactionRelation: FactionRelation{
				"bread_weavers":     -3, // Standard trade/labor friction
				"withered_root":     -8, // Sabotaging their urban foundries
				"salt_vow_corsairs": 4,  // Good trade partners for moving metal
				"bleeding_quill":    -4, // Hate the tax collectors sniffing around
				"needle_eye_ring":   -2, // Annoyed by petty alley thievery
			},
			ProfessionRelation: ProfessionRelation{
				"blacksmith": 8,
				"guard":      -4, // Keep out of our underground foundries
				"tax_man":    -9, // Absolute zero-tolerance policy
			},
			EntityRelation: EntityRelation{},
		},
	},

	"withered_root": {
		Relation: Relation{
			SpeciesRelation: SpeciesRelation{
				"fay": 8,
				"elf": 4,
				"orc": -7, // Thinks orcs are loud, metallic tree-cutters
			},
			FactionRelation: FactionRelation{
				"smog_iron_cartel": -10, // Blood feud over urban tree cutting
				"needle_eye_ring":  6,   // Co-operate inside slums and crawlspaces
				"mire_blood_pack":  5,   // Shared love for the deep wild bogs
				"coffin_nail":      -4,  // Undead foul the pure forest soil
			},
			ProfessionRelation: ProfessionRelation{
				"guard":      -7,
				"herbalist":  8,
				"lumberjack": -12, // Immediate lethal hostility
			},
			EntityRelation: EntityRelation{},
		},
	},

	"coffin_nail": {
		Relation: Relation{
			SpeciesRelation: SpeciesRelation{
				"undead": 10, // Absolute unity with the dead
			},
			FactionRelation: FactionRelation{
				"withered_root":  -5,
				"bread_weavers":  0, // Indifferent as long as sanitation runs
				"rust_walkers":   4, // Respects the ancient marching dead
				"bleeding_quill": -3,
			},
			ProfessionRelation: ProfessionRelation{
				"priest":      -10, // Holy magic is fundamentally dangerous
				"gravedigger": 6,   // Handlers of the dead are respected allies
				"guard":       -2,
			},
			EntityRelation: EntityRelation{},
		},
	},

	"astrological_assembly": {
		Relation: Relation{
			SpeciesRelation: SpeciesRelation{
				"elf": 3, // Appreciates historical academic archives
			},
			FactionRelation: FactionRelation{
				"bleeding_quill":   -7, // Constantly fighting over forbidden text rights
				"smog_iron_cartel": 2,  // Buy dynamic machinery configurations from them
			},
			ProfessionRelation: ProfessionRelation{
				"wizard":     8,
				"royal_mage": -8, // Bitter dropouts hate the mainstream college
				"scholar":    5,
			},
			EntityRelation: EntityRelation{},
		},
	},

	"bread_weavers": {
		Relation: Relation{
			SpeciesRelation: SpeciesRelation{
				"dwarf": 4,
				"orc":   2, // Appreciates heavy lifting mill workers
			},
			FactionRelation: FactionRelation{
				"smog_iron_cartel": -2, // Foundries inflate fuel/coal prices
				"needle_eye_ring":  -9, // Thieves keep stealing our flour sacks!
				"bleeding_quill":   -8, // Corrupt tax auditors are a social plague
			},
			ProfessionRelation: ProfessionRelation{
				"baker":  10,
				"farmer": 8,
				"guard":  2, // Friendly unless a strike hits
			},
			EntityRelation: EntityRelation{},
		},
	},

	"salt_vow_corsairs": {
		Relation: Relation{
			SpeciesRelation: SpeciesRelation{
				"orc":   4,
				"dwarf": 2,
			},
			FactionRelation: FactionRelation{
				"smog_iron_cartel": 6,  // Smuggling iron makes great money
				"needle_eye_ring":  -3, // Keep the small pickpockets off our boats
				"mire_blood_pack":  -4, // Swamp ambushes interfere with river routes
			},
			ProfessionRelation: ProfessionRelation{
				"sailor":   10,
				"merchant": 4,
				"customs":  -8, // Evading port checkers is priority #1
			},
			EntityRelation: EntityRelation{},
		},
	},

	"bleeding_quill": {
		Relation: Relation{
			SpeciesRelation: SpeciesRelation{},
			FactionRelation: FactionRelation{
				"smog_iron_cartel":      -6, // Know they are hiding taxable gold bars
				"astrological_assembly": -8, // Hunting their heretical library cells
				"bread_weavers":         -5,
			},
			ProfessionRelation: ProfessionRelation{
				"tax_man": 10,
				"scribe":  8,
				"bandit":  -9, // Interfere with crown collection wagons
			},
			EntityRelation: EntityRelation{},
		},
	},

	"rust_walkers": {
		Relation: Relation{
			SpeciesRelation: SpeciesRelation{
				"undead": 2,
			},
			FactionRelation: FactionRelation{
				// Mindless swarm: treats every living group identically hostile
				"smog_iron_cartel":      -8,
				"withered_root":         -8,
				"coffin_nail":           0, // Ignores other undead structures completely
				"bread_weavers":         -8,
				"salt_vow_corsairs":     -8,
				"bleeding_quill":        -8,
				"needle_eye_ring":       -8,
				"mire_blood_pack":       -8,
				"astrological_assembly": -8,
			},
			ProfessionRelation: ProfessionRelation{},
			EntityRelation:     EntityRelation{},
		},
	},

	"needle_eye_ring": {
		Relation: Relation{
			SpeciesRelation: SpeciesRelation{
				"fay": 6,
				"elf": 4,
			},
			FactionRelation: FactionRelation{
				"withered_root":     5,  // Share hiding networks in city walls
				"bread_weavers":     -4, // Easy targets but they hit hard
				"salt_vow_corsairs": -2,
			},
			ProfessionRelation: ProfessionRelation{
				"thief": 9,
				"guard": -10, // Direct operational enemy
			},
			EntityRelation: EntityRelation{},
		},
	},

	"mire_blood_pack": {
		Relation: Relation{
			SpeciesRelation: SpeciesRelation{
				"orc": 3, // High concentration of swamp-orcs
			},
			FactionRelation: FactionRelation{
				"withered_root":     4,  // Respects nature reclaiming city boundaries
				"salt_vow_corsairs": -6, // Constantly sailing heavily armed through bogs
			},
			ProfessionRelation: ProfessionRelation{
				"knight":   -12, // Strip shiny armor collectors on sight
				"traveler": -6,
			},
			EntityRelation: EntityRelation{},
		},
	},
}
