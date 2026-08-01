package engine

import (
	"fmt"
	"log"
	"maps"
	"math/rand"
	"slices"

	"simuz/internal/entity"
	"simuz/internal/items"
	"simuz/internal/relation"
	"simuz/internal/species"
	"simuz/internal/world"
)

// DefaultMaxLevel is the sane cap applied when a species has no
// hardcoded level rule. MaxAge is a lifespan in years, not a level,
// so using it directly would produce absurd spawn levels (e.g. elf 700).
const DefaultMaxLevel = 10

type SpawnRule struct {
	ID              string
	LocationID      string
	Species         string
	Faction         string
	Profession      string
	FactionID       string
	DesiredCount    int
	Interval        int
	MinLevel        int
	MaxLevel        int
	Equipment       []string
	RequireFaction  string // if set, only spawn when location controlled by this faction (or uncontrolled)
	BlockIfEnemyCtl bool   // skip if location controlled by hostile-to-faction controller
}

type SpawnManager struct {
	Rules []SpawnRule
}

// NewSpawnManager creates a SpawnManager with rules for every species
// registered in the species registry. Hardcoded rules for known species
// are preserved; any species without a specific rule gets a default entry
// so that all species can appear during initialization.
func NewSpawnManager() *SpawnManager {
	sm := &SpawnManager{}

	// 1. Add hardcoded rules for species with location-specific spawning.
	sm.Rules = []SpawnRule{
		// --- Core species ---
		{ID: "orc_patrol", LocationID: "orc_camp", Species: "orc", Faction: "", Profession: "warrior", FactionID: "orc", DesiredCount: 4, Interval: 120, MinLevel: 1, MaxLevel: 3},
		{ID: "wolf_pack", LocationID: "wolf_den", Species: "wolf", Faction: "", Profession: "", FactionID: "beast", DesiredCount: 4, Interval: 90, MinLevel: 1, MaxLevel: 2},
		{ID: "bandit_camp", LocationID: "bandit_camp", Species: "human", Faction: "", Profession: "bandit", FactionID: "bandit", DesiredCount: 4, Interval: 150, MinLevel: 1, MaxLevel: 2},
		{ID: "bear_den", LocationID: "bear_den", Species: "bear", Faction: "", Profession: "", FactionID: "beast", DesiredCount: 2, Interval: 200, MinLevel: 3, MaxLevel: 5},
		{ID: "boar_herd", LocationID: "boar_wallow", Species: "boar", Faction: "", Profession: "", FactionID: "beast", DesiredCount: 2, Interval: 180, MinLevel: 1, MaxLevel: 3},
		{ID: "rat_infest", LocationID: "rat_king_lair_entrance", Species: "rat", Faction: "", Profession: "", FactionID: "vermin", DesiredCount: 3, Interval: 60, MinLevel: 1, MaxLevel: 1},
		{ID: "rat_corridor", LocationID: "rat_king_lair_corridor", Species: "rat", Faction: "", Profession: "", FactionID: "vermin", DesiredCount: 3, Interval: 60, MinLevel: 1, MaxLevel: 2},
		{ID: "spider_nest", LocationID: "spider_grove", Species: "spider", Faction: "", Profession: "", FactionID: "beast", DesiredCount: 2, Interval: 120, MinLevel: 1, MaxLevel: 3},
		{ID: "goblin_gatherers", LocationID: "goblin_hollow", Species: "goblin", Faction: "", Profession: "gatherer", FactionID: "goblin", DesiredCount: 2, Interval: 180, MinLevel: 1, MaxLevel: 1},
		{ID: "kobold_warren", LocationID: "kobold_warren", Species: "kobold", Faction: "", Profession: "warrior", FactionID: "kobold", DesiredCount: 4, Interval: 150, MinLevel: 1, MaxLevel: 2},
		{ID: "ash_scorpions", LocationID: "scorpion_dunes", Species: "spider", Faction: "", Profession: "", FactionID: "beast", DesiredCount: 2, Interval: 160, MinLevel: 2, MaxLevel: 4},
		{ID: "ash_orcs", LocationID: "ash_ruins", Species: "orc", Faction: "", Profession: "warrior", FactionID: "orc", DesiredCount: 2, Interval: 180, MinLevel: 2, MaxLevel: 4},
		{ID: "town_bard", LocationID: "tavern", Species: "human", Faction: "", Profession: "bard", FactionID: "", DesiredCount: 1, Interval: 0, MinLevel: 1, MaxLevel: 3},
		{ID: "town_priest", LocationID: "temple", Species: "human", Faction: "", Profession: "priest", FactionID: "", DesiredCount: 1, Interval: 0, MinLevel: 1, MaxLevel: 3},
		{ID: "human_politician", LocationID: "tavern", Species: "human", Faction: "", Profession: "politician", FactionID: "", DesiredCount: 1, Interval: 0, MinLevel: 3, MaxLevel: 5},
		{ID: "orc_chief", LocationID: "orc_camp", Species: "orc", Faction: "", Profession: "politician", FactionID: "orc", DesiredCount: 1, Interval: 0, MinLevel: 4, MaxLevel: 6},
		{ID: "dwarf_thane", LocationID: "dwarf_keep", Species: "dwarf", Faction: "", Profession: "politician", FactionID: "", DesiredCount: 1, Interval: 0, MinLevel: 4, MaxLevel: 6},
		{ID: "elf_archon", LocationID: "fey_glade", Species: "elf", Faction: "", Profession: "politician", FactionID: "", DesiredCount: 1, Interval: 0, MinLevel: 5, MaxLevel: 7},

		// --- Half-species ---
		{ID: "half_orc_warrior", LocationID: "orc_camp", Species: "half_orc", Faction: "", Profession: "warrior", FactionID: "orc", DesiredCount: 2, Interval: 0, MinLevel: 1, MaxLevel: 4},
		{ID: "half_elf_scholar", LocationID: "fey_glade", Species: "half_elf", Faction: "", Profession: "scholar", FactionID: "", DesiredCount: 1, Interval: 0, MinLevel: 1, MaxLevel: 5},
		{ID: "half_dwarf_miner", LocationID: "dwarf_keep", Species: "half_dwarf", Faction: "", Profession: "miner", FactionID: "", DesiredCount: 2, Interval: 0, MinLevel: 1, MaxLevel: 4},
		{ID: "half_goblin_scavenger", LocationID: "goblin_hollow", Species: "half_goblin", Faction: "", Profession: "gatherer", FactionID: "goblin", DesiredCount: 2, Interval: 0, MinLevel: 1, MaxLevel: 2},
		{ID: "half_hobgoblin_soldier", LocationID: "orc_camp", Species: "half_hobgoblin", Faction: "", Profession: "warrior", FactionID: "orc", DesiredCount: 2, Interval: 0, MinLevel: 2, MaxLevel: 5},
		{ID: "half_gnoll_marauder", LocationID: "orc_camp", Species: "half_gnoll", Faction: "", Profession: "warrior", FactionID: "orc", DesiredCount: 2, Interval: 0, MinLevel: 1, MaxLevel: 3},
		{ID: "half_kobold_scout", LocationID: "kobold_warren", Species: "half_kobold", Faction: "", Profession: "scout", FactionID: "kobold", DesiredCount: 2, Interval: 0, MinLevel: 1, MaxLevel: 2},
		{ID: "half_fey_druid", LocationID: "fey_glade", Species: "half_fey", Faction: "", Profession: "druid", FactionID: "", DesiredCount: 1, Interval: 0, MinLevel: 2, MaxLevel: 5},

		// --- Undead ---
		{ID: "skeleton_graveyard", LocationID: "graveyard", Species: "skeleton", Faction: "", Profession: "warrior", FactionID: "undead", DesiredCount: 4, Interval: 120, MinLevel: 1, MaxLevel: 3},
		{ID: "zombie_moat", LocationID: "castle_dungeon", Species: "zombie", Faction: "", Profession: "", FactionID: "undead", DesiredCount: 3, Interval: 90, MinLevel: 1, MaxLevel: 2},
		{ID: "ghost_keep", LocationID: "castle_keep", Species: "ghost", Faction: "", Profession: "", FactionID: "undead", DesiredCount: 2, Interval: 150, MinLevel: 3, MaxLevel: 5},
		{ID: "wraith_crypt", LocationID: "crypt", Species: "wraith", Faction: "", Profession: "", FactionID: "undead", DesiredCount: 2, Interval: 180, MinLevel: 4, MaxLevel: 6},
		{ID: "lich_tower", LocationID: "dark_tower", Species: "lich", Faction: "", Profession: "necromancer", FactionID: "undead", DesiredCount: 1, Interval: 0, MinLevel: 7, MaxLevel: 9},
		{ID: "vampire_mansion", LocationID: "mansion", Species: "vampire", Faction: "", Profession: "politician", FactionID: "undead", DesiredCount: 1, Interval: 0, MinLevel: 5, MaxLevel: 7},

		// --- Fey & magical ---
		{ID: "fairy_grove", LocationID: "fey_glade", Species: "fairy", Faction: "", Profession: "", FactionID: "fey", DesiredCount: 3, Interval: 120, MinLevel: 1, MaxLevel: 3},
		{ID: "dryad_forest", LocationID: "ancient_forest", Species: "dryad", Faction: "", Profession: "", FactionID: "fey", DesiredCount: 2, Interval: 150, MinLevel: 2, MaxLevel: 4},
		{ID: "satyr_cavern", LocationID: "mountain_cave", Species: "satyr", Faction: "", Profession: "", FactionID: "fey", DesiredCount: 2, Interval: 140, MinLevel: 2, MaxLevel: 4},
		{ID: "pixie_mushroom", LocationID: "mushroom_forest", Species: "pixie", Faction: "", Profession: "", FactionID: "fey", DesiredCount: 4, Interval: 100, MinLevel: 1, MaxLevel: 2},
		{ID: "treant_wood", LocationID: "ancient_forest", Species: "treant", Faction: "", Profession: "", FactionID: "fey", DesiredCount: 1, Interval: 0, MinLevel: 5, MaxLevel: 7},

		// --- Dragons & reptiles ---
		{ID: "dragon_lair", LocationID: "dragon_cave", Species: "dragon", Faction: "", Profession: "", FactionID: "dragon", DesiredCount: 1, Interval: 0, MinLevel: 8, MaxLevel: 10},
		{ID: "lizardfolk_swamp", LocationID: "swamp", Species: "lizardfolk", Faction: "", Profession: "warrior", FactionID: "lizardfolk", DesiredCount: 3, Interval: 150, MinLevel: 2, MaxLevel: 4},
		{ID: "wyvern_peak", LocationID: "mountain_peak", Species: "wyvern", Faction: "", Profession: "", FactionID: "dragon", DesiredCount: 2, Interval: 180, MinLevel: 3, MaxLevel: 5},
		{ID: "basilisk_lair", LocationID: "cave_system", Species: "basilisk", Faction: "", Profession: "", FactionID: "beast", DesiredCount: 1, Interval: 0, MinLevel: 4, MaxLevel: 6},

		// --- Fey & small races ---
		{ID: "gnome_hollow", LocationID: "gnome_hollow", Species: "gnome", Faction: "", Profession: "miner", FactionID: "gnome", DesiredCount: 3, Interval: 150, MinLevel: 1, MaxLevel: 3},
		{ID: "halfling_village", LocationID: "halfling_village", Species: "halfling", Faction: "", Profession: "farmer", FactionID: "halfling", DesiredCount: 4, Interval: 0, MinLevel: 1, MaxLevel: 3},
		{ID: "tiefling_city", LocationID: "city_slums", Species: "tiefling", Faction: "", Profession: "warrior", FactionID: "tiefling", DesiredCount: 2, Interval: 0, MinLevel: 1, MaxLevel: 4},
		{ID: "aasimar_shrine", LocationID: "shrine", Species: "aasimar", Faction: "", Profession: "priest", FactionID: "aasimar", DesiredCount: 1, Interval: 0, MinLevel: 3, MaxLevel: 5},
		{ID: "goliath_mountain", LocationID: "mountain_peak", Species: "goliath", Faction: "", Profession: "warrior", FactionID: "goliath", DesiredCount: 2, Interval: 0, MinLevel: 3, MaxLevel: 5},

		// --- Beastfolk & hybrids ---
		{ID: "minotaur_labyrinth", LocationID: "labyrinth", Species: "minotaur", Faction: "", Profession: "warrior", FactionID: "beastfolk", DesiredCount: 2, Interval: 0, MinLevel: 4, MaxLevel: 6},
		{ID: "centaur_plain", LocationID: "open_plains", Species: "centaur", Faction: "", Profession: "warrior", FactionID: "beastfolk", DesiredCount: 3, Interval: 0, MinLevel: 2, MaxLevel: 4},
		{ID: "merfolk_coast", LocationID: "coastal_cave", Species: "merfolk", Faction: "", Profession: "", FactionID: "beastfolk", DesiredCount: 3, Interval: 0, MinLevel: 1, MaxLevel: 3},
		{ID: "harpy_cliff", LocationID: "cliff_nest", Species: "harpy", Faction: "", Profession: "", FactionID: "beastfolk", DesiredCount: 2, Interval: 120, MinLevel: 2, MaxLevel: 4},
		{ID: "werewolf_forest", LocationID: "dark_forest", Species: "werewolf", Faction: "", Profession: "warrior", FactionID: "beastfolk", DesiredCount: 2, Interval: 0, MinLevel: 3, MaxLevel: 5},
		{ID: "werebear_mountain", LocationID: "mountain_cave", Species: "werebear", Faction: "", Profession: "", FactionID: "beastfolk", DesiredCount: 1, Interval: 0, MinLevel: 4, MaxLevel: 6},

		// --- Monstrous ---
		{ID: "bugbear_cave", LocationID: "goblin_hollow", Species: "bugbear", Faction: "", Profession: "warrior", FactionID: "goblin", DesiredCount: 2, Interval: 0, MinLevel: 3, MaxLevel: 5},
		{ID: "ogre_mountain", LocationID: "ogre_stronghold", Species: "ogre", Faction: "", Profession: "warrior", FactionID: "ogre", DesiredCount: 2, Interval: 0, MinLevel: 4, MaxLevel: 6},
		{ID: "troll_bridge", LocationID: "swamp", Species: "troll", Faction: "", Profession: "", FactionID: "beast", DesiredCount: 1, Interval: 0, MinLevel: 5, MaxLevel: 7},
		{ID: "giant_peak", LocationID: "mountain_peak", Species: "giant", Faction: "", Profession: "warrior", FactionID: "giant", DesiredCount: 1, Interval: 0, MinLevel: 6, MaxLevel: 8},
		{ID: "mimic_chest", LocationID: "dungeon", Species: "mimic", Faction: "", Profession: "", FactionID: "monster", DesiredCount: 1, Interval: 0, MinLevel: 3, MaxLevel: 5},
		{ID: "slime_cave", LocationID: "cave_system", Species: "slime", Faction: "", Profession: "", FactionID: "vermin", DesiredCount: 4, Interval: 60, MinLevel: 1, MaxLevel: 2},
		{ID: "golem_ruins", LocationID: "ancient_ruins", Species: "golem", Faction: "", Profession: "", FactionID: "construct", DesiredCount: 1, Interval: 0, MinLevel: 5, MaxLevel: 7},

		// --- Divine ---
		{ID: "deity_shrine", LocationID: "temple", Species: "deity", Faction: "", Profession: "priest", FactionID: "divine", DesiredCount: 1, Interval: 0, MinLevel: 8, MaxLevel: 10},
	}

	// 2. Track which species already have at least one rule.
	covered := make(map[string]bool)
	for _, r := range sm.Rules {
		covered[r.Species] = true
	}

	// 3. Add a default spawn rule for every species not already covered.
	//    This ensures all registered species can appear during initialization,
	//    fixing the issue where only a subset of species were being generated.
	for _, sp := range species.Registry {
		if covered[sp.ID] {
			continue
		}
		// Pick a reasonable default location and faction based on species traits.
		locID := defaultLocationForSpecies(sp.ID)
		factionID := defaultFactionForSpecies(sp.ID)
		profession := defaultProfessionForSpecies(sp.ID)

		// Use a sane default max level instead of MaxAge (which is a lifespan in years, not a level).
		// Without this cap, species like elf (MaxAge=700) would spawn entities at level 700.
		maxLvl := DefaultMaxLevel

		rule := SpawnRule{
			ID:           "default_" + sp.ID,
			LocationID:   locID,
			Species:      sp.ID,
			Faction:      "",
			Profession:   profession,
			FactionID:    factionID,
			DesiredCount: 2,
			Interval:     0, // spawn once at init
			MinLevel:     1,
			MaxLevel:     maxLvl,
		}
		sm.Rules = append(sm.Rules, rule)
		covered[sp.ID] = true
	}

	return sm
}

