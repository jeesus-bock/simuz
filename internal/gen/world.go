// Package gen contains world generation helpers and seeded simulation setup utilities.
package gen

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"

	"simuz/internal/entity"
	"simuz/internal/relation"
	"simuz/internal/world"
)

type Generator struct {
	Seed  string
	RNG   *rand.Rand
	World *world.World
}

func NewGenerator(seed string) *Generator {
	h := sha256.Sum256([]byte(seed))
	rng := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(h[:8]))))
	return &Generator{
		Seed:  seed,
		RNG:   rng,
		World: world.NewWorld(),
	}
}

// Structural blueprint for generating randomized geographic profiles
type biomeProfile struct {
	namePrefixes []string
	nameSuffixes []string
	baseClimate  world.Season // Assuming world.Season dictates initial climate arrays
}

// Global data matrix providing vocabulary seeds for unique, varied world regions
var biomeMatrix = map[string]biomeProfile{
	"highlands": {
		namePrefixes: []string{"Northern", "Frost", "Iron", "Craggy", "Bleak"},
		nameSuffixes: []string{"Highlands", "Peaks", "Ridges", "Crags"},
		baseClimate:  world.Winter,
	},
	"swamp": {
		namePrefixes: []string{"Sunken", "Mire", "Rotting", "Murky", "Gloom"},
		nameSuffixes: []string{"Marches", "Bogs", "Fens", "Wetlands"},
		baseClimate:  world.Autumn,
	},
	"plains": {
		namePrefixes: []string{"Golden", "Whispering", "Sun-Drenched", "Vast"},
		nameSuffixes: []string{"Plains", "Meadows", "Steppes", "Prairies"},
		baseClimate:  world.Spring,
	},
	"forest": {
		namePrefixes: []string{"Crystal", "Ancient", "Shadow", "Nettle", "Whisper"},
		nameSuffixes: []string{"Forest", "Woods", "Thickets", "Groves"},
		baseClimate:  world.Spring,
	},
	"waste": {
		namePrefixes: []string{"Ash", "Scorched", "Barren", "Sulfur", "Bleak"},
		nameSuffixes: []string{"Desert", "Wastes", "Barrens", "Flats"},
		baseClimate:  world.Summer,
	},
}

// Generate procedurally creates a unique, scaled layout of the entire world simulation.
func (g *Generator) Generate() (*world.World, []*entity.Entity) {
	// 1. Establish the master cosmic meta-node container
	worldLoc := world.NewLocation("aetheria", "Aetheria", world.LocWorld, "", world.Position{})
	worldLoc.Weather = world.NewWeather(world.Clear, 15)
	g.World.AddLocation(worldLoc)

	// 2. Procedurally determine how many regions this specific world pass will contain (e.g., 4 to 8)
	numRegions := 4 + g.RNG.Intn(5)

	// Create an index tracker slice to shuffle biome types so we get diverse variations
	biomeTypes := []string{"highlands", "swamp", "plains", "forest", "waste"}

	for i := 0; i < numRegions; i++ {
		// Pick a random biome profile archetype from the matrix map
		bType := biomeTypes[g.RNG.Intn(len(biomeTypes))]
		profile := biomeMatrix[bType]

		// 3. Procedural Language Mixing: Assemble a completely unique geographic name descriptor
		rID := fmt.Sprintf("region_%s_%d", bType, i)
		pfx := profile.namePrefixes[g.RNG.Intn(len(profile.namePrefixes))]
		sfx := profile.nameSuffixes[g.RNG.Intn(len(profile.nameSuffixes))]
		rName := fmt.Sprintf("%s %s", pfx, sfx)

		// 4. Circle-Packing Position Fix: Distribute coordinates organically in expanding spirals
		// This mathematically guarantees layout zones never perfectly mirror each other or clump safely
		angle := float64(i)*(2*math.Pi/float64(numRegions)) + (g.RNG.Float64() * 0.4)
		distance := 300.0 + (g.RNG.Float64() * 250.0) // Varied spatial radius boundaries
		rx := math.Round(400.0 + distance*math.Cos(angle))
		ry := math.Round(400.0 + distance*math.Sin(angle))

		// 5. Generate and register the concrete region context node
		loc := world.NewLocation(rID, rName, world.LocRegion, "aetheria", world.Position{X: rx, Y: ry})

		// Feed the weather generator the specific seasonal preference of the generated biome type
		loc.Weather = world.GenerateWeatherFor(profile.baseClimate, rID, g.RNG)
		g.World.AddLocation(loc)
	}

	// 6. Execute systemic connection passes based on the newly generated variable layout
	g.generateWildSites()
	g.generateRegionExits()

	// 7. Populate towns and entities dynamically
	entities := g.generateSettlements()

	return g.World, entities
}

// generateRegionExits automatically scans the 2D grid layout, finds the closest
// geographic neighbors, and builds a logically sound bidirectional traversal mesh.
func (g *Generator) generateRegionExits() {
	// 1. Gather all procedurally generated regions from your active world registry
	var regions []*world.Location
	for _, loc := range g.World.AllLocations() {
		if loc.Type == world.LocRegion {
			regions = append(regions, loc)
		}
	}

	// If there aren't enough sectors to form pairs, abort early to prevent compilation faults
	if len(regions) < 2 {
		return
	}

	// 2. Proximity Graph Loop: For every region, find its absolute closest neighbors
	for i, rA := range regions {
		// We track the two closest neighbor nodes to ensure every zone has at least two exit links,
		// preventing isolated islands or dead-end map anomalies.
		var firstClosest *world.Location
		var secondClosest *world.Location

		firstDist := math.MaxFloat64
		secondDist := math.MaxFloat64

		for j, rB := range regions {
			if i == j {
				continue // Skip evaluating a region against itself
			}

			// Calculate Euclidean straight-line grid distance: sqrt((x1-x2)^2 + (y1-y2)^2)
			dx := rA.Position.X - rB.Position.X
			dy := rA.Position.Y - rB.Position.Y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < firstDist {
				secondDist = firstDist
				secondClosest = firstClosest
				firstDist = dist
				firstClosest = rB
			} else if dist < secondDist {
				secondDist = dist
				secondClosest = rB
			}
		}

		// 3. Connect the closest neighbor nodes dynamically using spatial directional tags
		if firstClosest != nil {
			dirA, dirB := calculateDirections(rA.Position, firstClosest.Position)
			g.World.AddBidirectionalExit(rA.ID, firstClosest.ID, dirA, dirB)
		}
		if secondClosest != nil {
			dirA, dirB := calculateDirections(rA.Position, secondClosest.Position)
			g.World.AddBidirectionalExit(rA.ID, secondClosest.ID, dirA, dirB)
		}
	}
}

