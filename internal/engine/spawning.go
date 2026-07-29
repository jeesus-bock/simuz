// Package engine contains the simulation engine, tick processing, and related systems.
package engine

import (
	"fmt"
	"log"
	"maps"
	"math/rand"
	"slices"

	"simuz/internal/entity"
	"simuz/internal/relation"
	"simuz/internal/items"
	"simuz/internal/species"
	"simuz/internal/world"
)

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

func NewSpawnManager() *SpawnManager {
	return &SpawnManager{
		Rules: []SpawnRule{
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
		},
	}
}

func (sm *SpawnManager) ProcessSpawns(w *world.World, worldEntities *entity.Manager, tick int, rng *rand.Rand) {
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

	ent := entity.NewEntity(id, name, rule.Species, attrs, level, relation.EmptyRelation)
	ent.LocationID = rule.LocationID
	ent.Faction = rule.Faction
	ent.Profession = rule.Profession
	if species, exists := species.GetByID(rule.Species); exists && species.CanReproduce {
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
		if rule.Profession == "bandit" {
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
	humanNames := []string{"Aldric", "Brenna", "Cedric", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jenna", "Kol", "Lyssa", "Maren", "Nolan", "Opal", "Petra", "Quinn", "Rhea", "Soren", "Tessa"}
	names := map[string][]string{
		"orc": orcNames, "wolf": wolfNames, "bear": bearNames,
		"boar": boarNames, "rat": ratNames, "spider": spiderNames, "goblin": goblinNames,
		"kobold": koboldNames, "human": humanNames,
	}
	pool, ok := names[species]
	if !ok || len(pool) == 0 {
		return species + "_spawn"
	}
	return pool[rng.Intn(len(pool))]
}

func defaultScripts(species, profession string) []string {
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
		switch profession {
		case "bard":
			return []string{"bard"}
		case "priest":
			return []string{"priest"}
		case "bandit":
			return []string{"bandit"}
		case "ranger":
			return []string{"ranger"}
		case "merchant":
			return []string{"merchant"}
		case "cultist":
			return []string{"cult_member"}
		default:
			return []string{"aggressive"}
		}
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

// averageAttrs returns a child's attributes averaged from both parents with small random variation.
func averageAttrs(a, b entity.Attributes, rng *rand.Rand) entity.Attributes {
	return entity.Attributes{
		STR: clampInt((a.STR+b.STR)/2+rng.Intn(3)-1, 3, 20),
		DEX: clampInt((a.DEX+b.DEX)/2+rng.Intn(3)-1, 3, 20),
		CON: clampInt((a.CON+b.CON)/2+rng.Intn(3)-1, 3, 20),
		INT: clampInt((a.INT+b.INT)/2+rng.Intn(3)-1, 3, 20),
		WIS: clampInt((a.WIS+b.WIS)/2+rng.Intn(3)-1, 3, 20),
		CHA: clampInt((a.CHA+b.CHA)/2+rng.Intn(3)-1, 3, 20),
	}
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
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
	if a.Reproduction.Pregnant || b.Reproduction.Pregnant {
		return false
	}
	// Immortal or undead species do not reproduce.
	if a.Species == "deity" || a.Species == "vampire" {
		return false
	}
	// Prevent incest: entities with a parent-child relationship cannot mate.
	if rel, ok := a.Relationships[b.ID]; ok {
		if rel.Type == entity.RelationshipParent || rel.Type == entity.RelationshipChild {
			return false
		}
	}
	return true
}

// SpawnBaby creates a new offspring entity from two parents.
// The caller is responsible for providing a unique ID for the baby.
func SpawnBaby(parent1, parent2 *entity.Entity, id, babyName string, tick uint64, rng func(int) int) *entity.Entity {
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
		relation.EmptyRelation,
	)
	baby.LocationID = parent1.LocationID
	baby.Faction = parent1.Faction
	baby.Profession = ""
	baby.AI = entity.EntityAI{
		Type:         "scripted",
		ScriptIDs:    defaultScripts(baby.Species, baby.Profession),
		FactionID:    parent1.Faction,
		SleepCycle:   defaultSleepCycle(baby.Species),
		HomeLocation: parent1.LocationID,
	}
	baby.XP = randomXPForLevel(1, rng)

	// Inherit gender randomly from one of the parents
	baby.Gender = entity.GetRndGender()

	// Mark the female parent as pregnant and record the father
	if parent1.Gender == entity.GenderFemale {
		parent1.Reproduction.Pregnant = true
		parent1.Reproduction.FatherID = parent2.ID
	} else if parent2.Gender == entity.GenderFemale {
		parent2.Reproduction.Pregnant = true
		parent2.Reproduction.FatherID = parent1.ID
	}

	// Establish relationships: mate bond and parent-child links
	parent1.AddRelationship(parent2.ID, entity.RelationshipMate, tick)
	parent2.AddRelationship(parent1.ID, entity.RelationshipMate, tick)
	parent1.AddRelationship(baby.ID, entity.RelationshipParent, tick)
	parent2.AddRelationship(baby.ID, entity.RelationshipParent, tick)
	baby.AddRelationship(parent1.ID, entity.RelationshipChild, tick)
	baby.AddRelationship(parent2.ID, entity.RelationshipChild, tick)

	// Give the baby a random amount of XP appropriate for its level.
	baby.XP = randomXPForLevel(1, rng)

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

// SpeciesGestationTicks maps species to their gestation period in ticks.
var SpeciesGestationTicks = map[string]int{
	"human":  280,
	"orc":    200,
	"elf":    300,
	"goblin": 100,
	"kobold": 80,
	"wolf":   60,
	"bear":   90,
	"boar":   70,
	"rat":    30,
	"spider": 40,
}

// GestationTicksForSpecies returns the gestation period for a species in ticks.
// Falls back to 200 ticks if the species has no entry in the map.
func GestationTicksForSpecies(species string) int {
	if ticks, ok := SpeciesGestationTicks[species]; ok {
		return ticks
	}
	return 200
}

// StartPregnancy marks a female entity as pregnant with a given father and start tick.
func StartPregnancy(mother, father *entity.Entity, tick uint64) {
	if mother == nil || father == nil {
		return
	}
	if mother.Gender != entity.GenderFemale {
		return
	}
	if mother.Reproduction.Pregnant {
		return
	}
	mother.Reproduction.Pregnant = true
	mother.Reproduction.PregnantSinceTick = tick
	mother.Reproduction.FatherID = father.ID
}

// ProcessPregnancy checks all entities for completed pregnancies and spawns babies.
func ProcessPregnancy(em *entity.Manager, tick uint64, rng *rand.Rand) {
	for _, e := range em.All() {
		if !e.Reproduction.Pregnant {
			continue
		}
		if tick < e.Reproduction.PregnantSinceTick {
			continue
		}
		gestation := GestationTicksForSpecies(e.Species)
		if tick-e.Reproduction.PregnantSinceTick < uint64(gestation) {
			continue
		}
		// Find father entity
		var father *entity.Entity
		for _, cand := range em.All() {
			if cand.ID == e.Reproduction.FatherID {
				father = cand
				break
			}
		}
		if father == nil {
			// father not found, clear pregnancy
			e.Reproduction.Pregnant = false
			e.Reproduction.PregnantSinceTick = 0
			e.Reproduction.FatherID = ""
			continue
		}
		// Generate baby ID and name
		babyName := generateName(e.Species, rng)
		babyID := fmt.Sprintf("%s_baby_%s_%s_%d", e.Species, babyName, e.LocationID, tick)
		// Clear pregnancy before spawning so CanMate won't reject the pair.
		e.Reproduction.Pregnant = false
		e.Reproduction.PregnantSinceTick = 0
		e.Reproduction.FatherID = ""
		baby := SpawnBaby(e, father, babyID, babyName, tick, func(n int) int { return rng.Intn(n) })
		if baby != nil {
			em.Add(baby)
		}
	}
}

// SeedFamilies creates multi-generational family groups at simulation start.
// It places 2–5 families across civilian (indoor) locations in the world.
func SeedFamilies(em *entity.Manager, w *world.World, rng *rand.Rand) {
	const maxFamilies = 5

	var candidates []*world.Location
	for _, loc := range w.AllLocations() {
		if loc.IsOutside {
			continue // families live in towns, villages, and buildings
		}
		if w.IsDivineRealm(loc.ID) {
			continue
		}
		candidates = append(candidates, loc)
	}
	if len(candidates) == 0 {
		return
	}

	numFamilies := 2 + rng.Intn(maxFamilies-1)
	for i := 0; i < numFamilies; i++ {
		loc := candidates[rng.Intn(len(candidates))]
		seedFamilyAtLocation(em, loc.ID, rng)
	}
}

func seedFamilyAtLocation(em *entity.Manager, locID string, rng *rand.Rand) {
	species := pickFamilySpecies(rng)
	tick := uint64(0) // relationships start at tick 0 for seeded families

	// Generation 1: Grandparents (old, level 6–9)
	grandpa := createSeededEntity(species, "male", 6+rng.Intn(4), locID, rng)
	grandma := createSeededEntity(species, "female", 6+rng.Intn(4), locID, rng)

	grandpa.AddRelationship(grandma.ID, entity.RelationshipMate, tick)
	grandma.AddRelationship(grandpa.ID, entity.RelationshipMate, tick)

	// Generation 2: Parents (children of grandparents, level 3–6)
	numParents := 1 + rng.Intn(2) // 1–2 parents
	parents := make([]*entity.Entity, 0, numParents)
	for i := 0; i < numParents; i++ {
		gender := "male"
		if i == 0 {
			gender = "female"
		}
		parent := createSeededEntity(species, gender, 3+rng.Intn(4), locID, rng)
		parents = append(parents, parent)

		// Grandparent ↔ parent relationships
		grandpa.AddRelationship(parent.ID, entity.RelationshipChild, tick)
		grandma.AddRelationship(parent.ID, entity.RelationshipChild, tick)
		parent.AddRelationship(grandpa.ID, entity.RelationshipParent, tick)
		parent.AddRelationship(grandma.ID, entity.RelationshipParent, tick)
	}

	// Mate relationship between parents (if 2 parents)
	if len(parents) >= 2 {
		parents[0].AddRelationship(parents[1].ID, entity.RelationshipMate, tick)
		parents[1].AddRelationship(parents[0].ID, entity.RelationshipMate, tick)
	}

	// Generation 3: Children (children of parents, level 1–3)
	numChildren := 1 + rng.Intn(3) // 1–3 children
	children := make([]*entity.Entity, 0, numChildren)
	for i := 0; i < numChildren; i++ {
		gender := "male"
		if rng.Intn(2) == 0 {
			gender = "female"
		}
		child := createSeededEntity(species, gender, 1+rng.Intn(3), locID, rng)
		children = append(children, child)

		// Parent ↔ child relationships
		for _, parent := range parents {
			parent.AddRelationship(child.ID, entity.RelationshipChild, tick)
			child.AddRelationship(parent.ID, entity.RelationshipParent, tick)
		}

		// Sibling relationships
		for _, sibling := range children {
			if sibling.ID != child.ID {
				child.AddRelationship(sibling.ID, entity.RelationshipSibling, tick)
				sibling.AddRelationship(child.ID, entity.RelationshipSibling, tick)
			}
		}
	}

	// Generation 4: Grandchildren (children of one adult parent, level 1)
	if len(parents) > 0 {
		for _, parent := range parents {
			if parent.Level >= 3 && rng.Intn(100) < 30 {
				gcGender := "male"
				if rng.Intn(2) == 0 {
					gcGender = "female"
				}
				gc := createSeededEntity(species, gcGender, 1, locID, rng)

				parent.AddRelationship(gc.ID, entity.RelationshipChild, tick)
				gc.AddRelationship(parent.ID, entity.RelationshipParent, tick)

				// Grandparent ↔ grandchild relationships
				grandpa.AddRelationship(gc.ID, entity.RelationshipChild, tick)
				grandma.AddRelationship(gc.ID, entity.RelationshipChild, tick)
				gc.AddRelationship(grandpa.ID, entity.RelationshipParent, tick)
				gc.AddRelationship(grandma.ID, entity.RelationshipParent, tick)

				// Sibling relationships with existing children
				for _, sibling := range children {
					gc.AddRelationship(sibling.ID, entity.RelationshipSibling, tick)
					sibling.AddRelationship(gc.ID, entity.RelationshipSibling, tick)
				}

				em.Add(gc)
				break // one grandchild per family is enough
			}
		}
	}

	// Add all entities to the manager
	em.Add(grandpa)
	em.Add(grandma)
	for _, p := range parents {
		em.Add(p)
	}
	for _, c := range children {
		em.Add(c)
	}

	log.Printf("[seed] seeded %s family at %s: grandpa=%s grandma=%s parents=%d children=%d",
		species, locID, grandpa.Name, grandma.Name, len(parents), len(children))
}

func pickFamilySpecies(rng *rand.Rand) string {
	species := slices.Collect(maps.Keys(species.Registry))
	return species[rng.Intn(len(species))]
}

func createSeededEntity(species, gender string, level int, locID string, rng *rand.Rand) *entity.Entity {
	attrs := baseSpeciesAttrs(species, rng)
	name := generateName(species, rng)
	id := fmt.Sprintf("%s_seed_%s_%s_%d", species, gender, name, rng.Intn(100000))

	ent := entity.NewEntity(id, name, species, attrs, level, relation.EmptyRelation)
	ent.Gender = gender
	ent.LocationID = locID
	ent.Faction = "civilian"
	ent.Profession = pickCivilianProfession(rng)
	ent.AI = entity.EntityAI{
		Type:         "passive",
		SleepCycle:   defaultSleepCycle(species),
		HomeLocation: locID,
	}
	// Seeded entities start with a random amount of XP appropriate for their level.
	ent.XP = randomXPForLevel(level, rng.Intn)
	return ent
}

func pickCivilianProfession(rng *rand.Rand) string {
	professions := []string{"", "farmer", "merchant", "herbalist", "miner", "fisherman", "craftsman", "scholar", "bard", "priest"}
	return professions[rng.Intn(len(professions))]
}