// defaultLocationForSpecies returns a sensible default location ID for a species.
func defaultLocationForSpecies(speciesID string) string {
	switch speciesID {
	case "orc":
		return "orc_camp"
	case "wolf":
		return "wolf_den"
	case "bear":
		return "bear_den"
	case "boar":
		return "boar_wallow"
	case "rat":
		return "rat_king_lair_entrance"
	case "spider":
		return "spider_grove"
	case "goblin":
		return "goblin_hollow"
	case "kobold":
		return "kobold_warren"
	case "human":
		return "tavern"
	case "dwarf":
		return "dwarf_keep"
	case "elf":
		return "fey_glade"
	case "half_orc":
		return "orc_camp"
	case "half_elf":
		return "fey_glade"
	case "half_dwarf":
		return "dwarf_keep"
	case "half_goblin":
		return "goblin_hollow"
	case "half_hobgoblin":
		return "orc_camp"
	case "half_gnoll":
		return "orc_camp"
	case "half_kobold":
		return "kobold_warren"
	case "half_fey":
		return "fey_glade"
	// Undead
	case "skeleton":
		return "graveyard"
	case "zombie":
		return "castle_dungeon"
	case "ghost":
		return "castle_keep"
	case "wraith":
		return "crypt"
	case "lich":
		return "dark_tower"
	case "vampire":
		return "mansion"
	// Fey & magical
	case "fairy":
		return "fey_glade"
	case "dryad":
		return "ancient_forest"
	case "satyr":
		return "mountain_cave"
	case "pixie":
		return "mushroom_forest"
	case "treant":
		return "ancient_forest"
	// Dragons & reptiles
	case "dragon":
		return "dragon_cave"
	case "lizardfolk":
		return "swamp"
	case "wyvern":
		return "mountain_peak"
	case "basilisk":
		return "cave_system"
	// Small races
	case "gnome":
		return "gnome_hollow"
	case "halfling":
		return "halfling_village"
	case "tiefling":
		return "city_slums"
	case "aasimar":
		return "shrine"
	case "goliath":
		return "mountain_peak"
	// Beastfolk & hybrids
	case "minotaur":
		return "labyrinth"
	case "centaur":
		return "open_plains"
	case "merfolk":
		return "coastal_cave"
	case "harpy":
		return "cliff_nest"
	case "werewolf":
		return "dark_forest"
	case "werebear":
		return "mountain_cave"
	// Monstrous
	case "bugbear":
		return "goblin_hollow"
	case "ogre":
		return "ogre_stronghold"
	case "troll":
		return "swamp"
	case "giant":
		return "mountain_peak"
	case "mimic":
		return "dungeon"
	case "slime":
		return "cave_system"
	case "golem":
		return "ancient_ruins"
	// Divine
	case "deity":
		return "temple"
	// Fallback
	default:
		return "tavern"
	}
}