// calculateDirections maps trigonometry vector angles into intuitive compass strings.
func calculateDirections(posA, posB world.Position) (string, string) {
	dx := posB.X - posA.X
	dy := posB.Y - posA.Y

	// Calculate radiant vector arc angle (-Pi to +Pi)
	angle := math.Atan2(dy, dx)
	// Convert radians into standard degrees (0 to 360) for easy slicing
	degrees := angle * (180.0 / math.Pi)
	if degrees < 0 {
		degrees += 360.0
	}

	// 4. Compass Matrix Partition: Divide 360 degrees into 8 clear directional wedges
	var dirA, dirB string
	switch {
	case degrees >= 337.5 || degrees < 22.5:
		dirA, dirB = "east", "west"
	case degrees >= 22.5 && degrees < 67.5:
		dirA, dirB = "southeast", "northwest"
	case degrees >= 67.5 && degrees < 112.5:
		dirA, dirB = "south", "north"
	case degrees >= 112.5 && degrees < 157.5:
		dirA, dirB = "southwest", "northeast"
	case degrees >= 157.5 && degrees < 202.5:
		dirA, dirB = "west", "east"
	case degrees >= 202.5 && degrees < 247.5:
		dirA, dirB = "northwest", "southeast"
	case degrees >= 247.5 && degrees < 292.5:
		dirA, dirB = "north", "south"
	default:
		dirA, dirB = "northeast", "southwest"
	}

	return dirA, dirB
}
func (g *Generator) generateSettlements() []*entity.Entity {
	var all []*entity.Entity
	all = append(all, g.generateTown("frosthold", "Frosthold", "northern_highlands", 50, 60)...)
	all = append(all, g.generateTown("stillwater", "Stillwater", "sunken_marches", 20, 30)...)
	all = append(all, g.generateTown("golden_gate", "Golden Gate", "golden_plains", 100, 80)...)
	all = append(all, g.generateFisherman()...)
	all = append(all, g.generateRatKingLair()...)
	all = append(all, g.generateTravelingSalesmen()...)
	all = append(all, g.generateBards()...)
	all = append(all, g.generateFarms()...)
	all = append(all, g.generateHostiles()...)
	all = append(all, g.generateBeasts()...)
	all = append(all, g.generateNewArchetypes()...)
	all = append(all, g.generateNewCreatures()...)
	all = append(all, g.generateTownExtras()...)
	all = append(all, g.generateWildernessBosses()...)
	all = append(all, g.generateGoblinAmbushers()...)
	return all
}

func (g *Generator) generateHostiles() []*entity.Entity {
	var all []*entity.Entity

	orcDefs := []struct {
		id, name, locID string
	}{
		{"orc_skar", "Skar", "orc_camp"},
		{"orc_grom", "Grom", "orc_camp"},
		{"orc_grak", "Grak", "orc_camp"},
		{"orc_uruk", "Uruk", "ash_ruins"},
	}
	elfDefs := []struct {
		id, name, locID string
	}{
		{"elf_aerin", "Aerin", "fey_glade"},
		{"elf_thalion", "Thalion", "fey_glade"},
		{"elf_lyra", "Lyra", "fey_glade"},
	}
	thiefDefs := []struct {
		id, name, locID string
	}{
		{"thief_rat", "Rattle", "frosthold_inn_common"},
		{"thief_creep", "Creep", "stillwater_inn_common"},
		{"thief_sneak", "Sneak", "golden_gate_inn_common"},
	}
	banditDefs := []struct {
		id, name, locID string
	}{
		{"bandit_knife", "Knife", "bandit_camp"},
		{"bandit_blade", "Blade", "bandit_camp"},
		{"bandit_jack", "Jack", "bandit_camp"},
		{"bandit_rog", "Rog", "bandit_camp"},
	}

	for _, o := range orcDefs {
		ent := entity.NewEntity(o.id, o.name, "orc", entity.RandomAttributes(g.RNG.Intn), 3+g.RNG.Intn(3), relation.OrcRelation)
		ent.LocationID = o.locID
		ent.Faction = "orc"
		ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "orc", SleepCycle: "diurnal", HomeLocation: o.locID}
		equipItem(ent, lookup("leather_armor"))
		if g.RNG.Intn(100) < 50 {
			equipItem(ent, lookup("leather_helmet"))
		}
		if g.RNG.Intn(100) < 40 {
			equipItem(ent, lookup("wooden_shield"))
		}
		// Orcs always fight armed — heavy weapons preferred.
		roll := g.RNG.Intn(100)
		switch {
		case roll < 35:
			equipItem(ent, lookup("orc_cleaver"))
		case roll < 60:
			equipItem(ent, lookup("iron_axe"))
		case roll < 80:
			equipItem(ent, lookup("iron_spear"))
		default:
			equipItem(ent, lookup("iron_sword"))
		}
		giveCurrency(ent, g.RNG.Intn(10), g.RNG.Intn(5), 0)
		all = append(all, ent)
	}

	for _, e := range elfDefs {
		ent := entity.NewEntity(e.id, e.name, "elf", entity.RandomAttributes(g.RNG.Intn), 3+g.RNG.Intn(3), relation.ElfRelation)
		ent.LocationID = e.locID
		ent.Faction = "elf"
		ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "elf", SleepCycle: "diurnal", HomeLocation: e.locID}
		equipItem(ent, lookup("fine_clothes"))
		equipItem(ent, lookup("leather_boots"))
		if g.RNG.Intn(100) < 60 {
			equipItem(ent, lookup("short_sword"))
		} else {
			equipItem(ent, lookup("iron_spear"))
		}
		if g.RNG.Intn(100) < 30 {
			equipItem(ent, lookup("wooden_shield"))
		}
		giveCurrency(ent, 5+g.RNG.Intn(10), 2+g.RNG.Intn(5), 0)
		all = append(all, ent)
	}

	for _, t := range thiefDefs {
		ent := entity.NewEntity(t.id, t.name, "human", entity.RandomAttributes(g.RNG.Intn), 2+g.RNG.Intn(2), relation.ThiefRelation)
		ent.LocationID = t.locID
		ent.Faction = "thieves_guild"
		ent.Profession = "thief"
		ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"thief"}, FactionID: "thief", SleepCycle: "nocturnal", HomeLocation: t.locID}
		equipItem(ent, lookup("common_clothes"))
		equipItem(ent, lookup("leather_boots"))
		if g.RNG.Intn(100) < 70 {
			equipItem(ent, lookup("dagger"))
		} else {
			equipItem(ent, lookup("short_sword"))
		}
		giveCurrency(ent, g.RNG.Intn(15), g.RNG.Intn(5), 0)
		all = append(all, ent)
	}

	for _, b := range banditDefs {
		ent := entity.NewEntity(b.id, b.name, "human", entity.RandomAttributes(g.RNG.Intn), 2+g.RNG.Intn(3), relation.BanditRelation)
		ent.LocationID = b.locID
		ent.Faction = ""
		ent.Profession = "bandit"
		ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "bandit", SleepCycle: "diurnal", HomeLocation: b.locID}
		equipItem(ent, lookup("leather_armor"))
		equipItem(ent, lookup("leather_boots"))
		if g.RNG.Intn(100) < 45 {
			equipItem(ent, lookup("leather_helmet"))
		}
		if g.RNG.Intn(100) < 35 {
			equipItem(ent, lookup("wooden_shield"))
		}
		// Bandits always carry a real fighting weapon.
		roll := g.RNG.Intn(100)
		switch {
		case roll < 40:
			equipItem(ent, lookup("iron_sword"))
		case roll < 65:
			equipItem(ent, lookup("short_sword"))
		case roll < 85:
			equipItem(ent, lookup("iron_axe"))
		default:
			equipItem(ent, lookup("iron_spear"))
		}
		giveCurrency(ent, g.RNG.Intn(20), g.RNG.Intn(10), 0)
		all = append(all, ent)
	}

	return all
}

