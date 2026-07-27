package engine

import (
	"fmt"
	"log"
	"math/rand"

	"simuz/internal/entity"
	"simuz/internal/items"
	"simuz/internal/world"
)

type SpawnRule struct {
	ID              string
	LocationID      string
	Species         string
	Faction         string
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

func NewSpawnManager() *SpawnManager {
	return &SpawnManager{
		Rules: []SpawnRule{
			{ID: "orc_patrol", LocationID: "orc_camp", Species: "orc", Faction: "orc", FactionID: "orc", DesiredCount: 4, Interval: 120, MinLevel: 1, MaxLevel: 3, RequireFaction: "orc"},
			{ID: "wolf_pack", LocationID: "wolf_den", Species: "wolf", Faction: "beast", FactionID: "beast", DesiredCount: 4, Interval: 90, MinLevel: 1, MaxLevel: 2},
			{ID: "bandit_camp", LocationID: "bandit_camp", Species: "human", Faction: "bandit", FactionID: "bandit", DesiredCount: 4, Interval: 150, MinLevel: 1, MaxLevel: 2, RequireFaction: "bandit"},
			{ID: "bear_den", LocationID: "bear_den", Species: "bear", Faction: "beast", FactionID: "beast", DesiredCount: 2, Interval: 200, MinLevel: 3, MaxLevel: 5},
			{ID: "boar_herd", LocationID: "boar_wallow", Species: "boar", Faction: "beast", FactionID: "beast", DesiredCount: 2, Interval: 180, MinLevel: 1, MaxLevel: 3},
			{ID: "rat_infest", LocationID: "rat_king_lair_entrance", Species: "rat", Faction: "vermin", FactionID: "vermin", DesiredCount: 3, Interval: 60, MinLevel: 1, MaxLevel: 1},
			{ID: "rat_corridor", LocationID: "rat_king_lair_corridor", Species: "rat", Faction: "vermin", FactionID: "vermin", DesiredCount: 3, Interval: 60, MinLevel: 1, MaxLevel: 2},
			{ID: "spider_nest", LocationID: "spider_grove", Species: "spider", Faction: "beast", FactionID: "beast", DesiredCount: 2, Interval: 120, MinLevel: 1, MaxLevel: 3},
			{ID: "goblin_gatherers", LocationID: "goblin_hollow", Species: "goblin", Faction: "goblin", FactionID: "goblin", DesiredCount: 2, Interval: 180, MinLevel: 1, MaxLevel: 1, RequireFaction: "goblin"},
			{ID: "kobold_warren", LocationID: "kobold_warren", Species: "kobold", Faction: "kobold", FactionID: "kobold", DesiredCount: 4, Interval: 150, MinLevel: 1, MaxLevel: 2, RequireFaction: "kobold"},
			{ID: "ash_scorpions", LocationID: "scorpion_dunes", Species: "spider", Faction: "beast", FactionID: "beast", DesiredCount: 2, Interval: 160, MinLevel: 2, MaxLevel: 4},
			{ID: "ash_orcs", LocationID: "ash_ruins", Species: "orc", Faction: "orc", FactionID: "orc", DesiredCount: 2, Interval: 180, MinLevel: 2, MaxLevel: 4, RequireFaction: "orc"},
		},
	}
}

func (sm *SpawnManager) ProcessSpawns(w *world.World, worldEntities *entity.Manager, tick int, rng *rand.Rand) {
	for i := range sm.Rules {
		rule := &sm.Rules[i]
		if tick%rule.Interval != 0 {
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

func countAliveAtLocation(em *entity.Manager, locID, faction, species string) int {
	count := 0
	for _, e := range em.All() {
		if e.Alive && e.LocationID == locID && e.Faction == faction && e.Species == species {
			count++
		}
	}
	return count
}

func spawnEntity(rule *SpawnRule, em *entity.Manager, tick, idx int, rng *rand.Rand) *entity.Entity {
	level := rule.MinLevel
	if rule.MaxLevel > rule.MinLevel {
		level += rng.Intn(rule.MaxLevel - rule.MinLevel + 1)
	}
	attrs := baseSpeciesAttrs(rule.Species, rng)
	name := generateName(rule.Species, rng)
	id := fmt.Sprintf("%s_spawn_%s_%s_%d_%d", rule.Species, name, rule.LocationID, tick, idx)

	ent := entity.NewEntity(id, name, rule.Species, attrs, level)
	ent.LocationID = rule.LocationID
	ent.Faction = rule.Faction
	ent.AI = entity.EntityAI{
		Type:         "scripted",
		ScriptIDs:    defaultScripts(rule.Species),
		FactionID:    rule.FactionID,
		SleepCycle:   defaultSleepCycle(rule.Species),
		HomeLocation: rule.LocationID,
	}
	equipSpawn(ent, rule, rng)
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
		equipSpawnItem(ent, "leather_armor")
		if rng.Intn(100) < 40 {
			equipSpawnItem(ent, "leather_helmet")
		}
		roll := rng.Intn(100)
		switch {
		case roll < 35:
			equipSpawnItem(ent, "orc_cleaver")
		case roll < 60:
			equipSpawnItem(ent, "iron_axe")
		case roll < 80:
			equipSpawnItem(ent, "iron_spear")
		default:
			equipSpawnItem(ent, "iron_sword")
		}
	case "human":
		if rule.Faction == "bandit" {
			equipSpawnItem(ent, "leather_armor")
			equipSpawnItem(ent, "leather_boots")
			roll := rng.Intn(100)
			switch {
			case roll < 40:
				equipSpawnItem(ent, "iron_sword")
			case roll < 65:
				equipSpawnItem(ent, "short_sword")
			case roll < 85:
				equipSpawnItem(ent, "iron_axe")
			default:
				equipSpawnItem(ent, "iron_spear")
			}
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
	case "wolf", "bear":
		equipSpawnItem(ent, "claws")
	case "boar":
		equipSpawnItem(ent, "tusks")
	case "spider", "rat":
		equipSpawnItem(ent, "fangs")
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
	names := map[string][]string{
		"orc": orcNames, "wolf": wolfNames, "bear": bearNames,
		"boar": boarNames, "rat": ratNames, "spider": spiderNames, "goblin": goblinNames,
		"kobold": koboldNames,
	}
	pool, ok := names[species]
	if !ok || len(pool) == 0 {
		return species + "_spawn"
	}
	return pool[rng.Intn(len(pool))]
}

func defaultScripts(species string) []string {
	switch species {
	case "orc":
		return []string{"aggressive"}
	case "wolf":
		return []string{"hunting"}
	case "bear", "boar":
		return []string{"aggressive"}
	case "rat":
		return []string{"defensive"}
	case "spider":
		return []string{"scouting"}
	case "goblin":
		return []string{"gathering"}
	case "kobold":
		return []string{"kobold"}
	default:
		return []string{"aggressive"}
	}
}

func defaultSleepCycle(species string) string {
	switch species {
	case "spider":
		return "nocturnal"
	default:
		return "diurnal"
	}
}

// CanMate checks whether two entities are compatible for reproduction.
func CanMate(a, b *entity.Entity) bool {
	if a == nil || b == nil {
		return false
	}
	if a.Species != b.Species {
		return false
	}
	if a.Gender == b.Gender {
		return false
	}
	if !a.Alive || !b.Alive {
		return false
	}
	if a.Pregnant || b.Pregnant {
		return false
	}
	// Immortal or undead species do not reproduce.
	if a.Species == "deity" || a.Species == "vampire" {
		return false
	}
	return true
}

// SpawnBaby creates a new offspring entity from two parents.
// The caller is responsible for providing a unique ID for the baby.
func SpawnBaby(parent1, parent2 *entity.Entity, id, babyName string, rng func(int) int) *entity.Entity {
	if !CanMate(parent1, parent2) {
		return nil
	}

	attrs := inheritAttributes(parent1.Attributes, parent2.Attributes, rng)
	baby := entity.NewEntity(
		id,
		babyName,
		parent1.Species,
		attrs,
		1,
	)

	// Inherit gender randomly from one of the parents
	if rng(2) == 0 {
		baby.Gender = parent1.Gender
	} else {
		baby.Gender = parent2.Gender
	}

	// Mark the female parent as pregnant
	if parent1.Gender == entity.GenderFemale {
		parent1.Pregnant = true
	} else if parent2.Gender == entity.GenderFemale {
		parent2.Pregnant = true
	}

	return baby
}

func inheritAttributes(a, b entity.Attributes, rng func(int) int) entity.Attributes {
	return entity.Attributes{
		STR: clampAttr((a.STR+b.STR)/2 + rng(3) - 1),
		DEX: clampAttr((a.DEX+b.DEX)/2 + rng(3) - 1),
		CON: clampAttr((a.CON+b.CON)/2 + rng(3) - 1),
		INT: clampAttr((a.INT+b.INT)/2 + rng(3) - 1),
		WIS: clampAttr((a.WIS+b.WIS)/2 + rng(3) - 1),
		CHA: clampAttr((a.CHA+b.CHA)/2 + rng(3) - 1),
	}
}

func clampAttr(v int) int {
	if v < 3 {
		return 3
	}
	if v > 20 {
		return 20
	}
	return v
}

// GestationTicks defines how many ticks a pregnancy lasts before the baby is born.
const GestationTicks = 200

// StartPregnancy marks a female entity as pregnant with a given father and start tick.
func StartPregnancy(mother, father *entity.Entity, tick uint64) {
	if mother == nil || father == nil {
		return
	}
	if mother.Gender != entity.GenderFemale {
		return
	}
	if mother.Pregnant {
		return
	}
	mother.Pregnant = true
	mother.PregnantSinceTick = tick
	mother.FatherID = father.ID
}

// ProcessPregnancy checks all entities for completed pregnancies and spawns babies.
func ProcessPregnancy(em *entity.Manager, tick uint64, rng *rand.Rand) {
	for _, e := range em.All() {
		if !e.Pregnant {
			continue
		}
		if tick < e.PregnantSinceTick {
			continue
		}
		if tick-e.PregnantSinceTick < GestationTicks {
			continue
		}
		// Find father entity
		var father *entity.Entity
		for _, cand := range em.All() {
			if cand.ID == e.FatherID {
				father = cand
				break
			}
		}
		if father == nil {
			// father not found, clear pregnancy
			e.Pregnant = false
			e.PregnantSinceTick = 0
			e.FatherID = ""
			continue
		}
		// Generate baby ID and name
		babyName := generateName(e.Species, rng)
		babyID := fmt.Sprintf("%s_baby_%s_%s_%d", e.Species, babyName, e.LocationID, tick)
		baby := SpawnBaby(e, father, babyID, babyName, func(n int) int { return rng.Intn(n) })
		if baby != nil {
			em.Add(baby)
			// Clear pregnancy
			e.Pregnant = false
			e.PregnantSinceTick = 0
			e.FatherID = ""
		}
	}
}