// defaultFactionForSpecies returns a default faction ID for a species.
func defaultFactionForSpecies(speciesID string) string {
	switch speciesID {
	case "orc", "half_orc", "half_hobgoblin", "half_gnoll":
		return "orc"
	case "wolf", "bear", "boar", "rat", "spider", "goblin", "kobold", "bugbear", "ogre", "troll", "giant", "slime", "mimic":
		return "beast"
	case "human", "half_elf", "half_dwarf", "half_goblin", "half_kobold", "half_fey", "gnome", "halfling", "tiefling", "aasimar", "goliath", "minotaur", "centaur", "merfolk", "harpy", "werewolf", "werebear", "lizardfolk", "basilisk", "dryad", "satyr", "pixie", "treant", "fairy":
		return ""
	case "skeleton", "zombie", "ghost", "wraith", "lich", "vampire":
		return "undead"
	case "dragon":
		return "dragon"
	case "golem", "construct":
		return "construct"
	case "deity":
		return "divine"
	default:
		return ""
	}
}

// defaultProfessionForSpecies returns a default profession for a species.
func defaultProfessionForSpecies(speciesID string) string {
	switch speciesID {
	case "orc", "half_orc", "half_hobgoblin", "half_gnoll", "bugbear", "ogre", "troll", "giant", "minotaur", "centaur", "lizardfolk", "werewolf", "werebear", "goliath":
		return "warrior"
	case "wolf", "bear", "boar", "rat", "spider", "basilisk", "harpy", "slime", "mimic", "golem", "construct":
		return ""
	case "goblin":
		return "gatherer"
	case "kobold", "half_kobold":
		return "warrior"
	case "human", "half_elf", "half_dwarf", "half_goblin", "half_fey", "gnome", "halfling", "tiefling", "aasimar", "dryad", "satyr", "pixie", "treant", "fairy", "merfolk":
		return ""
	case "skeleton", "zombie", "ghost", "wraith":
		return "warrior"
	case "lich":
		return "necromancer"
	case "vampire":
		return "politician"
	case "dragon":
		return ""
	case "deity":
		return "priest"
	default:
		return ""
	}
}