func (g *Generator) generateBeasts() []*entity.Entity {
	beastSpawns := []struct {
		id, name, species, locID string
		level                    int
		attrs                    entity.Attributes
		nocturnal                bool
	}{
		{"wolf_shadow", "Shadow", "wolf", "wolf_den", 2, entity.Attributes{STR: 14, DEX: 16, CON: 12, INT: 4, WIS: 8, CHA: 4}, false},
		{"wolf_fang", "Fang", "wolf", "wolf_den", 3, entity.Attributes{STR: 15, DEX: 15, CON: 13, INT: 4, WIS: 8, CHA: 4}, false},
		{"wolf_snarl", "Snarl", "wolf", "wolf_den", 2, entity.Attributes{STR: 13, DEX: 17, CON: 11, INT: 4, WIS: 8, CHA: 4}, false},
		{"bear_brown", "Bruin", "bear", "bear_den", 5, entity.Attributes{STR: 20, DEX: 10, CON: 18, INT: 3, WIS: 7, CHA: 3}, false},
		{"bear_black", "Ursa", "bear", "bear_den", 4, entity.Attributes{STR: 18, DEX: 11, CON: 16, INT: 3, WIS: 7, CHA: 3}, false},
		{"boar_tusk", "Tusker", "boar", "boar_wallow", 3, entity.Attributes{STR: 16, DEX: 12, CON: 15, INT: 3, WIS: 6, CHA: 3}, false},
		{"boar_bristle", "Bristle", "boar", "boar_wallow", 2, entity.Attributes{STR: 14, DEX: 13, CON: 14, INT: 3, WIS: 6, CHA: 3}, false},
		{"bat_swoop", "Swoop", "bat", "ash_ruins", 1, entity.Attributes{STR: 5, DEX: 18, CON: 6, INT: 2, WIS: 10, CHA: 2}, true},
		{"spider_fang", "Fang", "spider", "spider_grove", 3, entity.Attributes{STR: 12, DEX: 16, CON: 10, INT: 2, WIS: 8, CHA: 2}, true},
		{"spider_web", "Web", "spider", "spider_grove", 2, entity.Attributes{STR: 10, DEX: 17, CON: 9, INT: 2, WIS: 8, CHA: 2}, true},
	}

	var all []*entity.Entity
	for _, b := range beastSpawns {
		ent := entity.NewEntity(b.id, b.name, b.species, b.attrs, b.level, relation.BeastRelation)
		ent.LocationID = b.locID
		ent.Faction = "beast"
		cycle := "diurnal"
		if b.nocturnal {
			cycle = "nocturnal"
		}
		if b.species == "wolf" {
			ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"hunting"}, FactionID: "beast", SleepCycle: cycle, HomeLocation: b.locID}
		} else if b.nocturnal {
			ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"scouting"}, FactionID: "beast", SleepCycle: cycle, HomeLocation: b.locID}
		} else {
			ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "beast", SleepCycle: cycle, HomeLocation: b.locID}
		}
		equipNaturalWeapon(ent, b.species)
		all = append(all, ent)
	}
	return all
}

// equipNaturalWeapon gives beasts species-appropriate natural weapons.
func equipNaturalWeapon(ent *entity.Entity, species string) {
	switch species {
	case "wolf", "bear", "badger", "dog":
		equipItem(ent, lookup("claws"))
	case "spider", "bat", "rat", "rat_king":
		equipItem(ent, lookup("fangs"))
	case "boar":
		equipItem(ent, lookup("tusks"))
	}
}

func (g *Generator) generateTown(id, name, regionID string, x, y float64) []*entity.Entity {
	town := world.NewLocation(id, name, world.LocCity, regionID, world.Position{X: x, Y: y})
	town.IsOutside = false
	g.World.AddLocation(town)

	buildings := []struct {
		id, bname, btype string
		bx, by           float64
	}{
		{id + "_inn", "The Sleeping Dragon", "building", 0, 0},
		{id + "_temple", "Temple of Light", "building", -5, -3},
		{id + "_guardhouse", "Guardhouse", "building", 2, -2},
		{id + "_blacksmith", "The Iron Anvil", "building", -2, 4},
	}

	for _, b := range buildings {
		loc := world.NewLocation(b.id, b.bname, world.LocBuilding, id, world.Position{X: b.bx, Y: b.by})
		loc.IsOutside = false
		switch b.id {
		case id + "_blacksmith":
			loc.Tags = []string{"forge", "blacksmith"}
		case id + "_temple":
			loc.Tags = []string{"cauldron", "temple"}
		case id + "_guardhouse":
			loc.Tags = []string{"workbench", "guardhouse"}
		case id + "_inn":
			loc.Tags = []string{"inn"}
		case id + "_market":
			loc.Tags = []string{"market"}
		}
		g.World.AddLocation(loc)
	}

	g.generateRooms(id + "_inn")

	return g.generateNPCs(id)
}

func (g *Generator) generateRooms(innID string) {
	rooms := []struct {
		id, name string
		x, y     float64
	}{
		{innID + "_common", "Common Room", 0, 0},
		{innID + "_cellar", "Cellar", 3, 0},
		{innID + "_kitchen", "Kitchen", 0, 3},
	}

	for _, r := range rooms {
		loc := world.NewLocation(r.id, r.name, world.LocRoom, innID, world.Position{X: r.x, Y: r.y})
		loc.IsOutside = false
		if r.id == innID+"_common" {
			loc.Tags = []string{"inn"}
		}
		g.World.AddLocation(loc)
	}
}

func (g *Generator) generateNPCs(townID string) []*entity.Entity {
	npcDefs := []struct {
		id, name, species, profession, locID string
	}{
		{townID + "_greta", "Greta", "human", "innkeeper", townID + "_inn_common"},
		{townID + "_sven", "Sven", "human", "blacksmith", townID + "_blacksmith"},
		{townID + "_guard_captain", "Captain Halvar", "human", "guard", townID + "_guardhouse"},
		{townID + "_priest", "Father Luan", "human", "priest", townID + "_temple"},
	}
	var entities []*entity.Entity
	for _, n := range npcDefs {
		ent := entity.NewEntity(n.id, n.name, n.species, entity.RandomAttributes(g.RNG.Intn), 1, relation.CivilianRelation)
		ent.LocationID = n.locID
		ent.AI = entity.EntityAI{
			Type:         "passive",
			FactionID:    townID,
			HomeLocation: n.locID,
			SleepCycle:   "diurnal",
		}
		ent.Faction = "civilian"

		switch n.profession {
		case "innkeeper":
			ent.Profession = n.profession
			ent.AI.Type = "scripted"
			ent.AI.ScriptIDs = []string{"innkeeper"}
			equipItem(ent, lookup("common_clothes"))
			equipItem(ent, lookup("tankard"))
			addItem(ent, lookup("beer"))
			addItem(ent, lookup("beer"))
			addItem(ent, lookup("wine"))
			addItem(ent, lookup("liquor"))
			addItem(ent, lookup("ale"))
			addItem(ent, lookup("mead"))
			addItem(ent, lookup("brandy"))
			giveCurrency(ent, 10+g.RNG.Intn(20), 3+g.RNG.Intn(5), 0)
		case "blacksmith":
			ent.Profession = n.profession
			ent.Relation = relation.CivilianRelation
			ent.AI.Type = "scripted"
			ent.AI.ScriptIDs = []string{"blacksmith"}
			equipItem(ent, lookup("work_tunic"))
			equipItem(ent, lookup("smith_hammer"))
			addItem(ent, lookup("iron_ore"))
			addItem(ent, lookup("iron_ore"))
			addItem(ent, lookup("iron_ore"))
			addItem(ent, lookup("coal"))
			addItem(ent, lookup("coal"))
			addItem(ent, lookup("leather"))
			addItem(ent, lookup("cloth"))
			addItem(ent, lookup("leather_strips"))
			giveCurrency(ent, 5+g.RNG.Intn(10), 5+g.RNG.Intn(10), 1+g.RNG.Intn(3))
		case "guard":
			ent.Profession = n.profession
			ent.Relation = relation.CivilianRelation
			ent.AI.Type = "scripted"
			ent.AI.ScriptIDs = []string{"guard"}
			ent.AI.Brave = true
			equipItem(ent, lookup("chainmail"))
			equipItem(ent, lookup("iron_helmet"))
			equipItem(ent, lookup("iron_boots"))
			equipItem(ent, lookup("iron_sword"))
			equipItem(ent, lookup("iron_shield"))
			equipItem(ent, lookup("leather_gloves"))
			giveCurrency(ent, 5, 3, 0)
		case "priest":
			ent.Profession = n.profession
			ent.Relation = relation.CivilianRelation
			ent.AI.Type = "scripted"
			ent.AI.ScriptIDs = []string{"priest"}
			equipItem(ent, lookup("priest_robe"))
			equipItem(ent, lookup("holy_symbol"))
			giveCurrency(ent, 3+g.RNG.Intn(6), 2+g.RNG.Intn(3), 0)
		default:
			ent.Relation = relation.CivilianRelation
			equipItem(ent, lookup("common_clothes"))
			giveCurrency(ent, 1+g.RNG.Intn(5), 0, 0)
		}

		entities = append(entities, ent)
	}
	return entities
}

func (g *Generator) generateTravelingSalesmen() []*entity.Entity {
	salesmen := []struct {
		id, name, locID string
	}{
		{"merchant_marcus", "Marcus", "frosthold_inn_common"},
		{"merchant_elara", "Elara", "stillwater_market"},
		{"merchant_finn", "Finn", "golden_gate_market"},
	}
	tradeItems := []string{
		"common_clothes", "fine_clothes", "simple_robe", "work_tunic",
		"dagger", "short_sword", "cudgel", "tankard",
		"leather_helmet", "leather_boots", "leather_gloves",
		"wooden_shield", "holy_symbol",
		"beer", "wine", "liquor", "night_bloom", "sage_leaf",
	}
	var entities []*entity.Entity
	for _, s := range salesmen {
		ent := entity.NewEntity(s.id, s.name, "human", entity.RandomAttributes(g.RNG.Intn), 3, relation.MerchantRelation)
		ent.LocationID = s.locID
		ent.AI = entity.EntityAI{
			Type:         "scripted",
			ScriptIDs:    []string{"traveling_salesman"},
			FactionID:    "merchant",
			HomeLocation: s.locID,
			SleepCycle:   "diurnal",
		}
		ent.Faction = "merchant"
		equipItem(ent, lookup("fine_clothes"))
		equipItem(ent, lookup("leather_boots"))
		giveCurrency(ent, 50+g.RNG.Intn(100), 20+g.RNG.Intn(40), 10+g.RNG.Intn(20))
		for _, itemID := range tradeItems {
			def := lookup(itemID)
			if def != nil && g.RNG.Intn(100) < 60 {
				qty := 1 + g.RNG.Intn(3)
				for k := 0; k < qty; k++ {
					addItem(ent, def)
				}
			}
		}
		entities = append(entities, ent)
	}
	return entities
}

func (g *Generator) generateBards() []*entity.Entity {
	type bardDef struct {
		id, name, instrument string
		startLoc             string
		quality              string
	}

	bards := []bardDef{
		{"bard_lira", "Lira", "lute", "frosthold_inn_common", "lousy"},
		{"bard_finneas", "Finneas", "lute", "stillwater_inn_common", "mediocre"},
		{"bard_aria", "Aria", "flute", "golden_gate_inn_common", "great"},
	}

	var entities []*entity.Entity
	for _, b := range bards {
		attrs := entity.RandomAttributes(g.RNG.Intn)
		switch b.quality {
		case "lousy":
			attrs.CHA = 7 + g.RNG.Intn(4)
		case "mediocre":
			attrs.CHA = 11 + g.RNG.Intn(5)
		case "great":
			attrs.CHA = 16 + g.RNG.Intn(5)
		}

		ent := entity.NewEntity(b.id, b.name, "human", attrs, 1, relation.CivilianRelation)
		ent.LocationID = b.startLoc
		ent.AI = entity.EntityAI{
			Type:         "scripted",
			ScriptIDs:    []string{"bard"},
			FactionID:    "civilian",
			HomeLocation: b.startLoc,
			SleepCycle:   "diurnal",
		}
		ent.Faction = "civilian"
		ent.Mood = "neutral"

		equipItem(ent, lookup("fine_clothes"))
		equipItem(ent, lookup(b.instrument))
		giveCurrency(ent, 5+g.RNG.Intn(10), 2+g.RNG.Intn(5), 0)

		addItem(ent, lookup("beer"))
		addItem(ent, lookup("wine"))

		entities = append(entities, ent)
	}
	return entities
}