func (sm *SpawnManager) ProcessSpawns(w *world.World, worldEntities *entity.EntityManager, tick int, rng *rand.Rand) {
	for i := range sm.Rules {
		rule := &sm.Rules[i]
		if rule.Interval < 0 {
			continue
		}
		if rule.Interval == 0 {
			if tick != 0 {
				continue
			}
		} else if tick%rule.Interval != 0 {
			continue
		}
		if w != nil {
			loc := w.Location(rule.LocationID)
			if loc != nil && loc.ControllingFaction != "" {
				if rule.RequireFaction != "" && loc.ControllingFaction != rule.RequireFaction {
					continue
				}
				if rule.BlockIfEnemyCtl && loc.ControllingFaction != rule.Faction {
					continue
				}
			}
		}
		current := countAliveAtLocation(worldEntities, rule.LocationID, rule.Faction, rule.Species)
		needed := rule.DesiredCount - current
		if needed <= 0 {
			continue
		}
		for j := 0; j < needed; j++ {
			spawn := spawnEntity(rule, worldEntities, tick, j, rng)
			if spawn != nil {
				log.Printf("[spawn] spawned %s (%s) at %s", spawn.Name, spawn.Species, rule.LocationID)
			}
		}
	}
}

func countAliveAtLocation(em *entity.EntityManager, locID, faction, species string) int {
	count := 0
	for _, e := range em.All() {
		if e.Alive && e.LocationID == locID && e.Faction == faction && e.Species == species {
			count++
		}
	}
	return count
}

func spawnEntity(rule *SpawnRule, em *entity.EntityManager, tick, idx int, rng *rand.Rand) *entity.Entity {
	level := rule.MinLevel
	if rule.MaxLevel > rule.MinLevel {
		level += rng.Intn(rule.MaxLevel - rule.MinLevel + 1)
	}

	// Cap the level to a sane default if the species rule doesn't provide one.
	// MaxAge is a lifespan in years, not a level — using it directly would
	// create wildly overpowered entities at initialization.
	if level > DefaultMaxLevel {
		level = DefaultMaxLevel
	}

	attrs := baseSpeciesAttrs(rule.Species, rng)
	name := generateName(rule.Species, rng)
	id := fmt.Sprintf("%s_spawn_%s_%s_%d_%d", rule.Species, name, rule.LocationID, tick, idx)

	ent := entity.NewEntity(id, name, rule.Species, attrs, level, relation.EmptyRelation)
	ent.LocationID = rule.LocationID
	ent.Faction = rule.Faction
	ent.Profession = rule.Profession
	if sp, exists := species.GetByID(rule.Species); exists && sp.CanReproduce {
		ent.Gender = entity.GetRndGender()
	}
	ent.AI = entity.EntityAI{
		Type:         "scripted",
		ScriptIDs:    defaultScripts(rule.Species, rule.Profession),
		FactionID:    rule.FactionID,
		SleepCycle:   defaultSleepCycle(rule.Species),
		HomeLocation: rule.LocationID,
	}
	equipSpawn(ent, rule, rng)
	ent.XP = randomXPForLevel(level, rng.Intn)
	em.Add(ent)
	return ent
}

func equipSpawnItem(ent *entity.Entity, defID string) {
	def := items.GetDef(defID)
	if def == nil {
		return
	}
	inst := items.NewItemInstance(defID+"_"+ent.ID, defID, def, 1)
	ent.AddItem(inst)
	if def.Slot != "" {
		ent.Equip(&ent.Inventory[len(ent.Inventory)-1])
	}
}

// equipSpawn arms respawned fighters so camps stay dangerous over time.
func equipSpawn(ent *entity.Entity, rule *SpawnRule, rng *rand.Rand) {
	if len(rule.Equipment) > 0 {
		for _, id := range rule.Equipment {
			equipSpawnItem(ent, id)
		}
		return
	}
	switch rule.Species {
	case "orc":
		if ent.Attributes.STR >= 17 {
			equipSpawnItem(ent, "plate_mail")
			equipSpawnItem(ent, "plate_helmet")
			equipSpawnItem(ent, "plate_boots")
			roll := rng.Intn(100)
			switch {
			case roll < 30:
				equipSpawnItem(ent, "great_axe")
			case roll < 60:
				equipSpawnItem(ent, "two_handed_sword")
			default:
				equipSpawnItem(ent, "battle_axe")
			}
		} else if ent.Attributes.STR >= 15 {
			equipSpawnItem(ent, "studded_leather")
			if rng.Intn(100) < 30 {
				equipSpawnItem(ent, "iron_helmet")
			}
			roll := rng.Intn(100)
			switch {
			case roll < 30:
				equipSpawnItem(ent, "orc_cleaver")
			case roll < 55:
				equipSpawnItem(ent, "battle_axe")
			case roll < 75:
				equipSpawnItem(ent, "great_axe")
			default:
				equipSpawnItem(ent, "two_handed_sword")
			}
		} else {
			equipSpawnItem(ent, "leather_armor")
			if rng.Intn(100) < 40 {
				equipSpawnItem(ent, "leather_helmet")
			}
			roll := rng.Intn(100)
			switch {
			case roll < 35:
				equipSpawnItem(ent, "orc_cleaver")
			case roll < 60:
				equipSpawnItem(ent, "hand_axe")
			case roll < 80:
				equipSpawnItem(ent, "iron_spear")
			default:
				equipSpawnItem(ent, "iron_sword")
			}
		}
	case "human":
		if rule.Profession == "bandit" {
			if ent.Attributes.STR >= 16 {
				equipSpawnItem(ent, "plate_mail")
				equipSpawnItem(ent, "plate_boots")
				roll := rng.Intn(100)
				switch {
				case roll < 30:
					equipSpawnItem(ent, "two_handed_sword")
				case roll < 55:
					equipSpawnItem(ent, "great_axe")
				case roll < 75:
					equipSpawnItem(ent, "battle_axe")
				default:
					equipSpawnItem(ent, "longsword")
				}
			} else if ent.Attributes.STR >= 14 {
				equipSpawnItem(ent, "hard_leather_armor")
				equipSpawnItem(ent, "leather_boots")
				roll := rng.Intn(100)
				switch {
				case roll < 30:
					equipSpawnItem(ent, "longsword")
				case roll < 55:
					equipSpawnItem(ent, "battle_axe")
				case roll < 75:
					equipSpawnItem(ent, "mace")
				default:
					equipSpawnItem(ent, "two_handed_sword")
				}
			} else {
				equipSpawnItem(ent, "leather_armor")
				equipSpawnItem(ent, "leather_boots")
				roll := rng.Intn(100)
				switch {
				case roll < 30:
					equipSpawnItem(ent, "short_sword")
				case roll < 55:
					equipSpawnItem(ent, "hand_axe")
				case roll < 75:
					equipSpawnItem(ent, "hunting_knife")
				default:
					equipSpawnItem(ent, "cudgel")
				}
			}
		}
		if rule.Profession == "bard" {
			equipSpawnItem(ent, "work_tunic")
			equipSpawnItem(ent, "lute")
		}
		if rule.Profession == "priest" {
			equipSpawnItem(ent, "robes")
			equipSpawnItem(ent, "holy_symbol")
		}
	case "goblin":
		equipSpawnItem(ent, "work_tunic")
		if rng.Intn(100) < 60 {
			equipSpawnItem(ent, "goblin_shiv")
		} else {
			equipSpawnItem(ent, "cudgel")
		}
	case "kobold":
		equipSpawnItem(ent, "work_tunic")
		roll := rng.Intn(100)
		switch {
		case roll < 55:
			equipSpawnItem(ent, "dagger")
		case roll < 80:
			equipSpawnItem(ent, "short_sword")
		default:
			equipSpawnItem(ent, "goblin_shiv")
		}
	case "wolf", "bear", "werewolf", "werebear":
		equipSpawnItem(ent, "claws")
	case "boar":
		equipSpawnItem(ent, "tusks")
	case "spider", "rat", "slime":
		equipSpawnItem(ent, "fangs")
	case "dragon":
		equipSpawnItem(ent, "dragon_breath")
	case "ogre":
		equipSpawnItem(ent, "great_club")
	case "troll":
		equipSpawnItem(ent, "club")
	case "giant":
		equipSpawnItem(ent, "boulder")
	case "minotaur":
		equipSpawnItem(ent, "axe")
	case "golem":
		equipSpawnItem(ent, "stone_fist")
	case "bugbear":
		equipSpawnItem(ent, "morningstar")
	case "harpy":
		equipSpawnItem(ent, "talons")
	case "centaur":
		equipSpawnItem(ent, "short_sword")
	case "merfolk":
		equipSpawnItem(ent, "trident")
	case "skeleton":
		equipSpawnItem(ent, "rusty_sword")
	case "zombie":
		equipSpawnItem(ent, "rusty_axe")
	case "ghost":
		equipSpawnItem(ent, "ethereal_touch")
	case "wraith":
		equipSpawnItem(ent, "soul_scythe")
	case "lich":
		equipSpawnItem(ent, "staff_of_death")
	case "vampire":
		equipSpawnItem(ent, "fangs")
	case "fairy":
		equipSpawnItem(ent, "wand")
	case "dryad":
		equipSpawnItem(ent, "staff_of_vines")
	case "satyr":
		equipSpawnItem(ent, "pan_flute")
	case "pixie":
		equipSpawnItem(ent, "tiny_dagger")
	case "treant":
		equipSpawnItem(ent, "branch")
	case "gnome":
		equipSpawnItem(ent, "pickaxe")
	case "halfling":
		equipSpawnItem(ent, "short_sword")
	case "tiefling":
		equipSpawnItem(ent, "horn")
	case "aasimar":
		equipSpawnItem(ent, "holy_symbol")
	case "goliath":
		equipSpawnItem(ent, "greatclub")
	case "lizardfolk":
		equipSpawnItem(ent, "spear")
	case "basilisk":
		equipSpawnItem(ent, "petrifying_gaze")
	case "mimic":
		equipSpawnItem(ent, "bite")
	case "deity":
		equipSpawnItem(ent, "divine_scepter")
	}
}