func (g *Generator) generateRatKingLair() []*entity.Entity {
	dungeon := world.NewLocation("rat_king_lair", "Rat King Lair", world.LocCity, "northern_highlands", world.Position{X: 80, Y: 70})
	dungeon.Tags = []string{"dungeon"}
	dungeon.IsOutside = false
	g.World.AddLocation(dungeon)

	entrance := world.NewLocation("rat_king_lair_entrance", "Entrance Chamber", world.LocRoom, "rat_king_lair", world.Position{})
	entrance.IsOutside = false
	g.World.AddLocation(entrance)

	corridor := world.NewLocation("rat_king_lair_corridor", "Deep Corridor", world.LocRoom, "rat_king_lair", world.Position{X: 0, Y: 5})
	corridor.IsOutside = false
	g.World.AddLocation(corridor)

	throne := world.NewLocation("rat_king_lair_throne", "Throne Chamber", world.LocRoom, "rat_king_lair", world.Position{X: 0, Y: 10})
	throne.IsOutside = false
	g.World.AddLocation(throne)

	ratAttrs1 := entity.Attributes{STR: 8, DEX: 12, CON: 10, INT: 2, WIS: 6, CHA: 2}
	ratAttrs2 := entity.Attributes{STR: 10, DEX: 14, CON: 12, INT: 3, WIS: 7, CHA: 3}

	var all []*entity.Entity
	for i := 0; i < 2; i++ {
		rat := entity.NewEntity("rat_scram"+fmt.Sprint(i), "Scram", "rat", ratAttrs1, 1, relation.VerminRelation)
		rat.LocationID = "rat_king_lair_entrance"
		rat.Faction = "vermin"
		rat.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"defensive"}, FactionID: "vermin", SleepCycle: "none", HomeLocation: "rat_king_lair_entrance"}
		all = append(all, rat)
	}
	for i := 0; i < 2; i++ {
		rat := entity.NewEntity("rat_gleam"+fmt.Sprint(i), "Gleam", "rat", ratAttrs2, 2, relation.VerminRelation)
		rat.LocationID = "rat_king_lair_corridor"
		rat.Faction = "vermin"
		rat.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"defensive"}, FactionID: "vermin", SleepCycle: "none", HomeLocation: "rat_king_lair_corridor"}
		all = append(all, rat)
	}
	// Rat scouts patrol around lair entrance
	for i := 0; i < 2; i++ {
		scout := entity.NewEntity("rat_scout"+fmt.Sprint(i), "Rat_Scout", "rat", ratAttrs1, 1, relation.VerminRelation)
		scout.LocationID = "rat_king_lair_entrance"
		scout.Faction = "vermin"
		scout.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"scouting"}, FactionID: "vermin", SleepCycle: "none", HomeLocation: "rat_king_lair_entrance"}
		all = append(all, scout)
	}

	ratKing := entity.NewEntity("rat_king", "Skreet the Unseen", "rat_king", entity.Attributes{STR: 18, DEX: 10, CON: 20, INT: 8, WIS: 12, CHA: 6}, 8, relation.VerminRelation)
	ratKing.LocationID = "rat_king_lair_throne"
	ratKing.Faction = "vermin"
	ratKing.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"rat_king"}, FactionID: "vermin", SleepCycle: "none"}
	ratKing.MaxHP = 200
	ratKing.HP = 200
	ratKing.Immortal = false
	all = append(all, ratKing)

	return all
}

func (g *Generator) generateFisherman() []*entity.Entity {
	pond := world.NewLocation("stillwater_pond", "Fishing Pond", world.LocRoom, "stillwater", world.Position{X: 10, Y: 0})
	pond.IsOutside = true
	g.World.AddLocation(pond)

	fisher := entity.NewEntity("stillwater_fisher", "Oswin", "human", entity.RandomAttributes(g.RNG.Intn), 2, relation.CivilianRelation)
	fisher.LocationID = "stillwater_pond"
	fisher.Faction = "civilian"
	fisher.AI = entity.EntityAI{
		Type:         "scripted",
		ScriptIDs:    []string{"fisherman"},
		FactionID:    "civilian",
		HomeLocation: "stillwater_pond",
		SleepCycle:   "diurnal",
	}
	rod := lookup("fishing_rod")
	bait := lookup("bait")
	if rod != nil {
		fisher.AddItem(NewItemInstance(rod, 1))
	}
	if bait != nil {
		fisher.AddItem(NewItemInstance(bait, 3))
	}
	giveCurrency(fisher, 5, 2, 0)
	return []*entity.Entity{fisher}
}

func (g *Generator) generateNewArchetypes() []*entity.Entity {
	var all []*entity.Entity

	// Goblin gatherers — collect resources, avoid combat, but still armed
	goblinAttrs := entity.Attributes{STR: 10, DEX: 14, CON: 10, INT: 6, WIS: 8, CHA: 4}
	for i := 0; i < 2; i++ {
		goblin := entity.NewEntity("goblin_gather_"+fmt.Sprint(i), "Goblin"+fmt.Sprint(i), "goblin", goblinAttrs, 1, relation.GoblinRelation)
		goblin.LocationID = "goblin_hollow"
		goblin.Faction = "goblin"
		goblin.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"gathering"}, FactionID: "goblin", SleepCycle: "diurnal", HomeLocation: "goblin_hollow"}
		equipItem(goblin, lookup("work_tunic"))
		if g.RNG.Intn(100) < 60 {
			equipItem(goblin, lookup("goblin_shiv"))
		} else {
			equipItem(goblin, lookup("cudgel"))
		}
		giveCurrency(goblin, 1, 0, 0)
		all = append(all, goblin)
	}

	// Nature healer — a fey creature that heals injured entities
	healerAttrs := entity.Attributes{STR: 6, DEX: 10, CON: 12, INT: 14, WIS: 18, CHA: 16}
	healer := entity.NewEntity("fey_healer", "Willow", "fey", healerAttrs, 3, relation.CivilianRelation)
	healer.LocationID = "fey_glade"
	healer.Faction = "fey"
	healer.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"healing"}, FactionID: "fey", SleepCycle: "diurnal", HomeLocation: "fey_glade"}
	healer.MaxHP = 30
	healer.HP = 30
	all = append(all, healer)

	// Defensive badger — stays in its territory, attacks only when cornered
	badgerAttrs := entity.Attributes{STR: 12, DEX: 10, CON: 16, INT: 4, WIS: 10, CHA: 4}
	badger := entity.NewEntity("badger_defensive", "Brutus", "badger", badgerAttrs, 2, relation.BeastRelation)
	badger.LocationID = "bear_den"
	badger.Faction = "beast"
	badger.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"defensive"}, FactionID: "beast", SleepCycle: "diurnal", HomeLocation: "bear_den"}
	equipNaturalWeapon(badger, "badger")
	all = append(all, badger)

	return all
}

func (g *Generator) generateNewCreatures() []*entity.Entity {
	var all []*entity.Entity

	// Ravenmoor Manor - Vampire's home
	manor := world.NewLocation("ravenmoor_manor", "Ravenmoor Manor", world.LocCity, "northern_highlands", world.Position{X: 80, Y: 150})
	manor.IsOutside = false
	g.World.AddLocation(manor)

	manorHall := world.NewLocation("manor_hall", "Manorial Hall", world.LocBuilding, "ravenmoor_manor", world.Position{})
	manorHall.IsOutside = false
	g.World.AddLocation(manorHall)

	coffinChamber := world.NewLocation("coffin_chamber", "Coffin Chamber", world.LocRoom, "manor_hall", world.Position{X: 0, Y: 5})
	coffinChamber.IsOutside = false
	g.World.AddLocation(coffinChamber)

	// Count Valerius - Vampire
	vampireAttrs := entity.Attributes{STR: 18, DEX: 14, CON: 16, INT: 14, WIS: 12, CHA: 16}
	vampire := entity.NewEntity("vampire_valerius", "Count Valerius", "vampire", vampireAttrs, 14, relation.VampireRelation)
	vampire.LocationID = "coffin_chamber"
	vampire.Faction = "undead"
	vampire.MaxHP = 100
	vampire.HP = 100
	vampire.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"vampire"}, FactionID: "undead", SleepCycle: "nocturnal", HomeLocation: "coffin_chamber"}
	vampire.Profession = "count"
	equipItem(vampire, lookup("dark_robe"))
	equipItem(vampire, lookup("vampire_fang"))
	all = append(all, vampire)

	// Hag's Cottage
	hagCottage := world.NewLocation("hag_cottage", "Hag's Cottage", world.LocBuilding, "golden_plains", world.Position{X: 350, Y: 320})
	hagCottage.IsOutside = false
	hagCottage.Tags = []string{"cottage"}
	g.World.AddLocation(hagCottage)

	// Mirelda - Hag
	hagAttrs := entity.Attributes{STR: 12, DEX: 10, CON: 14, INT: 16, WIS: 18, CHA: 8}
	hag := entity.NewEntity("hag_mirelda", "Mirelda", "hag", hagAttrs, 10, relation.HagRelation)
	hag.LocationID = "hag_cottage"
	hag.Faction = "hag"
	hag.MaxHP = 60
	hag.HP = 60
	hag.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"hag"}, FactionID: "hag", SleepCycle: "diurnal", HomeLocation: "hag_cottage"}
	equipItem(hag, lookup("simple_robe"))
	all = append(all, hag)

	// Kobolds - Pack in kobold warren
	koboldAttrs := entity.Attributes{STR: 10, DEX: 14, CON: 10, INT: 8, WIS: 8, CHA: 8}
	for i := 0; i < 8; i++ {
		kobold := entity.NewEntity("kobold_"+fmt.Sprint(i), "Kobold"+fmt.Sprint(i), "kobold", koboldAttrs, 1+g.RNG.Intn(2), relation.KoboldRelation)
		kobold.LocationID = "kobold_warren"
		kobold.Faction = "kobold"
		kobold.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"kobold"}, FactionID: "kobold", SleepCycle: "diurnal", HomeLocation: "kobold_warren"}
		equipItem(kobold, lookup("work_tunic"))
		if g.RNG.Intn(100) < 25 {
			equipItem(kobold, lookup("leather_helmet"))
		}
		roll := g.RNG.Intn(100)
		switch {
		case roll < 55:
			equipItem(kobold, lookup("dagger"))
		case roll < 75:
			equipItem(kobold, lookup("short_sword"))
		case roll < 90:
			equipItem(kobold, lookup("goblin_shiv"))
		default:
			equipItem(kobold, lookup("cudgel"))
		}
		if g.RNG.Intn(100) < 20 {
			equipItem(kobold, lookup("wooden_shield"))
		}
		giveCurrency(kobold, g.RNG.Intn(5), 0, 0)
		all = append(all, kobold)
	}

	// Fairy - In fey glade
	fairyAttrs := entity.Attributes{STR: 6, DEX: 18, CON: 8, INT: 14, WIS: 16, CHA: 14}
	fairy := entity.NewEntity("fairy_sparkle", "Sparkle", "fey", fairyAttrs, 3, relation.FeyRelation)
	fairy.LocationID = "fey_glade"
	fairy.Faction = "fey"
	fairy.MaxHP = 20
	fairy.HP = 20
	fairy.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"fairy"}, FactionID: "fey", SleepCycle: "diurnal", HomeLocation: "fey_glade"}
	all = append(all, fairy)

	return all
}