func baseSpeciesAttrs(species string, rng *rand.Rand) entity.Attributes {
	switch species {
	case "orc":
		return entity.Attributes{STR: 14 + rng.Intn(4), DEX: 10 + rng.Intn(3), CON: 13 + rng.Intn(3), INT: 6 + rng.Intn(3), WIS: 6 + rng.Intn(3), CHA: 5 + rng.Intn(3)}
	case "wolf":
		return entity.Attributes{STR: 12 + rng.Intn(4), DEX: 14 + rng.Intn(3), CON: 11 + rng.Intn(3), INT: 3 + rng.Intn(2), WIS: 7 + rng.Intn(3), CHA: 3 + rng.Intn(2)}
	case "bear":
		return entity.Attributes{STR: 18 + rng.Intn(3), DEX: 9 + rng.Intn(3), CON: 16 + rng.Intn(3), INT: 2 + rng.Intn(2), WIS: 6 + rng.Intn(3), CHA: 2 + rng.Intn(2)}
	case "boar":
		return entity.Attributes{STR: 14 + rng.Intn(3), DEX: 11 + rng.Intn(3), CON: 13 + rng.Intn(3), INT: 2 + rng.Intn(2), WIS: 5 + rng.Intn(3), CHA: 2 + rng.Intn(2)}
	case "rat":
		return entity.Attributes{STR: 6 + rng.Intn(3), DEX: 12 + rng.Intn(3), CON: 8 + rng.Intn(3), INT: 2 + rng.Intn(2), WIS: 5 + rng.Intn(3), CHA: 2 + rng.Intn(2)}
	case "spider":
		return entity.Attributes{STR: 10 + rng.Intn(3), DEX: 15 + rng.Intn(4), CON: 8 + rng.Intn(3), INT: 2 + rng.Intn(2), WIS: 7 + rng.Intn(3), CHA: 2 + rng.Intn(2)}
	case "goblin":
		return entity.Attributes{STR: 8 + rng.Intn(3), DEX: 12 + rng.Intn(3), CON: 10 + rng.Intn(3), INT: 8 + rng.Intn(3), WIS: 6 + rng.Intn(3), CHA: 6 + rng.Intn(3)}
	case "kobold":
		return entity.Attributes{STR: 8 + rng.Intn(3), DEX: 14 + rng.Intn(3), CON: 9 + rng.Intn(3), INT: 8 + rng.Intn(3), WIS: 7 + rng.Intn(3), CHA: 6 + rng.Intn(3)}
	case "half_orc":
		return entity.Attributes{STR: 13 + rng.Intn(4), DEX: 10 + rng.Intn(3), CON: 12 + rng.Intn(3), INT: 6 + rng.Intn(3), WIS: 6 + rng.Intn(3), CHA: 5 + rng.Intn(3)}
	case "half_elf":
		return entity.Attributes{STR: 10 + rng.Intn(3), DEX: 11 + rng.Intn(3), CON: 10 + rng.Intn(3), INT: 10 + rng.Intn(3), WIS: 10 + rng.Intn(3), CHA: 11 + rng.Intn(3)}
	case "half_dwarf":
		return entity.Attributes{STR: 12 + rng.Intn(3), DEX: 10 + rng.Intn(3), CON: 14 + rng.Intn(3), INT: 8 + rng.Intn(3), WIS: 9 + rng.Intn(3), CHA: 8 + rng.Intn(3)}
	case "half_goblin":
		return entity.Attributes{STR: 9 + rng.Intn(3), DEX: 13 + rng.Intn(3), CON: 10 + rng.Intn(3), INT: 9 + rng.Intn(3), WIS: 7 + rng.Intn(3), CHA: 7 + rng.Intn(3)}
	case "half_hobgoblin":
		return entity.Attributes{STR: 13 + rng.Intn(3), DEX: 11 + rng.Intn(3), CON: 12 + rng.Intn(3), INT: 7 + rng.Intn(3), WIS: 7 + rng.Intn(3), CHA: 6 + rng.Intn(3)}
	case "half_gnoll":
		return entity.Attributes{STR: 12 + rng.Intn(3), DEX: 11 + rng.Intn(3), CON: 11 + rng.Intn(3), INT: 5 + rng.Intn(3), WIS: 6 + rng.Intn(3), CHA: 5 + rng.Intn(3)}
	case "half_kobold":
		return entity.Attributes{STR: 9 + rng.Intn(3), DEX: 13 + rng.Intn(3), CON: 9 + rng.Intn(3), INT: 9 + rng.Intn(3), WIS: 8 + rng.Intn(3), CHA: 7 + rng.Intn(3)}
	case "half_fey":
		return entity.Attributes{STR: 8 + rng.Intn(3), DEX: 13 + rng.Intn(3), CON: 9 + rng.Intn(3), INT: 11 + rng.Intn(3), WIS: 12 + rng.Intn(3), CHA: 13 + rng.Intn(3)}
	// Undead
	case "skeleton":
		return entity.Attributes{STR: 10 + rng.Intn(3), DEX: 10 + rng.Intn(3), CON: 8 + rng.Intn(3), INT: 3 + rng.Intn(2), WIS: 4 + rng.Intn(2), CHA: 2 + rng.Intn(2)}
	case "zombie":
		return entity.Attributes{STR: 12 + rng.Intn(3), DEX: 6 + rng.Intn(2), CON: 12 + rng.Intn(3), INT: 2 + rng.Intn(2), WIS: 3 + rng.Intn(2), CHA: 2 + rng.Intn(2)}
	case "ghost":
		return entity.Attributes{STR: 6 + rng.Intn(2), DEX: 14 + rng.Intn(3), CON: 6 + rng.Intn(2), INT: 10 + rng.Intn(3), WIS: 12 + rng.Intn(3), CHA: 10 + rng.Intn(3)}
	case "wraith":
		return entity.Attributes{STR: 8 + rng.Intn(2), DEX: 15 + rng.Intn(3), CON: 7 + rng.Intn(2), INT: 12 + rng.Intn(3), WIS: 14 + rng.Intn(3), CHA: 11 + rng.Intn(3)}
	case "lich":
		return entity.Attributes{STR: 10 + rng.Intn(3), DEX: 12 + rng.Intn(3), CON: 10 + rng.Intn(3), INT: 18 + rng.Intn(4), WIS: 16 + rng.Intn(4), CHA: 14 + rng.Intn(3)}
	case "vampire":
		return entity.Attributes{STR: 12 + rng.Intn(3), DEX: 14 + rng.Intn(3), CON: 11 + rng.Intn(3), INT: 10 + rng.Intn(3), WIS: 12 + rng.Intn(3), CHA: 13 + rng.Intn(3)}
	// Fey & magical
	case "fairy":
		return entity.Attributes{STR: 4 + rng.Intn(2), DEX: 16 + rng.Intn(4), CON: 6 + rng.Intn(2), INT: 10 + rng.Intn(3), WIS: 12 + rng.Intn(3), CHA: 14 + rng.Intn(3)}
	case "dryad":
		return entity.Attributes{STR: 8 + rng.Intn(2), DEX: 12 + rng.Intn(3), CON: 10 + rng.Intn(3), INT: 12 + rng.Intn(3), WIS: 14 + rng.Intn(3), CHA: 13 + rng.Intn(3)}
	case "satyr":
		return entity.Attributes{STR: 10 + rng.Intn(3), DEX: 13 + rng.Intn(3), CON: 10 + rng.Intn(3), INT: 9 + rng.Intn(3), WIS: 10 + rng.Intn(3), CHA: 12 + rng.Intn(3)}
	case "pixie":
		return entity.Attributes{STR: 3 + rng.Intn(2), DEX: 17 + rng.Intn(4), CON: 5 + rng.Intn(2), INT: 11 + rng.Intn(3), WIS: 13 + rng.Intn(3), CHA: 15 + rng.Intn(3)}
	case "treant":
		return entity.Attributes{STR: 18 + rng.Intn(3), DEX: 6 + rng.Intn(2), CON: 18 + rng.Intn(3), INT: 10 + rng.Intn(3), WIS: 14 + rng.Intn(3), CHA: 8 + rng.Intn(3)}
	// Dragons & reptiles
	case "dragon":
		return entity.Attributes{STR: 18 + rng.Intn(4), DEX: 12 + rng.Intn(3), CON: 16 + rng.Intn(4), INT: 14 + rng.Intn(4), WIS: 16 + rng.Intn(4), CHA: 16 + rng.Intn(4)}
	case "lizardfolk":
		return entity.Attributes{STR: 13 + rng.Intn(3), DEX: 11 + rng.Intn(3), CON: 12 + rng.Intn(3), INT: 8 + rng.Intn(3), WIS: 9 + rng.Intn(3), CHA: 7 + rng.Intn(3)}
	case "basilisk":
		return entity.Attributes{STR: 12 + rng.Intn(3), DEX: 10 + rng.Intn(3), CON: 12 + rng.Intn(3), INT: 8 + rng.Intn(3), WIS: 10 + rng.Intn(3), CHA: 6 + rng.Intn(3)}
	// Small races
	case "gnome":
		return entity.Attributes{STR: 7 + rng.Intn(2), DEX: 11 + rng.Intn(3), CON: 9 + rng.Intn(2), INT: 14 + rng.Intn(4), WIS: 12 + rng.Intn(3), CHA: 10 + rng.Intn(3)}
	case "halfling":
		return entity.Attributes{STR: 8 + rng.Intn(2), DEX: 14 + rng.Intn(3), CON: 9 + rng.Intn(2), INT: 10 + rng.Intn(3), WIS: 11 + rng.Intn(3), CHA: 10 + rng.Intn(3)}
	case "tiefling":
		return entity.Attributes{STR: 10 + rng.Intn(3), DEX: 12 + rng.Intn(3), CON: 10 + rng.Intn(3), INT: 11 + rng.Intn(3), WIS: 10 + rng.Intn(3), CHA: 12 + rng.Intn(3)}
	case "aasimar":
		return entity.Attributes{STR: 10 + rng.Intn(3), DEX: 11 + rng.Intn(3), CON: 10 + rng.Intn(3), INT: 12 + rng.Intn(3), WIS: 14 + rng.Intn(3), CHA: 13 + rng.Intn(3)}
	case "goliath":
		return entity.Attributes{STR: 16 + rng.Intn(4), DEX: 10 + rng.Intn(3), CON: 15 + rng.Intn(3), INT: 8 + rng.Intn(3), WIS: 10 + rng.Intn(3), CHA: 8 + rng.Intn(3)}
	// Beastfolk & hybrids
	case "minotaur":
		return entity.Attributes{STR: 16 + rng.Intn(3), DEX: 10 + rng.Intn(3), CON: 14 + rng.Intn(3), INT: 6 + rng.Intn(3), WIS: 8 + rng.Intn(3), CHA: 6 + rng.Intn(3)}
	case "centaur":
		return entity.Attributes{STR: 14 + rng.Intn(3), DEX: 13 + rng.Intn(3), CON: 12 + rng.Intn(3), INT: 8 + rng.Intn(3), WIS: 10 + rng.Intn(3), CHA: 8 + rng.Intn(3)}
	case "merfolk":
		return entity.Attributes{STR: 11 + rng.Intn(3), DEX: 13 + rng.Intn(3), CON: 11 + rng.Intn(3), INT: 10 + rng.Intn(3), WIS: 12 + rng.Intn(3), CHA: 11 + rng.Intn(3)}
	case "harpy":
		return entity.Attributes{STR: 8 + rng.Intn(2), DEX: 16 + rng.Intn(3), CON: 8 + rng.Intn(2), INT: 9 + rng.Intn(3), WIS: 11 + rng.Intn(3), CHA: 12 + rng.Intn(3)}
	// Monstrous
	case "bugbear":
		return entity.Attributes{STR: 14 + rng.Intn(3), DEX: 12 + rng.Intn(3), CON: 11 + rng.Intn(3), INT: 7 + rng.Intn(3), WIS: 8 + rng.Intn(3), CHA: 7 + rng.Intn(3)}
	case "ogre":
		return entity.Attributes{STR: 18 + rng.Intn(3), DEX: 7 + rng.Intn(2), CON: 16 + rng.Intn(3), INT: 5 + rng.Intn(2), WIS: 6 + rng.Intn(2), CHA: 5 + rng.Intn(2)}
	case "troll":
		return entity.Attributes{STR: 16 + rng.Intn(3), DEX: 8 + rng.Intn(2), CON: 14 + rng.Intn(3), INT: 5 + rng.Intn(2), WIS: 6 + rng.Intn(2), CHA: 5 + rng.Intn(2)}
	case "giant":
		return entity.Attributes{STR: 18 + rng.Intn(4), DEX: 8 + rng.Intn(3), CON: 16 + rng.Intn(4), INT: 7 + rng.Intn(3), WIS: 9 + rng.Intn(3), CHA: 7 + rng.Intn(3)}
	case "mimic":
		return entity.Attributes{STR: 10 + rng.Intn(3), DEX: 8 + rng.Intn(2), CON: 12 + rng.Intn(3), INT: 6 + rng.Intn(2), WIS: 7 + rng.Intn(2), CHA: 4 + rng.Intn(2)}
	case "golem":
		return entity.Attributes{STR: 14 + rng.Intn(3), DEX: 6 + rng.Intn(2), CON: 16 + rng.Intn(4), INT: 4 + rng.Intn(2), WIS: 6 + rng.Intn(2), CHA: 3 + rng.Intn(2)}
	// Divine
	case "deity":
		return entity.Attributes{STR: 16 + rng.Intn(4), DEX: 14 + rng.Intn(4), CON: 14 + rng.Intn(4), INT: 18 + rng.Intn(4), WIS: 18 + rng.Intn(4), CHA: 18 + rng.Intn(4)}
	// Fallback
	default:
		return entity.RandomAttributes(func(n int) int { return rng.Intn(n) })
	}
}