func (g *Generator) generateFarms() []*entity.Entity {
	towns := []struct {
		townID string
		x, y   float64
	}{
		{"frosthold", 3, -5},
		{"stillwater", 3, -5},
		{"golden_gate", 3, -5},
	}

	farmerNames := []string{"Hugh", "Mira", "Torvin"}

	animalDefs := []struct {
		species string
		count   int
		attrs   entity.Attributes
	}{
		{"chicken", 4, entity.Attributes{STR: 4, DEX: 10, CON: 6, INT: 2, WIS: 6, CHA: 3}},
		{"pig", 2, entity.Attributes{STR: 10, DEX: 8, CON: 12, INT: 3, WIS: 6, CHA: 4}},
		{"cow", 1, entity.Attributes{STR: 14, DEX: 6, CON: 16, INT: 3, WIS: 8, CHA: 4}},
		{"sheep", 2, entity.Attributes{STR: 8, DEX: 8, CON: 12, INT: 3, WIS: 6, CHA: 4}},
		{"goat", 1, entity.Attributes{STR: 10, DEX: 10, CON: 12, INT: 3, WIS: 6, CHA: 4}},
	}

	var all []*entity.Entity
	for ti, town := range towns {
		farmID := town.townID + "_farm"
		farmLoc := world.NewLocation(farmID, town.townID+" Farm", world.LocBuilding, town.townID, world.Position{X: town.x, Y: town.y})
		farmLoc.IsOutside = false
		farmLoc.Tags = []string{"farm", "campfire"}
		g.World.AddLocation(farmLoc)

		for _, ad := range animalDefs {
			for i := 0; i < ad.count; i++ {
				animalID := farmID + "_" + ad.species + fmt.Sprint(i)
				animal := entity.NewEntity(animalID, ad.species+"_"+fmt.Sprint(i), ad.species, ad.attrs, 1, relation.FarmAnimalRelation)
				animal.LocationID = farmID
				animal.Faction = "civilian"
				animal.AI = entity.EntityAI{Type: "passive", FactionID: town.townID, HomeLocation: farmID, SleepCycle: "diurnal"}
				all = append(all, animal)
			}
		}

		dog := entity.NewEntity(farmID+"_dog", "Mutt", "dog", entity.Attributes{STR: 8, DEX: 12, CON: 10, INT: 4, WIS: 8, CHA: 6}, 1, relation.FarmAnimalRelation)
		dog.LocationID = farmID
		dog.Faction = "civilian"
		dog.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"dog"}, FactionID: town.townID, HomeLocation: farmID, SleepCycle: "diurnal"}
		all = append(all, dog)

		farmerID := town.townID + "_farmer"
		farmer := entity.NewEntity(farmerID, farmerNames[ti], "human", entity.RandomAttributes(g.RNG.Intn), 2, relation.CivilianRelation)
		farmer.LocationID = farmID
		farmer.Faction = "civilian"
		farmer.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"farmer"}, FactionID: town.townID, HomeLocation: farmID, SleepCycle: "diurnal"}
		equipItem(farmer, lookup("work_tunic"))
		equipItem(farmer, lookup("leather_boots"))
		addItem(farmer, lookup("grain"))
		addItem(farmer, lookup("grain"))
		giveCurrency(farmer, 5+g.RNG.Intn(10), 2+g.RNG.Intn(5), 0)
		all = append(all, farmer)
	}

	return all
}

func (g *Generator) generateTownExtras() []*entity.Entity {
	var all []*entity.Entity

	townIDs := []string{"frosthold", "stillwater", "golden_gate"}
	minerNames := []string{"Bramble", "Denton", "Gorin"}
	patronNames := []string{"Bran", "Piper", "Tobias"}

	for i, townID := range townIDs {
		mineID := townID + "_mine"
		mine := world.NewLocation(mineID, "Abandoned Mine", world.LocBuilding, townID, world.Position{X: -3, Y: 5})
		mine.IsOutside = false
		g.World.AddLocation(mine)

		miner := entity.NewEntity(townID+"_miner", minerNames[i], "human", entity.RandomAttributes(g.RNG.Intn), 2, relation.CivilianRelation)
		miner.LocationID = mineID
		miner.Faction = "civilian"
		miner.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"miner"}, FactionID: townID, HomeLocation: mineID, SleepCycle: "diurnal"}
		equipItem(miner, lookup("work_tunic"))
		equipItem(miner, lookup("pickaxe"))
		giveCurrency(miner, 5+g.RNG.Intn(10), 1, 0)
		all = append(all, miner)

		hutID := townID + "_herbalist_hut"
		hut := world.NewLocation(hutID, "Herbalist Hut", world.LocBuilding, townID, world.Position{X: 4, Y: -4})
		hut.IsOutside = false
		hut.Tags = []string{"campfire", "cauldron"}
		g.World.AddLocation(hut)

		herbalist := entity.NewEntity(townID+"_herbalist", "Herbalist "+fmt.Sprint(i+1), "human", entity.RandomAttributes(g.RNG.Intn), 2, relation.CivilianRelation)
		herbalist.LocationID = hutID
		herbalist.Faction = "civilian"
		herbalist.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"herbalist"}, FactionID: townID, HomeLocation: hutID, SleepCycle: "diurnal"}
		equipItem(herbalist, lookup("simple_robe"))
		equipItem(herbalist, lookup("herb_pouch"))
		herbalist.AddItem(NewItemInstance(lookup("herb"), 5))
		giveCurrency(herbalist, 5+g.RNG.Intn(10), 1, 0)
		all = append(all, herbalist)

		courier := entity.NewEntity(townID+"_courier", "Courier "+fmt.Sprint(i+1), "human", entity.RandomAttributes(g.RNG.Intn), 2, relation.CivilianRelation)
		courier.LocationID = townID + "_inn_common"
		courier.Faction = "civilian"
		courier.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"courier"}, FactionID: townID, HomeLocation: townID + "_inn_common", SleepCycle: "diurnal"}
		equipItem(courier, lookup("common_clothes"))
		equipItem(courier, lookup("leather_boots"))
		giveCurrency(courier, 2+g.RNG.Intn(5), 1, 0)
		all = append(all, courier)

		child := entity.NewEntity(townID+"_child", "Child "+fmt.Sprint(i+1), "human", entity.RandomAttributes(g.RNG.Intn), 1, relation.ChildRelation)
		child.LocationID = townID + "_inn_common"
		child.Faction = "civilian"
		child.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"child"}, FactionID: townID, HomeLocation: townID + "_inn_common", SleepCycle: "diurnal"}
		equipItem(child, lookup("common_clothes"))
		giveCurrency(child, 1, 0, 0)
		all = append(all, child)

		patron := entity.NewEntity(townID+"_patron", patronNames[i], "human", entity.RandomAttributes(g.RNG.Intn), 1, relation.CivilianRelation)
		patron.LocationID = townID + "_inn_common"
		patron.Faction = "civilian"
		patron.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"bar_patron"}, FactionID: townID, HomeLocation: townID + "_inn_common", SleepCycle: "diurnal"}
		equipItem(patron, lookup("common_clothes"))
		giveCurrency(patron, 5+g.RNG.Intn(10), 2, 0)
		all = append(all, patron)
	}

	return all
}