func generateName(species string, rng *rand.Rand) string {
	orcNames := []string{"Mog", "Zog", "Thrak", "Gruul", "Drok", "Krag", "Snag", "Ruk"}
	wolfNames := []string{"Howl", "Rip", "Claw", "Fang", "Snap", "Growl"}
	bearNames := []string{"Claw", "Grunt", "Maw", "Huff"}
	boarNames := []string{"Snout", "Tusk", "Grunta", "Bristle"}
	ratNames := []string{"Squeak", "Nibble", "Skitter", "Dart", "Pip"}
	spiderNames := []string{"Legs", "Weaver", "Sting", "Crawl"}
	goblinNames := []string{"Snag", "Grib", "Nog", "Blink", "Mug"}
	koboldNames := []string{"Skrit", "Yip", "Klik", "Drak", "Snik"}
	humanNames := []string{"Aldric", "Brenna", "Cedric", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jenna", "Kol", "Lyssa", "Maren", "Nolan", "Opal", "Petra", "Quinn", "Rhea", "Soren", "Tessa"}
	halfOrcNames := []string{"Grolf", "Krom", "Thok", "Brak", "Drog", "Mog", "Urgal", "Garok", "Krelka", "Draga", "Shaka", "Greta", "Lara", "Orsha", "Bruna", "Hurga"}
	halfElfNames := []string{"Aldric", "Cedric", "Eamon", "Gareth", "Thalor", "Elric", "Caelum", "Fenrik", "Aldrica", "Brenna", "Delara", "Fiona", "Thalyra", "Elyra", "Caelia", "Riven"}
	halfDwarfNames := []string{"Dorn", "Brik", "Haldor", "Grund", "Torin", "Baldur", "Fjor", "Erik", "Dessa", "Brynn", "Hilda", "Greta", "Torva", "Barda", "Fjora", "Astrid"}
	halfGoblinNames := []string{"Snag", "Grib", "Nog", "Blink", "Mog", "Rik", "Wix", "Zep", "Grix", "Nix", "Plix", "Wyna", "Bria", "Mika", "Trix", "Flick"}
	halfHobgoblinNames := []string{"Durgath", "Skarrak", "Mograth", "Karg", "Brak", "Zol", "Vorn", "Ghrak", "Krel", "Shara", "Nixa", "Greta", "Velka", "Dasha", "Kira", "Zara"}
	halfGnollNames := []string{"Ripper", "Bonepick", "Snapper", "Gorr", "Ashclaw", "Maw", "Vex", "Grak", "Ripsnout", "Mama", "Vexa", "Grix", "Bonea", "Snapa", "Graw", "Krela"}
	halfKoboldNames := []string{"Skrit", "Yip", "Klik", "Drak", "Snik", "Rix", "Zik", "Vrik", "Skrix", "Yipa", "Klika", "Draka", "Snika", "Rika", "Zika", "Vrika"}
	halfFeyNames := []string{"Thorn", "Bram", "Alder", "Rowan", "Briar", "Fenn", "Oaken", "Willow", "Thyra", "Briar", "Alda", "Rowan", "Nyx", "Luma", "Sylph", "Faye"}
	// Undead
	skeletonNames := []string{"Bone", "Rattle", "Skel", "Mort", "Ash", "Grave", "Rust", "Hollow"}
	zombieNames := []string{"Rot", "Decay", "Corpse", "Shambler", "Ghoul", "Wretch", "Fester", "Blight"}
	ghostNames := []string{"Wisp", "Shade", "Specter", "Ethereal", "Phantom", "Wraith", "Banshee", "Apparition"}
	wraithNames := []string{"Dread", "Murk", "Gloom", "Shadow", "Void", "Eclipse", "Night", "Dusk"}
	lichNames := []string{"Necros", "Mortis", "Kael", "Xaren", "Velthar", "Zargoth", "Malachar", "Thrain"}
	vampireNames := []string{"Vlad", "Dracula", "Noctis", "Sanguis", "Morven", "Lysandra", "Kaelith", "Valerius"}
	// Fey & magical
	fairyNames := []string{"Tinker", "Glimmer", "Dewdrop", "Flicker", "Petal", "Moth", "Starlight", "Zephyr"}
	dryadNames := []string{"Aurora", "Sylva", "Thorn", "Briar", "Moss", "Fern", "Willow", "Ivy"}
	satyrNames := []string{"Pan", "Silenus", "Dion", "Lycus", "Phere", "Crotus", "Marsyas", "Oreas"}
	pixieNames := []string{"Tinker", "Flick", "Glimmer", "Dew", "Petal", "Moth", "Wisp", "Dust"}
	treantNames := []string{"Oldgrowth", "Deeproot", "Barkheart", "Thornbeard", "Greenmantle", "Rootwalker", "Timber", "Oakheart"}
	// Dragons & reptiles
	dragonNames := []string{"Smaug", "Vermithrax", "Draco", "Pyroth", "Frostclaw", "Stormwing", "Ember", "Shadowscale"}
	lizardfolkNames := []string{"Scales", "Thornscale", "Riptide", "Swampscale", "Coldscale", "Duskscale", "Brightscale", "Fangjaw"}
	basiliskNames := []string{"Stonegaze", "Petra", "Gorgon", "Serpentis", "Duskfang", "Coil", "Slither", "Basil"}
	// Small races
	gnomeNames := []string{"Tinker", "Gizmo", "Blix", "Zep", "Flick", "Dust", "Pip", "Nix"}
	halflingNames := []string{"Bravo", "Daisy", "Pippin", "Rosie", "Sam", "Nim", "Lottie", "Jory"}
	tieflingNames := []string{"Zariel", "Mephisto", "Asmodeus", "Fierna", "Glasya", "Lilith", "Baal", "Moloch"}
	aasimarNames := []string{"Auriel", "Celestine", "Seren", "Lumina", "Divine", "Radiant", "Seraph", "Healer"}
	goliathNames := []string{"Korg", "Brak", "Thok", "Dorn", "Haldor", "Grun", "Baldur", "Fjor"}
	// Beastfolk & hybrids
	minotaurNames := []string{"Asterion", "Minotaur", "Brawn", "Horn", "Gore", "Thorn", "Maze", "Labyrinth"}
	centaurNames := []string{"Chiron", "Bolt", "Gallop", "Swift", "Prowl", "Stripe", "Hoof", "Rush"}
	merfolkNames := []string{"Coral", "Tide", "Wave", "Splash", "Fin", "Shell", "Pearl", "Deep"}
	harpyNames := []string{"Screech", "Wing", "Gale", "Storm", "Razor", "Plume", "Talons", "Squawk"}
	werewolfNames := []string{"Fang", "Howl", "Rex", "Luna", "Shadow", "Feral", "Claw", "Prowl"}
	werebearNames := []string{"Grizz", "Claw", "Fang", "Roar", "Brawn", "Thorn", "Paw", "Maw"}
	// Monstrous
	bugbearNames := []string{"Gruk", "Skarn", "Thok", "Brak", "Grix", "Nix", "Vorn", "Ghrak"}
	ogreNames := []string{"Shrek", "Grond", "Thud", "Bash", "Crush", "Maul", "Grun", "Thok"}
	trollNames := []string{"Stone", "Rot", "Thud", "Grun", "Mud", "Bog", "Tusk", "Claw"}
	giantNames := []string{"Colossus", "Titan", "Boulder", "Thorn", "Gronn", "Dwarf", "Ogre", "Troll"}
	mimicNames := []string{"Chest", "Trap", "Mimic", "Shapeshifter", "Decoy", "False", "Trick", "Snare"}
	slimeNames := []string{"Ooze", "Goo", "Slime", "Muck", "Drip", "Splat", "Bloop", "Squish"}
	golemNames := []string{"Iron", "Stone", "Clay", "Metal", "Construct", "Forge", "Anvil", "Shard"}
	// Divine
	deityNames := []string{"Aurora", "Solaris", "Lunara", "Terra", "Ignis", "Aqua", "Aether", "Nyx"}

	names := map[string][]string{
		"orc": orcNames, "wolf": wolfNames, "bear": bearNames,
		"boar": boarNames, "rat": ratNames, "spider": spiderNames, "goblin": goblinNames,
		"kobold": koboldNames, "human": humanNames,
		"half_orc": halfOrcNames, "half_elf": halfElfNames, "half_dwarf": halfDwarfNames,
		"half_goblin": halfGoblinNames, "half_hobgoblin": halfHobgoblinNames,
		"half_gnoll": halfGnollNames, "half_kobold": halfKoboldNames, "half_fey": halfFeyNames,
		// Undead
		"skeleton": skeletonNames, "zombie": zombieNames, "ghost": ghostNames,
		"wraith": wraithNames, "lich": lichNames, "vampire": vampireNames,
		// Fey & magical
		"fairy": fairyNames, "dryad": dryadNames, "satyr": satyrNames,
		"pixie": pixieNames, "treant": treantNames,
		// Dragons & reptiles
		"dragon": dragonNames, "lizardfolk": lizardfolkNames, "basilisk": basiliskNames,
		// Small races
		"gnome": gnomeNames, "halfling": halflingNames, "tiefling": tieflingNames,
		"aasimar": aasimarNames, "goliath": goliathNames,
		// Beastfolk & hybrids
		"minotaur": minotaurNames, "centaur": centaurNames, "merfolk": merfolkNames,
		"harpy": harpyNames, "werewolf": werewolfNames, "werebear": werebearNames,
		// Monstrous
		"bugbear": bugbearNames, "ogre": ogreNames, "troll": trollNames,
		"giant": giantNames, "mimic": mimicNames, "slime": slimeNames,
		"golem": golemNames,
		// Divine
		"deity": deityNames,
	}
	pool, ok := names[species]
	if !ok || len(pool) == 0 {
		return species + "_spawn"
	}
	return pool[rng.Intn(len(pool))]
}

// scriptPriority returns a lower number for scripts that should run first.
// Survival/behavioral scripts run before profession scripts, which run before
// species-specific scripts.
func scriptPriority(name string) int {
	switch name {
	case "defensive", "aggressive", "hunting", "gathering", "scouting", "healing":
		return 0
	case "bard", "guard", "ranger", "priest", "farmer", "fisherman", "miner",
		"blacksmith", "innkeeper", "herbalist", "courier", "thief", "cultist",
		"traveling_salesman", "wizard", "bar_patron", "bandit_chief",
		"bread_weaver", "necromancer", "bandit", "politician", "diplomat",
		"berzerker":
		return 1
	default:
		return 2
	}
}

func defaultScripts(sp, profession string) []string {
	var scripts []string
	seen := map[string]bool{}

	// 1. Start with species DefaultScripts from the registry.
	if s, ok := species.GetByID(sp); ok {
		scripts = append(scripts, s.DefaultScripts...)
	}

	// 2. Add profession script if one exists for this profession.
	profScript := professionScript(profession)
	if profScript != "" && !seen[profScript] {
		scripts = append(scripts, profScript)
	}

	// 3. Deduplicate and add fallback.
	if len(scripts) == 0 {
		return []string{"aggressive"}
	}
	deduped := scripts[:0]
	for _, s := range scripts {
		if s != "" && !seen[s] {
			seen[s] = true
			deduped = append(deduped, s)
		}
	}
	if len(deduped) == 0 {
		return []string{"aggressive"}
	}

	// 4. Sort by priority: survival → profession → species.
	slices.SortStableFunc(deduped, func(a, b string) int {
		return scriptPriority(a) - scriptPriority(b)
	})
	return deduped
}

// professionScript maps a profession name to its Lua script ID.
func professionScript(profession string) string {
	switch profession {
	case "bard":
		return "bard"
	case "guard":
		return "guard"
	case "ranger":
		return "ranger"
	case "priest":
		return "priest"
	case "farmer":
		return "farmer"
	case "fisherman":
		return "fisherman"
	case "miner":
		return "miner"
	case "blacksmith":
		return "blacksmith"
	case "innkeeper":
		return "innkeeper"
	case "herbalist":
		return "herbalist"
	case "courier":
		return "courier"
	case "thief":
		return "thief"
	case "cultist":
		return "cultist"
	case "merchant", "traveling_salesman":
		return "traveling_salesman"
	case "wizard":
		return "wizard"
	case "necromancer":
		return "necromancer"
	case "politician":
		return "politician"
	case "berzerker":
		return "berzerker"
	case "druid":
		return "druid"
	case "scout":
		return "scout"
	default:
		return ""
	}
}

func defaultSleepCycle(species string) string {
	switch species {
	case "spider":
		return "nocturnal"
	case "ghost", "wraith", "lich", "vampire", "skeleton", "zombie":
		return "undead"
	case "fairy", "pixie", "dryad", "satyr", "treant":
		return "fey"
	default:
		return "diurnal"
	}
}

// randomXPForLevel returns a random XP value for a given level.
// The XP is randomized from 0 up to (but not including) the level-up threshold,
// so the entity stays at its intended level.
func randomXPForLevel(level int, rng func(int) int) int {
	if level <= 0 {
		return 0
	}
	return rng(level * 100)
}