func (g *Generator) generateWildernessBosses() []*entity.Entity {
	var all []*entity.Entity

	// Cultist camp in the Ash Desert
	cultistCamp := world.NewLocation("cultist_camp", "Cultist Camp", world.LocBuilding, "ash_desert", world.Position{X: 10, Y: 10})
	cultistCamp.IsOutside = true
	cultistCamp.Tags = []string{"campfire"}
	g.World.AddLocation(cultistCamp)

	cultistNames := []string{"Keth", "Vorg", "Zara"}
	for i, name := range cultistNames {
		cultist := entity.NewEntity("cultist_"+fmt.Sprint(i), name, "human", entity.RandomAttributes(g.RNG.Intn), 2+g.RNG.Intn(2), relation.HagRelation)
		cultist.LocationID = "cultist_camp"
		cultist.Faction = "cultist"
		cultist.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"cultist"}, FactionID: "cultist", HomeLocation: "cultist_camp", SleepCycle: "nocturnal"}
		equipItem(cultist, lookup("common_clothes"))
		equipItem(cultist, lookup("cultist_dagger"))
		giveCurrency(cultist, 2+g.RNG.Intn(5), 1, 0)
		all = append(all, cultist)
	}

	// Werewolf cottage in the Northern Highlands
	werewolfCottage := world.NewLocation("werewolf_cottage", "Werewolf Cottage", world.LocBuilding, "northern_highlands", world.Position{X: 120, Y: 220})
	werewolfCottage.IsOutside = false
	g.World.AddLocation(werewolfCottage)

	werewolf := entity.NewEntity("werewolf_cursed", "Cursed Traveler", "human", entity.Attributes{STR: 16, DEX: 14, CON: 14, INT: 8, WIS: 10, CHA: 8}, 4, relation.CivilianRelation)
	werewolf.LocationID = "werewolf_cottage"
	werewolf.Faction = "werewolf"
	werewolf.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"werewolf"}, FactionID: "werewolf", HomeLocation: "werewolf_cottage", SleepCycle: "nocturnal"}
	werewolf.MaxHP = 50
	werewolf.HP = 50
	equipItem(werewolf, lookup("common_clothes"))
	all = append(all, werewolf)

	// Graveyard in the Golden Plains
	graveyard := world.NewLocation("golden_plains_graveyard", "Old Graveyard", world.LocBuilding, "golden_plains", world.Position{X: 320, Y: 340})
	graveyard.IsOutside = true
	graveyard.Tags = []string{"cauldron"}
	g.World.AddLocation(graveyard)

	necromancer := entity.NewEntity("necromancer_morth", "Morth the Pale", "human", entity.Attributes{STR: 10, DEX: 12, CON: 12, INT: 18, WIS: 14, CHA: 10}, 5, relation.HagRelation)
	necromancer.LocationID = "golden_plains_graveyard"
	necromancer.Faction = "undead"
	necromancer.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"necromancer"}, FactionID: "undead", HomeLocation: "golden_plains_graveyard", SleepCycle: "nocturnal"}
	necromancer.MaxHP = 45
	necromancer.HP = 45
	equipItem(necromancer, lookup("dark_robe"))
	equipItem(necromancer, lookup("necromancer_staff"))
	giveCurrency(necromancer, 5, 2, 1)
	all = append(all, necromancer)

	// Dragon lair in the Ash Desert
	dragonLair := world.NewLocation("dragon_lair", "Dragon Lair", world.LocBuilding, "ash_desert", world.Position{X: 30, Y: 30})
	dragonLair.IsOutside = false
	g.World.AddLocation(dragonLair)

	dragon := entity.NewEntity("dragon_ash", "Ashscale the Ancient", "dragon", entity.Attributes{STR: 30, DEX: 16, CON: 28, INT: 14, WIS: 16, CHA: 20}, 10, relation.HagRelation)
	dragon.LocationID = "dragon_lair"
	dragon.Faction = "dragon"
	dragon.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"dragon"}, FactionID: "dragon", HomeLocation: "dragon_lair", SleepCycle: "none"}
	dragon.MaxHP = 250
	dragon.HP = 250
	equipItem(dragon, lookup("dragon_crown"))
	dragon.AddItem(NewItemInstance(lookup("dragon_scale"), 5))
	giveCurrency(dragon, 100, 50, 20)
	all = append(all, dragon)

	// Rangers roaming the wilderness
	rangerLocs := []string{"crystal_forest", "northern_highlands", "golden_plains"}
	rangerNames := []string{"Sylas", "Rowan", "Kael"}
	for i, loc := range rangerLocs {
		ranger := entity.NewEntity("ranger_"+fmt.Sprint(i), rangerNames[i], "human", entity.Attributes{STR: 14, DEX: 16, CON: 12, INT: 10, WIS: 14, CHA: 10}, 3, relation.HagRelation)
		ranger.LocationID = loc
		ranger.Faction = "ranger"
		ranger.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"ranger"}, FactionID: "ranger", HomeLocation: loc, SleepCycle: "diurnal"}
		equipItem(ranger, lookup("leather_armor"))
		equipItem(ranger, lookup("leather_boots"))
		equipItem(ranger, lookup("short_sword"))
		giveCurrency(ranger, 5+g.RNG.Intn(10), 2, 0)
		all = append(all, ranger)
	}

	// Bandit chief in the wilderness
	banditChief := entity.NewEntity("bandit_chief", "Blackheart", "human", entity.Attributes{STR: 18, DEX: 14, CON: 16, INT: 10, WIS: 10, CHA: 12}, 5, relation.HagRelation)
	banditChief.LocationID = "golden_plains"
	banditChief.Faction = "bandit"
	banditChief.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"bandit_chief"}, FactionID: "bandit", HomeLocation: "golden_plains", SleepCycle: "diurnal"}
	banditChief.MaxHP = 70
	banditChief.HP = 70
	equipItem(banditChief, lookup("chainmail"))
	equipItem(banditChief, lookup("iron_helmet"))
	equipItem(banditChief, lookup("iron_sword"))
	giveCurrency(banditChief, 20, 10, 2)
	all = append(all, banditChief)

	return all
}

func (g *Generator) generateGoblinAmbushers() []*entity.Entity {
	var all []*entity.Entity

	goblinAttrs := entity.Attributes{STR: 10, DEX: 16, CON: 10, INT: 6, WIS: 8, CHA: 4}
	for i := 0; i < 2; i++ {
		goblin := entity.NewEntity("goblin_ambush_"+fmt.Sprint(i), "GoblinAmbush"+fmt.Sprint(i), "goblin", goblinAttrs, 1+g.RNG.Intn(2), relation.GoblinRelation)
		goblin.LocationID = "northern_highlands"
		goblin.Faction = "goblin"
		goblin.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"goblin_ambush"}, FactionID: "goblin", SleepCycle: "diurnal", HomeLocation: "northern_highlands"}
		equipItem(goblin, lookup("work_tunic"))
		equipItem(goblin, lookup("dagger"))
		goblin.AddItem(NewItemInstance(lookup("trap_kit"), 2))
		giveCurrency(goblin, g.RNG.Intn(5), 0, 0)
		all = append(all, goblin)
	}

	return all
}

func GenerateDefaultWorld() (*world.World, []*entity.Entity) {
	g := NewGenerator("default")
	return g.Generate()
}

func GenerateDefaultEntities(w *world.World, rng func(int) int) []*entity.Entity {
	_ = w.Location("frosthold")
	return nil
}
