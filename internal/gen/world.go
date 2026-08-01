// Package gen contains world generation helpers and seeded simulation setup utilities.
package gen

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"sync"

	"simuz/internal/entity"
	"simuz/internal/relation"
	"simuz/internal/world"
)

type Generator struct {
	Seed  string
	mu    sync.Mutex // Protects the non-thread-safe RNG
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

type biomeProfile struct {
	namePrefixes []string
	nameSuffixes []string
	baseClimate  world.Season
}

// Kept package-private and structurally unmutated
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

func (g *Generator) Generate() (*world.World, []*entity.Entity) {
	g.mu.Lock()
	defer g.mu.Unlock()

	log.Printf("[gen] starting world generation with seed %q", g.Seed)

	worldLoc := world.NewLocation("aetheria", "Aetheria", world.LocWorld, "", world.Position{})
	worldLoc.Weather = world.NewWeather(world.Clear, 15)
	g.World.AddLocation(worldLoc)

	numRegions := 8 + g.RNG.Intn(5)
	log.Printf("[gen] creating %d regions", numRegions)
	biomeTypes := []string{"highlands", "swamp", "plains", "forest", "waste"}

	for i := 0; i < numRegions; i++ {
		bType := biomeTypes[g.RNG.Intn(len(biomeTypes))]
		profile := biomeMatrix[bType]

		rID := fmt.Sprintf("region_%s_%d", bType, i)
		pfx := profile.namePrefixes[g.RNG.Intn(len(profile.namePrefixes))]
		sfx := profile.nameSuffixes[g.RNG.Intn(len(profile.nameSuffixes))]
		rName := fmt.Sprintf("%s %s", pfx, sfx)

		// Spiral Packing Math Logic
		angle := float64(i)*(2*math.Pi/float64(numRegions)) + (g.RNG.Float64() * 0.4)
		distance := 300.0 + (g.RNG.Float64() * 250.0)
		rx := math.Round(400.0 + distance*math.Cos(angle))
		ry := math.Round(400.0 + distance*math.Sin(angle))

		loc := world.NewLocation(rID, rName, world.LocRegion, "aetheria", world.Position{X: rx, Y: ry})
		loc.Weather = world.GenerateWeatherFor(profile.baseClimate, rID, g.RNG)
		g.World.AddLocation(loc)
		log.Printf("[gen] region %s %q biome=%s pos=(%.0f,%.0f)", rID, rName, bType, rx, ry)
	}

	g.generateWildSites()
	g.generateRegionExits()
	entities := g.generateSettlements()
	entities = append(entities, g.generateHostiles()...)
	entities = append(entities, g.generateBeasts()...)
	entities = append(entities, g.generateHydras()...)
	entities = append(entities, g.generateBasilisks()...)
	entities = append(entities, g.generateCockatrices()...)
	entities = append(entities, g.generateManticores()...)
	entities = append(entities, g.generateGriffins()...)
	entities = append(entities, g.generateWyverns()...)
	entities = append(entities, g.generateUndead()...)
	entities = append(entities, g.generateAberrations()...)
	entities = append(entities, g.generateTrolls()...)
	entities = append(entities, g.generateOgres()...)
	entities = append(entities, g.generateEttins()...)
	entities = append(entities, g.generateCyclopses()...)
	entities = append(entities, g.generateMedusas()...)
	entities = append(entities, g.generateHalfOrcs()...)
	entities = append(entities, g.generateHalfElves()...)
	entities = append(entities, g.generateHalfDwarves()...)
	entities = append(entities, g.generateHalfGoblins()...)
	entities = append(entities, g.generateHalfHobgoblins()...)
	entities = append(entities, g.generateHalfGnolls()...)
	entities = append(entities, g.generateHalfKobolds()...)
	entities = append(entities, g.generateHalfFey()...)

	log.Printf("[gen] generation complete: %d locations, %d entities", len(g.World.AllLocations()), len(entities))
	return g.World, entities
}

// generateRegionExits automatically scans the 2D grid layout, finds the closest
// geographic neighbors, and builds a logically sound bidirectional traversal mesh.
func (g *Generator) generateRegionExits() {
	var regions []*world.Location
	for _, loc := range g.World.AllLocations() {
		if loc.Type == world.LocRegion {
			regions = append(regions, loc)
		}
	}

	if len(regions) < 2 {
		log.Printf("[gen] generateRegionExits: only %d regions, skipping", len(regions))
		return
	}

	log.Printf("[gen] generateRegionExits: connecting %d regions", len(regions))
	connections := 0

	for i, rA := range regions {
		var firstClosest *world.Location
		var secondClosest *world.Location

		firstDist := math.MaxFloat64
		secondDist := math.MaxFloat64

		for j, rB := range regions {
			if i == j {
				continue
			}

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

		if firstClosest != nil {
			dirA, dirB := calculateDirections(rA.Position, firstClosest.Position)
			g.World.AddBidirectionalExit(rA.ID, firstClosest.ID, dirA, dirB)
			connections++
			log.Printf("[gen] exit: %s <-> %s (%s/%s)", rA.ID, firstClosest.ID, dirA, dirB)
		}
		if secondClosest != nil {
			dirA, dirB := calculateDirections(rA.Position, secondClosest.Position)
			g.World.AddBidirectionalExit(rA.ID, secondClosest.ID, dirA, dirB)
			connections++
			log.Printf("[gen] exit: %s <-> %s (%s/%s)", rA.ID, secondClosest.ID, dirA, dirB)
		}
	}

	log.Printf("[gen] generateRegionExits: made %d connections", connections)
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

func (g *Generator) generateHostiles() []*entity.Entity {
	log.Printf("[gen] generateHostiles: spawning hostile entities")
	var all []*entity.Entity

	// Resolve real locations from the world, creating fallback camp locations when needed
	townInns := map[string]string{
		"frosthold":          "frosthold_inn_common",
		"stillwater":         "stillwater_inn_common",
		"golden_gate":        "golden_gate_inn_common",
	}
	regionByBiome := map[string]string{}
	for _, loc := range g.World.AllLocations() {
		if loc.Type == world.LocRegion {
			key := biomeFromRegionID(loc.ID)
			if _, ok := regionByBiome[key]; !ok {
				regionByBiome[key] = loc.ID
			}
		}
	}
	orcCampRegion := regionByBiome["highlands"]
	if orcCampRegion == "" {
		orcCampRegion = regionByBiome["plains"]
	}
	banditCampRegion := regionByBiome["plains"]
	if banditCampRegion == "" {
		banditCampRegion = regionByBiome["forest"]
	}
	ashRuinsRegion := regionByBiome["waste"]
	if ashRuinsRegion == "" {
		ashRuinsRegion = regionByBiome["plains"]
	}
	feyGladeRegion := regionByBiome["forest"]
	if feyGladeRegion == "" {
		feyGladeRegion = regionByBiome["plains"]
	}

	// Create a camp location for orcs when no highlands region exists
	orcCampID := "orc_camp"
	if existing := g.World.Location(orcCampID); existing == nil {
		campLoc := world.NewLocation(orcCampID, "Orc Camp", world.LocBuilding, orcCampRegion, world.Position{X: 10, Y: 10})
		campLoc.IsOutside = true
		campLoc.Tags = []string{"campfire", "orc"}
		g.World.AddLocation(campLoc)
		log.Printf("[gen] created fallback location: %s in %s", orcCampID, orcCampRegion)
	}

	banditCampID := "bandit_camp"
	if existing := g.World.Location(banditCampID); existing == nil {
		campLoc := world.NewLocation(banditCampID, "Bandit Camp", world.LocBuilding, banditCampRegion, world.Position{X: 20, Y: -10})
		campLoc.IsOutside = true
		campLoc.Tags = []string{"campfire", "hostile"}
		g.World.AddLocation(campLoc)
		log.Printf("[gen] created fallback location: %s in %s", banditCampID, banditCampRegion)
	}

	ashRuinsID := "ash_ruins"
	if existing := g.World.Location(ashRuinsID); existing == nil {
		ruinsLoc := world.NewLocation(ashRuinsID, "Ash Ruins", world.LocBuilding, ashRuinsRegion, world.Position{X: -15, Y: 25})
		ruinsLoc.IsOutside = true
		ruinsLoc.Tags = []string{"ruins", "fire"}
		g.World.AddLocation(ruinsLoc)
		log.Printf("[gen] created fallback location: %s in %s", ashRuinsID, ashRuinsRegion)
	}

	feyGladeID := "fey_glade"
	if existing := g.World.Location(feyGladeID); existing == nil {
		gladeLoc := world.NewLocation(feyGladeID, "Fey Glade", world.LocWildSite, feyGladeRegion, world.Position{X: 5, Y: -20})
		gladeLoc.IsOutside = true
		gladeLoc.Tags = []string{"glade", "fey"}
		g.World.AddLocation(gladeLoc)
		log.Printf("[gen] created fallback location: %s in %s", feyGladeID, feyGladeRegion)
	}

	orcDefs := []struct {
		id, name, locID string
	}{
		{"orc_skar", "Skar", orcCampID},
		{"orc_grom", "Grom", orcCampID},
		{"orc_grak", "Grak", orcCampID},
		{"orc_uruk", "Uruk", ashRuinsID},
	}
	elfDefs := []struct {
		id, name, locID string
	}{
		{"elf_aerin", "Aerin", feyGladeID},
		{"elf_thalion", "Thalion", feyGladeID},
		{"elf_lyra", "Lyra", feyGladeID},
	}
	thiefDefs := []struct {
		id, name, locID string
	}{
		{"thief_rat", "Rattle", townInns["frosthold"]},
		{"thief_creep", "Creep", townInns["stillwater"]},
		{"thief_sneak", "Sneak", townInns["golden_gate"]},
	}
	banditDefs := []struct {
		id, name, locID string
	}{
		{"bandit_knife", "Knife", banditCampID},
		{"bandit_blade", "Blade", banditCampID},
		{"bandit_jack", "Jack", banditCampID},
		{"bandit_rog", "Rog", banditCampID},
	}

	log.Printf("[gen] orcs: %d in %s (1 in %s), bandits: %d in %s", len(orcDefs), orcCampID, ashRuinsID, len(banditDefs), banditCampID)

	for _, o := range orcDefs {
		ent := entity.NewEntity(o.id, o.name, "orc", entity.RandomAttributes(g.RNG.Intn), 3+g.RNG.Intn(3), relation.OrcRelation)
		ent.LocationID = o.locID
		ent.Faction = "orc"
		ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "orc", SleepCycle: "diurnal", HomeLocation: o.locID}
		equipItem(ent, lookup("leather_armor"))
		equipItem(ent, lookup("leather_boots"))
		if g.RNG.Intn(100) < 70 {
			equipItem(ent, lookup("leather_helmet"))
		}
		if g.RNG.Intn(100) < 50 {
			equipItem(ent, lookup("iron_shield"))
		} else if g.RNG.Intn(100) < 50 {
			equipItem(ent, lookup("wooden_shield"))
		}
		equipItem(ent, lookup("orc_cleaver"))
		if g.RNG.Intn(100) < 40 {
			equipItem(ent, lookup("iron_axe"))
		}
		if g.RNG.Intn(100) < 25 {
			addItem(ent, lookup("beer"))
			addItem(ent, lookup("bandage"))
		}
		giveCurrency(ent, 5+g.RNG.Intn(20), g.RNG.Intn(10), g.RNG.Intn(5))
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] orc spawned: %s at %s (level %d)", o.name, o.locID, ent.Level)
	}

	log.Printf("[gen] elves: %d", len(elfDefs))
	for _, e := range elfDefs {
		ent := entity.NewEntity(e.id, e.name, "elf", entity.RandomAttributes(g.RNG.Intn), 3+g.RNG.Intn(3), relation.ElfRelation)
		ent.LocationID = e.locID
		ent.Faction = "elf"
		ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "elf", SleepCycle: "diurnal", HomeLocation: e.locID}
		equipItem(ent, lookup("fine_clothes"))
		equipItem(ent, lookup("leather_boots"))
		equipItem(ent, lookup("leather_gloves"))
		if g.RNG.Intn(100) < 60 {
			equipItem(ent, lookup("short_sword"))
		} else {
			equipItem(ent, lookup("iron_spear"))
		}
		if g.RNG.Intn(100) < 40 {
			equipItem(ent, lookup("wooden_shield"))
		}
		if g.RNG.Intn(100) < 30 {
			addItem(ent, lookup("herb"))
			addItem(ent, lookup("herb"))
		}
		giveCurrency(ent, 10+g.RNG.Intn(20), 5+g.RNG.Intn(10), g.RNG.Intn(5))
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] elf spawned: %s at %s (level %d)", e.name, e.locID, ent.Level)
	}

	log.Printf("[gen] thieves: %d", len(thiefDefs))
	for _, t := range thiefDefs {
		ent := entity.NewEntity(t.id, t.name, "human", entity.RandomAttributes(g.RNG.Intn), 2+g.RNG.Intn(2), relation.ThiefRelation)
		ent.LocationID = t.locID
		ent.Faction = "thieves_guild"
		ent.Profession = "thief"
		ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"thief"}, FactionID: "thief", SleepCycle: "nocturnal", HomeLocation: t.locID}
		equipItem(ent, lookup("common_clothes"))
		equipItem(ent, lookup("leather_boots"))
		equipItem(ent, lookup("leather_gloves"))
		if g.RNG.Intn(100) < 70 {
			equipItem(ent, lookup("dagger"))
		} else {
			equipItem(ent, lookup("short_sword"))
		}
		giveCurrency(ent, 10+g.RNG.Intn(30), 2+g.RNG.Intn(10), 5+g.RNG.Intn(5))
		addItem(ent, lookup("beer"))
		if g.RNG.Intn(100) < 25 {
			addItem(ent, lookup("bandage"))
		}
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] thief spawned: %s at %s (level %d)", t.name, t.locID, ent.Level)

	}

	log.Printf("[gen] bandits: %d", len(banditDefs))
	for _, b := range banditDefs {
		ent := entity.NewEntity(b.id, b.name, "human", entity.RandomAttributes(g.RNG.Intn), 2+g.RNG.Intn(3), relation.BanditRelation)
		ent.LocationID = b.locID
		ent.Faction = ""
		ent.Profession = "bandit"
		ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "bandit", SleepCycle: "diurnal", HomeLocation: b.locID}
		equipItem(ent, lookup("leather_armor"))
		equipItem(ent, lookup("leather_boots"))
		equipItem(ent, lookup("leather_gloves"))
		if g.RNG.Intn(100) < 60 {
			equipItem(ent, lookup("leather_helmet"))
		}
		if g.RNG.Intn(100) < 50 {
			equipItem(ent, lookup("iron_shield"))
		}
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
		if g.RNG.Intn(100) < 30 {
			addItem(ent, lookup("bandage"))
		}
		giveCurrency(ent, 10+g.RNG.Intn(40), g.RNG.Intn(20), g.RNG.Intn(10))
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] bandit spawned: %s at %s (level %d)", b.name, b.locID, ent.Level)
	}

	// === BEASTIAL HUMANOIDS ===
	log.Printf("[gen] spawning beastial humanoids")

	bugbearDefs := []struct {
		id, name, regionID string
	}{
		{"bugbear_gorr", "Gorr", "northern_highlands"},
		{"bugbear_ash", "Ashclaw", "forest"},
		{"bugbear_mire", "Mirefang", "swamp"},
	}
	for _, bb := range bugbearDefs {
		ent := entity.NewEntity(bb.id, bb.name, "bugbear", entity.RandomAttributes(g.RNG.Intn), 3+g.RNG.Intn(3), relation.BeastRelation)
		ent.LocationID = bb.regionID
		ent.Faction = "bugbear"
		ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "bugbear", SleepCycle: "nocturnal", HomeLocation: bb.regionID}
		equipItem(ent, lookup("leather_armor"))
		equipItem(ent, lookup("leather_boots"))
		if g.RNG.Intn(100) < 60 {
			equipItem(ent, lookup("leather_helmet"))
		}
		equipItem(ent, lookup("orc_cleaver"))
		if g.RNG.Intn(100) < 30 {
			addItem(ent, lookup("bandage"))
		}
		giveCurrency(ent, 5+g.RNG.Intn(15), g.RNG.Intn(5), 0)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] bugbear spawned: %s at %s (level %d)", bb.name, bb.regionID, ent.Level)
	}

	hobgoblinDefs := []struct {
		id, name, regionID string
	}{
		{"hobgoblin_warlord", "Durgath", "northern_highlands"},
		{"hobgoblin_scout", "Skarrak", "plains"},
		{"hobgoblin_brute", "Mograth", "highlands"},
	}
	for _, hg := range hobgoblinDefs {
		ent := entity.NewEntity(hg.id, hg.name, "hobgoblin", entity.RandomAttributes(g.RNG.Intn), 3+g.RNG.Intn(3), relation.BeastRelation)
		ent.LocationID = hg.regionID
		ent.Faction = "hobgoblin"
		ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "hobgoblin", SleepCycle: "diurnal", HomeLocation: hg.regionID}
		equipItem(ent, lookup("leather_armor"))
		equipItem(ent, lookup("leather_boots"))
		if g.RNG.Intn(100) < 50 {
			equipItem(ent, lookup("iron_shield"))
		}
		equipItem(ent, lookup("short_sword"))
		if g.RNG.Intn(100) < 25 {
			equipItem(ent, lookup("leather_helmet"))
		}
		giveCurrency(ent, 5+g.RNG.Intn(15), g.RNG.Intn(5), 0)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] hobgoblin spawned: %s at %s (level %d)", hg.name, hg.regionID, ent.Level)
	}

	gnollDefs := []struct {
		id, name, regionID string
	}{
		{"gnoll_hyena", "Rippersnout", "plains"},
		{"gnoll_scavenger", "Bonepick", "waste"},
		{"gnoll_hunter", "Snappingjaw", "swamp"},
	}
	for _, gn := range gnollDefs {
		ent := entity.NewEntity(gn.id, gn.name, "gnoll", entity.RandomAttributes(g.RNG.Intn), 3+g.RNG.Intn(3), relation.BeastRelation)
		ent.LocationID = gn.regionID
		ent.Faction = "gnoll"
		ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "gnoll", SleepCycle: "nocturnal", HomeLocation: gn.regionID}
		equipItem(ent, lookup("leather_armor"))
		equipItem(ent, lookup("leather_boots"))
		equipItem(ent, lookup("iron_spear"))
		if g.RNG.Intn(100) < 40 {
			addItem(ent, lookup("meat"))
		}
		giveCurrency(ent, g.RNG.Intn(8), 0, 0)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] gnoll spawned: %s at %s (level %d)", gn.name, gn.regionID, ent.Level)
	}

	lizardfolkDefs := []struct {
		id, name, regionID string
	}{
		{"lizardfolk_spear", "Swamasst", "swamp"},
		{"lizardfolk_hunter", "Venomscale", "swamp"},
		{"lizardfolk_shaman", "Sunscale", "waste"},
	}
	for _, lf := range lizardfolkDefs {
		ent := entity.NewEntity(lf.id, lf.name, "lizardfolk", entity.RandomAttributes(g.RNG.Intn), 3+g.RNG.Intn(3), relation.BeastRelation)
		ent.LocationID = lf.regionID
		ent.Faction = "lizardfolk"
		ent.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "lizardfolk", SleepCycle: "diurnal", HomeLocation: lf.regionID}
		equipItem(ent, lookup("work_tunic"))
		equipItem(ent, lookup("leather_boots"))
		equipItem(ent, lookup("iron_spear"))
		if g.RNG.Intn(100) < 40 {
			equipItem(ent, lookup("wooden_shield"))
		}
		if g.RNG.Intn(100) < 25 {
			addItem(ent, lookup("herb"))
		}
		giveCurrency(ent, 3+g.RNG.Intn(8), 0, 0)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] lizardfolk spawned: %s at %s (level %d)", lf.name, lf.regionID, ent.Level)
	}

	log.Printf("[gen] generateHostiles: done, total %d hostiles", len(all))
	return all
}

func (g *Generator) generateBeasts() []*entity.Entity {
	log.Printf("[gen] generateBeasts: spawning beast entities")
	beastSpawns := []struct {
		id, name, species, biomePref string
		level                        int
		attrs                        entity.Attributes
		nocturnal                    bool
	}{
		{"wolf_shadow", "Shadow", "wolf", "forest", 2, entity.Attributes{STR: 14, DEX: 16, CON: 12, INT: 4, WIS: 8, CHA: 4}, false},
		{"wolf_fang", "Silverfang", "wolf", "highlands", 3, entity.Attributes{STR: 15, DEX: 15, CON: 13, INT: 4, WIS: 8, CHA: 4}, false},
		{"wolf_snarl", "Snarl", "wolf", "forest", 2, entity.Attributes{STR: 13, DEX: 17, CON: 11, INT: 4, WIS: 8, CHA: 4}, false},
		{"bear_brown", "Bruin", "bear", "highlands", 5, entity.Attributes{STR: 20, DEX: 10, CON: 18, INT: 3, WIS: 7, CHA: 3}, false},
		{"bear_black", "Ursa", "bear", "forest", 4, entity.Attributes{STR: 18, DEX: 11, CON: 16, INT: 3, WIS: 7, CHA: 3}, false},
		{"boar_tusk", "Tusker", "boar", "plains", 3, entity.Attributes{STR: 16, DEX: 12, CON: 15, INT: 3, WIS: 6, CHA: 3}, false},
		{"boar_bristle", "Bristle", "boar", "plains", 2, entity.Attributes{STR: 14, DEX: 13, CON: 14, INT: 3, WIS: 6, CHA: 3}, false},
		{"bat_swoop", "Swoop", "bat", "waste", 1, entity.Attributes{STR: 5, DEX: 18, CON: 6, INT: 2, WIS: 10, CHA: 2}, true},
		{"spider_fang", "Webspinner", "spider", "swamp", 3, entity.Attributes{STR: 12, DEX: 16, CON: 10, INT: 2, WIS: 8, CHA: 2}, true},
		{"spider_web", "Silkweaver", "spider", "swamp", 2, entity.Attributes{STR: 10, DEX: 17, CON: 9, INT: 2, WIS: 8, CHA: 2}, true},
	}

	var all []*entity.Entity
	for _, b := range beastSpawns {
		ent := entity.NewEntity(b.id, b.name, b.species, b.attrs, b.level, relation.BeastRelation)

		targetLocID := g.findMatchingRegion(b.biomePref)
		ent.LocationID = targetLocID
		ent.Faction = "beast"

		cycle := "diurnal"
		if b.nocturnal {
			cycle = "nocturnal"
		}

		var script string
		switch {
		case b.species == "wolf":
			script = "hunting"
		case b.nocturnal:
			script = "scouting"
		default:
			script = "aggressive"
		}

		ent.AI = entity.EntityAI{
			Type:         "scripted",
			ScriptIDs:    []string{script},
			FactionID:    "beast",
			SleepCycle:   cycle,
			HomeLocation: targetLocID,
		}

		equipNaturalWeapon(ent, b.species)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
	}
	log.Printf("[gen] generateBeasts: done, total %d beasts", len(all))
	return all
}

// Helper to prevent hardcoded locID mismatches against your generated world grid
func (g *Generator) findMatchingRegion(biomePref string) string {
	for _, loc := range g.World.AllLocations() {
		if strings.Contains(loc.ID, fmt.Sprintf("region_%s_", biomePref)) {
			return loc.ID
		}
	}
	log.Printf("[gen] findMatchingRegion: no region found for biome %q, falling back to aetheria", biomePref)
	return "aetheria" // Fallback target container
}

// biomeFromRegionID extracts the biome key from a region ID
// (e.g., "region_highlands_0" -> "highlands"). Falls back to "plains".
func biomeFromRegionID(id string) string {
	for key := range biomeMatrix {
		if len(id) >= 7+len(key) && id[7:7+len(key)] == key {
			return key
		}
	}
	return "plains"
}

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
	log.Printf("[gen] generateTown: %q id=%s region=%s pos=(%.0f,%.0f)", name, id, regionID, x, y)
	town := world.NewLocation(id, name, world.LocCity, regionID, world.Position{X: x, Y: y})
	town.IsOutside = false
	g.World.AddLocation(town)

	buildings := []struct {
		id, bname, btype string
		bx, by           float64 // Relative offsets from town center (x, y)
	}{
		{id + "_inn", "The Sleeping Dragon", "inn", 0, 0},
		{id + "_temple", "Temple of Light", "temple", -5, -3},
		{id + "_guardhouse", "Guardhouse", "guardhouse", 2, -2},
		{id + "_blacksmith", "The Iron Anvil", "blacksmith", -2, 4},
		{id + "_market", "Open Plaza Market", "market", 4, 2}, // Included missing market
	}

	for _, b := range buildings {
		// FIX: Apply the town's (x, y) offset so buildings stay nested within the town geometry
		loc := world.NewLocation(b.id, b.bname, world.LocBuilding, id, world.Position{
			X: x + b.bx,
			Y: y + b.by,
		})
		loc.IsOutside = false

		// FIX: Switch cleanly on static types rather than dynamic concatenated strings
		switch b.btype {
		case "blacksmith":
			loc.Tags = []string{"forge", "blacksmith"}
		case "temple":
			loc.Tags = []string{"cauldron", "temple"}
		case "guardhouse":
			loc.Tags = []string{"workbench", "guardhouse"}
		case "inn":
			loc.Tags = []string{"inn"}
		case "market":
			loc.Tags = []string{"market"}
		}
		log.Printf("[gen] building: %q type=%s pos=(%.0f,%.0f)", b.bname, b.btype, x+b.bx, y+b.by)
		g.World.AddLocation(loc)
	}

	g.generateRooms(id)

	npcCount := 0
	npcs := g.generateNPCs(id, "human")
	npcCount = len(npcs)
	log.Printf("[gen] generateTown %q: done, %d NPCs spawned", name, npcCount)
	return npcs
}
func (g *Generator) generateRooms(innID string) {
	log.Printf("[gen] generateRooms: inn=%s", innID)
	rooms := []struct {
		id, name string
		x, y     float64
	}{
		{innID + "_common", "Common Room", 0, 0},
		{innID + "_cellar", "Cellar", 3, 0},
		{innID + "_kitchen", "Kitchen", 0, 3},
	}

	for _, r := range rooms {
		log.Printf("[gen] room: %q %s", r.name, r.id)
		loc := world.NewLocation(r.id, r.name, world.LocRoom, innID, world.Position{X: r.x, Y: r.y})
		loc.IsOutside = false
		if r.id == innID+"_common" {
			loc.Tags = []string{"inn"}
		}
		g.World.AddLocation(loc)
	}
}

func (g *Generator) generateTravelingSalesmen() []*entity.Entity {
	log.Printf("[gen] generateTravelingSalesmen: spawning traveling merchants")
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
		log.Printf("[gen] traveling_salesman: %s at %s", s.name, s.locID)
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
		itemCount := 0
		for _, itemID := range tradeItems {
			def := lookup(itemID)
			if def != nil && g.RNG.Intn(100) < 60 {
				qty := 1 + g.RNG.Intn(3)
				for k := 0; k < qty; k++ {
					addItem(ent, def)
					itemCount++
				}
			}
		}
		log.Printf("[gen] traveling_salesman %s: %d trade items", s.name, itemCount)
		entities = append(entities, ent)
	}
	log.Printf("[gen] generateTravelingSalesmen: done, total %d", len(entities))
	return entities
}

func (g *Generator) generateBards() []*entity.Entity {
	log.Printf("[gen] generateBards: spawning wandering bards")
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

		log.Printf("[gen] bard: %s (%s) quality=%s at %s", b.name, b.instrument, b.quality, b.startLoc)

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
	log.Printf("[gen] generateBards: done, total %d", len(entities))
	return entities
}

func (g *Generator) generateRatKingLair() []*entity.Entity {
	log.Printf("[gen] generateRatKingLair: creating rat king lair")
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
	log.Printf("[gen] rat_king_lair: spawning 2 rats in entrance")
	for i := 0; i < 2; i++ {
		rat := entity.NewEntity("rat_scram"+fmt.Sprint(i), "Scram", "rat", ratAttrs1, 1, relation.VerminRelation)
		rat.LocationID = "rat_king_lair_entrance"
		rat.Faction = "vermin"
		rat.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"defensive"}, FactionID: "vermin", SleepCycle: "none", HomeLocation: "rat_king_lair_entrance"}
		all = append(all, rat)
	}
	log.Printf("[gen] rat_king_lair: spawning 2 rats in corridor")
	for i := 0; i < 2; i++ {
		rat := entity.NewEntity("rat_gleam"+fmt.Sprint(i), "Gleam", "rat", ratAttrs2, 2, relation.VerminRelation)
		rat.LocationID = "rat_king_lair_corridor"
		rat.Faction = "vermin"
		rat.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"defensive"}, FactionID: "vermin", SleepCycle: "none", HomeLocation: "rat_king_lair_corridor"}
		all = append(all, rat)
	}
	log.Printf("[gen] rat_king_lair: spawning 2 scouts at entrance")
	for i := 0; i < 2; i++ {
		scout := entity.NewEntity("rat_scout"+fmt.Sprint(i), "Rat_Scout", "rat", ratAttrs1, 1, relation.VerminRelation)
		scout.LocationID = "rat_king_lair_entrance"
		scout.Faction = "vermin"
		scout.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"scouting"}, FactionID: "vermin", SleepCycle: "none", HomeLocation: "rat_king_lair_entrance"}
		all = append(all, scout)
	}

	log.Printf("[gen] rat_king_lair: spawning Rat King at throne")
	ratKing := entity.NewEntity("rat_king", "Skreet the Unseen", "rat_king", entity.Attributes{STR: 18, DEX: 10, CON: 20, INT: 8, WIS: 12, CHA: 6}, 8, relation.VerminRelation)
	ratKing.LocationID = "rat_king_lair_throne"
	ratKing.Faction = "vermin"
	ratKing.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"rat_king"}, FactionID: "vermin", SleepCycle: "none"}
	ratKing.MaxHP = 200
	ratKing.HP = 200
	ratKing.Immortal = false
	all = append(all, ratKing)

	log.Printf("[gen] generateRatKingLair: done, total %d entities", len(all))
	return all
}

func (g *Generator) generateFisherman() []*entity.Entity {
	log.Printf("[gen] generateFisherman: spawning fisherman at stillwater")
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
	log.Printf("[gen] generateFisherman: done, spawned Oswin at stillwater_pond")
	return []*entity.Entity{fisher}
}

func (g *Generator) generateNewArchetypes() []*entity.Entity {
	log.Printf("[gen] generateNewArchetypes: spawning goblins, healer, and badger")
	var all []*entity.Entity
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
		log.Printf("[gen] goblin_gatherer spawned: %s", goblin.Name)
	}

	healerAttrs := entity.Attributes{STR: 6, DEX: 10, CON: 12, INT: 14, WIS: 18, CHA: 16}
	healer := entity.NewEntity("fey_healer", "Willow", "fey", healerAttrs, 3, relation.CivilianRelation)
	healer.LocationID = "fey_glade"
	healer.Faction = "fey"
	healer.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"healing"}, FactionID: "fey", SleepCycle: "diurnal", HomeLocation: "fey_glade"}
	healer.MaxHP = 30
	healer.HP = 30
	all = append(all, healer)
	log.Printf("[gen] fey_healer spawned: Willow at fey_glade")

	badgerAttrs := entity.Attributes{STR: 12, DEX: 10, CON: 16, INT: 4, WIS: 10, CHA: 4}
	badger := entity.NewEntity("badger_defensive", "Brutus", "badger", badgerAttrs, 2, relation.BeastRelation)
	badger.LocationID = "bear_den"
	badger.Faction = "beast"
	badger.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"defensive"}, FactionID: "beast", SleepCycle: "diurnal", HomeLocation: "bear_den"}
	equipNaturalWeapon(badger, "badger")
	all = append(all, badger)
	log.Printf("[gen] defensive_badger spawned: Brutus at bear_den")

	log.Printf("[gen] generateNewArchetypes: done, total %d", len(all))
	return all
}

func (g *Generator) generateNewCreatures() []*entity.Entity {
	log.Printf("[gen] generateNewCreatures: spawning vampire, hag, kobolds, and fairy")
	var all []*entity.Entity

	manor := world.NewLocation("ravenmoor_manor", "Ravenmoor Manor", world.LocCity, "northern_highlands", world.Position{X: 80, Y: 150})
	manor.IsOutside = false
	g.World.AddLocation(manor)

	manorHall := world.NewLocation("manor_hall", "Manorial Hall", world.LocBuilding, "ravenmoor_manor", world.Position{})
	manorHall.IsOutside = false
	g.World.AddLocation(manorHall)

	coffinChamber := world.NewLocation("coffin_chamber", "Coffin Chamber", world.LocRoom, "manor_hall", world.Position{X: 0, Y: 5})
	coffinChamber.IsOutside = false
	g.World.AddLocation(coffinChamber)

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
	log.Printf("[gen] vampire spawned: Count Valerius at coffin_chamber")

	hagCottage := world.NewLocation("hag_cottage", "Hag's Cottage", world.LocBuilding, "golden_plains", world.Position{X: 350, Y: 320})
	hagCottage.IsOutside = false
	hagCottage.Tags = []string{"cottage"}
	g.World.AddLocation(hagCottage)

	hagAttrs := entity.Attributes{STR: 12, DEX: 10, CON: 14, INT: 16, WIS: 18, CHA: 8}
	hag := entity.NewEntity("hag_mirelda", "Mirelda", "hag", hagAttrs, 10, relation.HagRelation)
	hag.LocationID = "hag_cottage"
	hag.Faction = "hag"
	hag.MaxHP = 60
	hag.HP = 60
	hag.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"hag"}, FactionID: "hag", SleepCycle: "diurnal", HomeLocation: "hag_cottage"}
	equipItem(hag, lookup("simple_robe"))
	all = append(all, hag)
	log.Printf("[gen] hag spawned: Mirelda at hag_cottage")

	koboldAttrs := entity.Attributes{STR: 10, DEX: 14, CON: 10, INT: 8, WIS: 8, CHA: 8}
	log.Printf("[gen] spawning 8 kobolds at kobold_warren")
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

	fairyAttrs := entity.Attributes{STR: 6, DEX: 18, CON: 8, INT: 14, WIS: 16, CHA: 14}
	fairy := entity.NewEntity("fairy_sparkle", "Sparkle", "fey", fairyAttrs, 3, relation.FeyRelation)
	fairy.LocationID = "fey_glade"
	fairy.Faction = "fey"
	fairy.MaxHP = 20
	fairy.HP = 20
	fairy.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"fairy"}, FactionID: "fey", SleepCycle: "diurnal", HomeLocation: "fey_glade"}
	all = append(all, fairy)
	log.Printf("[gen] fairy spawned: Sparkle at fey_glade")

	log.Printf("[gen] generateNewCreatures: done, total %d", len(all))
	return all
}

func (g *Generator) generateFarms() []*entity.Entity {
	log.Printf("[gen] generateFarms: spawning farms with farmers and animals")
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
		log.Printf("[gen] farm: %s farm at pos=(%.0f,%.0f)", farmID, town.x, town.y)

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
		log.Printf("[gen] farmer spawned: %s at %s", farmerNames[ti], farmID)
	}

	log.Printf("[gen] generateFarms: done, total %d entities", len(all))
	return all
}

func (g *Generator) generateTownExtras() []*entity.Entity {
	log.Printf("[gen] generateTownExtras: spawning miners, herbalists, couriers, children, patrons")
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
		log.Printf("[gen] miner spawned: %s at %s", minerNames[i], mineID)

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
		log.Printf("[gen] herbalist spawned: %s at %s", herbalist.Name, hutID)

		courier := entity.NewEntity(townID+"_courier", "Courier "+fmt.Sprint(i+1), "human", entity.RandomAttributes(g.RNG.Intn), 2, relation.CivilianRelation)
		courier.LocationID = townID + "_inn_common"
		courier.Faction = "civilian"
		courier.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"courier"}, FactionID: townID, HomeLocation: townID + "_inn_common", SleepCycle: "diurnal"}
		equipItem(courier, lookup("common_clothes"))
		equipItem(courier, lookup("leather_boots"))
		giveCurrency(courier, 2+g.RNG.Intn(5), 1, 0)
		all = append(all, courier)
		log.Printf("[gen] courier spawned: %s at %s_inn_common", courier.Name, townID)

		child := entity.NewEntity(townID+"_child", "Child "+fmt.Sprint(i+1), "human", entity.RandomAttributes(g.RNG.Intn), 1, relation.ChildRelation)
		child.LocationID = townID + "_inn_common"
		child.Faction = "civilian"
		child.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"child"}, FactionID: townID, HomeLocation: townID + "_inn_common", SleepCycle: "diurnal"}
		equipItem(child, lookup("common_clothes"))
		giveCurrency(child, 1, 0, 0)
		all = append(all, child)
		log.Printf("[gen] child spawned: %s at %s_inn_common", child.Name, townID)

		patron := entity.NewEntity(townID+"_patron", patronNames[i], "human", entity.RandomAttributes(g.RNG.Intn), 1, relation.CivilianRelation)
		patron.LocationID = townID + "_inn_common"
		patron.Faction = "civilian"
		patron.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"bar_patron"}, FactionID: townID, HomeLocation: townID + "_inn_common", SleepCycle: "diurnal"}
		equipItem(patron, lookup("common_clothes"))
		giveCurrency(patron, 5+g.RNG.Intn(10), 2, 0)
		all = append(all, patron)
		log.Printf("[gen] patron spawned: %s at %s_inn_common", patron.Name, townID)
	}

	log.Printf("[gen] generateTownExtras: done, total %d entities", len(all))
	return all
}

func (g *Generator) generateWildernessBosses() []*entity.Entity {
	log.Printf("[gen] generateWildernessBosses: spawning bosses and NPCs")
	var all []*entity.Entity

	cultistCamp := world.NewLocation("cultist_camp", "Cultist Camp", world.LocBuilding, "ash_desert", world.Position{X: 10, Y: 10})
	cultistCamp.IsOutside = true
	cultistCamp.Tags = []string{"campfire"}
	g.World.AddLocation(cultistCamp)

	cultistNames := []string{"Keth", "Vorg", "Zara"}
	log.Printf("[gen] spawning %d cultists at cultist_camp", len(cultistNames))
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
	log.Printf("[gen] werewolf spawned: Cursed Traveler at werewolf_cottage")

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
	log.Printf("[gen] necromancer spawned: Morth the Pale at golden_plains_graveyard")

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
	log.Printf("[gen] dragon spawned: Ashscale the Ancient at dragon_lair")

	rangerLocs := []string{"crystal_forest", "northern_highlands", "golden_plains"}
	rangerNames := []string{"Sylas", "Rowan", "Kael"}
	log.Printf("[gen] spawning %d rangers", len(rangerNames))
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
		log.Printf("[gen] ranger spawned: %s at %s", rangerNames[i], loc)
	}

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
	log.Printf("[gen] bandit_chief spawned: Blackheart at golden_plains")

	log.Printf("[gen] generateWildernessBosses: done, total %d", len(all))
	return all
}

func (g *Generator) generateGoblinAmbushers() []*entity.Entity {
	log.Printf("[gen] generateGoblinAmbushers: spawning goblin ambushers in northern_highlands")
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
		log.Printf("[gen] goblin_ambusher spawned: %s", goblin.Name)
	}

	log.Printf("[gen] generateGoblinAmbushers: done, total %d", len(all))
	return all
}

// === DEDICATED BESTIAL MONSTER GENERATORS ===

func (g *Generator) generateHydras() []*entity.Entity {
	log.Printf("[gen] generateHydras: spawning hydras in swamps")
	var all []*entity.Entity
	hydraAttrs := entity.Attributes{STR: 20, DEX: 10, CON: 20, INT: 4, WIS: 11, CHA: 5}
	regions := []string{}
	for _, loc := range g.World.AllLocations() {
		if loc.Type == world.LocRegion {
			biome := biomeFromRegionID(loc.ID)
			if biome == "swamp" {
				regions = append(regions, loc.ID)
			}
		}
	}
	if len(regions) == 0 {
		regions = []string{"aetheria"}
	}
	for i := 0; i < 2; i++ {
		reg := regions[i%len(regions)]
		hydra := entity.NewEntity("hydra_bog_"+fmt.Sprint(i), "Bog Hydra", "hydra", hydraAttrs, 7, relation.BeastRelation)
		hydra.LocationID = reg
		hydra.Faction = "beast"
		hydra.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "beast", SleepCycle: "none", HomeLocation: reg}
		hydra.MaxHP = 120
		hydra.HP = 120
		equipItem(hydra, lookup("swamp_breath"))
		all = append(all, hydra)
		log.Printf("[gen] hydra spawned: Bog Hydra at %s", reg)
	}
	log.Printf("[gen] generateHydras: done, total %d", len(all))
	return all
}

func (g *Generator) generateBasilisks() []*entity.Entity {
	log.Printf("[gen] generateBasilisks: spawning basilisks in deserts and rocky wastes")
	var all []*entity.Entity
	basiliskAttrs := entity.Attributes{STR: 14, DEX: 8, CON: 16, INT: 2, WIS: 10, CHA: 6}
	for i := 0; i < 3; i++ {
		locID := "waste_basilisk_" + fmt.Sprint(i)
		loc := world.NewLocation(locID, "Petrifying Lair", world.LocWildSite, "region_waste_0", world.Position{X: float64(50 + i*30), Y: float64(50 + i*20)})
		loc.IsOutside = true
		loc.Tags = []string{"den", "beast", "petrifying"}
		g.World.AddLocation(loc)
		basilisk := entity.NewEntity("basilisk_"+fmt.Sprint(i), "Stonegaze", "basilisk", basiliskAttrs, 5, relation.BeastRelation)
		basilisk.LocationID = locID
		basilisk.Faction = "beast"
		basilisk.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "beast", SleepCycle: "nocturnal", HomeLocation: locID}
		all = append(all, basilisk)
		log.Printf("[gen] basilisk spawned: Stonegaze at %s", locID)
	}
	log.Printf("[gen] generateBasilisks: done, total %d", len(all))
	return all
}

func (g *Generator) generateCockatrices() []*entity.Entity {
	log.Printf("[gen] generateCockatrices: spawning cockatrices near farms and forests")
	var all []*entity.Entity
	cockatriceAttrs := entity.Attributes{STR: 6, DEX: 12, CON: 11, INT: 2, WIS: 9, CHA: 5}
	farmRegions := []string{"frosthold_farm", "stillwater_farm", "golden_gate_farm"}
	for i := 0; i < 4; i++ {
		reg := farmRegions[i%len(farmRegions)]
		locID := reg + "_cockatrice_" + fmt.Sprint(i)
		loc := world.NewLocation(locID, "Cockatrice Nest", world.LocWildSite, reg, world.Position{X: float64(8 + i*5), Y: float64(-3 - i*3)})
		loc.IsOutside = true
		loc.Tags = []string{"den", "beast"}
		g.World.AddLocation(loc)
		cock := entity.NewEntity("cockatrice_"+fmt.Sprint(i), "Pettrix", "cockatrice", cockatriceAttrs, 4, relation.BeastRelation)
		cock.LocationID = locID
		cock.Faction = "beast"
		cock.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "beast", SleepCycle: "nocturnal", HomeLocation: locID}
		all = append(all, cock)
		log.Printf("[gen] cockatrice spawned: Pettrix at %s", locID)
	}
	log.Printf("[gen] generateCockatrices: done, total %d", len(all))
	return all
}

func (g *Generator) generateManticores() []*entity.Entity {
	log.Printf("[gen] generateManticores: spawning manticores in forests and wastes")
	var all []*entity.Entity
	manticoreAttrs := entity.Attributes{STR: 17, DEX: 13, CON: 15, INT: 7, WIS: 10, CHA: 8}
	regions := []string{}
	for _, loc := range g.World.AllLocations() {
		if loc.Type == world.LocRegion {
			biome := biomeFromRegionID(loc.ID)
			if biome == "forest" || biome == "waste" {
				regions = append(regions, loc.ID)
			}
		}
	}
	if len(regions) == 0 {
		regions = []string{"aetheria"}
	}
	for i := 0; i < 3; i++ {
		reg := regions[i%len(regions)]
		locID := reg + "_manticore_" + fmt.Sprint(i)
		loc := world.NewLocation(locID, "Manticore Den", world.LocWildSite, reg, world.Position{X: float64(30 + i*40), Y: float64(-20 - i*30)})
		loc.IsOutside = true
		loc.Tags = []string{"den", "beast", "venomous"}
		g.World.AddLocation(loc)
		man := entity.NewEntity("manticore_"+fmt.Sprint(i), "Spikeclaw", "manticore", manticoreAttrs, 7, relation.BeastRelation)
		man.LocationID = locID
		man.Faction = "beast"
		man.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "beast", SleepCycle: "nocturnal", HomeLocation: locID}
		all = append(all, man)
		log.Printf("[gen] manticore spawned: Spikeclaw at %s", locID)
	}
	log.Printf("[gen] generateManticores: done, total %d", len(all))
	return all
}

func (g *Generator) generateGriffins() []*entity.Entity {
	log.Printf("[gen] generateGriffins: spawning griffins in highlands and mountains")
	var all []*entity.Entity
	griffinAttrs := entity.Attributes{STR: 17, DEX: 14, CON: 15, INT: 5, WIS: 12, CHA: 8}
	for i := 0; i < 2; i++ {
		locID := "highlands_griffin_" + fmt.Sprint(i)
		loc := world.NewLocation(locID, "Griffin Aerie", world.LocWildSite, "region_highlands_0", world.Position{X: float64(200 + i*60), Y: float64(-100 - i*40)})
		loc.IsOutside = true
		loc.Tags = []string{"nest", "beast"}
		g.World.AddLocation(loc)
		griffin := entity.NewEntity("griffin_gold_"+fmt.Sprint(i), "Goldwing", "griffin", griffinAttrs, 8, relation.BeastRelation)
		griffin.LocationID = locID
		griffin.Faction = "beast"
		griffin.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "beast", SleepCycle: "diurnal", HomeLocation: locID}
		all = append(all, griffin)
		log.Printf("[gen] griffin spawned: Goldwing at %s", locID)
	}
	log.Printf("[gen] generateGriffins: done, total %d", len(all))
	return all
}

func (g *Generator) generateWyverns() []*entity.Entity {
	log.Printf("[gen] generateWyverns: spawning wyverns in cliffs and mountains")
	var all []*entity.Entity
	wyvernAttrs := entity.Attributes{STR: 18, DEX: 13, CON: 16, INT: 6, WIS: 12, CHA: 8}
	for i := 0; i < 3; i++ {
		locID := "highlands_wyvern_" + fmt.Sprint(i)
		loc := world.NewLocation(locID, "Wyvern Perch", world.LocWildSite, "region_highlands_0", world.Position{X: float64(250 + i*50), Y: float64(80 + i*60)})
		loc.IsOutside = true
		loc.Tags = []string{"nest", "beast", "venomous"}
		g.World.AddLocation(loc)
		wyvern := entity.NewEntity("wyvern_venom_"+fmt.Sprint(i), "Venomwing", "wyvern", wyvernAttrs, 6, relation.BeastRelation)
		wyvern.LocationID = locID
		wyvern.Faction = "beast"
		wyvern.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "beast", SleepCycle: "nocturnal", HomeLocation: locID}
		all = append(all, wyvern)
		log.Printf("[gen] wyvern spawned: Venomwing at %s", locID)
	}
	log.Printf("[gen] generateWyverns: done, total %d", len(all))
	return all
}

func (g *Generator) generateUndead() []*entity.Entity {
	log.Printf("[gen] generateUndead: spawning undead in ruins and crypts")
	var all []*entity.Entity

	graveyardLoc := world.NewLocation("old_graveyard", "Old Graveyard", world.LocBuilding, "golden_plains", world.Position{X: 320, Y: 340})
	graveyardLoc.IsOutside = true
	graveyardLoc.Tags = []string{"graveyard", "undead_node"}
	g.World.AddLocation(graveyardLoc)

	skeletonAttrs := entity.Attributes{STR: 11, DEX: 12, CON: 0, INT: 3, WIS: 8, CHA: 3}
	for i := 0; i < 6; i++ {
		skel := entity.NewEntity("skeleton_"+fmt.Sprint(i), "Rattle", "skeleton", skeletonAttrs, 1, relation.BeastRelation)
		skel.LocationID = "old_graveyard"
		skel.Faction = "undead"
		skel.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "undead", SleepCycle: "none", HomeLocation: "old_graveyard"}
		all = append(all, skel)
		log.Printf("[gen] skeleton spawned: Rattle at old_graveyard")
	}

	zombieAttrs := entity.Attributes{STR: 13, DEX: 6, CON: 0, INT: 2, WIS: 6, CHA: 3}
	for i := 0; i < 4; i++ {
		zom := entity.NewEntity("zombie_"+fmt.Sprint(i), "Rotter", "zombie", zombieAttrs, 1, relation.BeastRelation)
		zom.LocationID = "old_graveyard"
		zom.Faction = "undead"
		zom.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "undead", SleepCycle: "none", HomeLocation: "old_graveyard"}
		all = append(all, zom)
		log.Printf("[gen] zombie spawned: Rotter at old_graveyard")
	}

	ghoulAttrs := entity.Attributes{STR: 13, DEX: 14, CON: 0, INT: 9, WIS: 11, CHA: 6}
	for i := 0; i < 3; i++ {
		ghoul := entity.NewEntity("ghoul_"+fmt.Sprint(i), "Carnor", "ghoul", ghoulAttrs, 2, relation.BeastRelation)
		ghoul.LocationID = "old_graveyard"
		ghoul.Faction = "undead"
		ghoul.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "undead", SleepCycle: "nocturnal", HomeLocation: "old_graveyard"}
		all = append(all, ghoul)
		log.Printf("[gen] ghoul spawned: Carnor at old_graveyard")
	}

	lichAttrs := entity.Attributes{STR: 12, DEX: 11, CON: 0, INT: 22, WIS: 18, CHA: 16}
	lich := entity.NewEntity("lich_korvinus", "Korvinus the Undying", "lich", lichAttrs, 10, relation.BeastRelation)
	lich.LocationID = "old_graveyard"
	lich.Faction = "undead"
	lich.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"necromancer"}, FactionID: "undead", SleepCycle: "none", HomeLocation: "old_graveyard"}
	lich.MaxHP = 100
	lich.HP = 100
	all = append(all, lich)
	log.Printf("[gen] lich spawned: Korvinus the Undying at old_graveyard")

	wraithAttrs := entity.Attributes{STR: 0, DEX: 16, CON: 0, INT: 12, WIS: 12, CHA: 15}
	for i := 0; i < 2; i++ {
		wraith := entity.NewEntity("wraith_"+fmt.Sprint(i), "Shade", "wraith", wraithAttrs, 4, relation.BeastRelation)
		wraith.LocationID = "old_graveyard"
		wraith.Faction = "undead"
		wraith.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "undead", SleepCycle: "none", HomeLocation: "old_graveyard"}
		all = append(all, wraith)
		log.Printf("[gen] wraith spawned: Shade at old_graveyard")
	}

	log.Printf("[gen] generateUndead: done, total %d", len(all))
	return all
}

func (g *Generator) generateAberrations() []*entity.Entity {
	log.Printf("[gen] generateAberrations: spawning mind flayers, beholders, and floating eyes")
	var all []*entity.Entity

	underdarkLoc := world.NewLocation("underdark_lair", "Underdark Lair", world.LocSubTerranean, "region_waste_0", world.Position{X: 500, Y: 20})
	underdarkLoc.IsOutside = false
	underdarkLoc.Tags = []string{"underdark", "aberration"}
	g.World.AddLocation(underdarkLoc)

	mindFlayerAttrs := entity.Attributes{STR: 11, DEX: 12, CON: 12, INT: 19, WIS: 17, CHA: 17}
	mindFlayer := entity.NewEntity("mind_flayer_xyrketh", "Xyr'Keth", "mind_flayer", mindFlayerAttrs, 10, relation.BeastRelation)
	mindFlayer.LocationID = "underdark_lair"
	mindFlayer.Faction = "aberration"
	mindFlayer.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "aberration", SleepCycle: "none", HomeLocation: "underdark_lair"}
	mindFlayer.MaxHP = 99
	mindFlayer.HP = 99
	all = append(all, mindFlayer)
	log.Printf("[gen] mind_flayer spawned: Xyr'Keth at underdark_lair")

	beholderAttrs := entity.Attributes{STR: 10, DEX: 14, CON: 18, INT: 17, WIS: 15, CHA: 17}
	beholder := entity.NewEntity("beholder_orbius", "Orbius the Eye Tyrant", "beholder", beholderAttrs, 13, relation.BeastRelation)
	beholder.LocationID = "underdark_lair"
	beholder.Faction = "aberration"
	beholder.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "aberration", SleepCycle: "none", HomeLocation: "underdark_lair"}
	beholder.MaxHP = 180
	beholder.HP = 180
	all = append(all, beholder)
	log.Printf("[gen] beholder spawned: Orbius at underdark_lair")

	floatingEyeAttrs := entity.Attributes{STR: 0, DEX: 10, CON: 0, INT: 14, WIS: 16, CHA: 12}
	for i := 0; i < 3; i++ {
		eye := entity.NewEntity("floating_eye_"+fmt.Sprint(i), "Watcher", "floating_eye", floatingEyeAttrs, 3, relation.BeastRelation)
		eye.LocationID = "underdark_lair"
		eye.Faction = "aberration"
		eye.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"scouting"}, FactionID: "aberration", SleepCycle: "none", HomeLocation: "underdark_lair"}
		all = append(all, eye)
		log.Printf("[gen] floating_eye spawned: Watcher at underdark_lair")
	}

	log.Printf("[gen] generateAberrations: done, total %d", len(all))
	return all
}

func (g *Generator) generateTrolls() []*entity.Entity {
	log.Printf("[gen] generateTrolls: spawning trolls in forests and swamps")
	var all []*entity.Entity
	trollAttrs := entity.Attributes{STR: 19, DEX: 11, CON: 18, INT: 6, WIS: 9, CHA: 6}
	regions := []string{}
	for _, loc := range g.World.AllLocations() {
		if loc.Type == world.LocRegion {
			biome := biomeFromRegionID(loc.ID)
			if biome == "forest" || biome == "swamp" {
				regions = append(regions, loc.ID)
			}
		}
	}
	if len(regions) == 0 {
		regions = []string{"aetheria"}
	}
	for i := 0; i < 2; i++ {
		reg := regions[i%len(regions)]
		troll := entity.NewEntity("troll_stump_"+fmt.Sprint(i), "Stump", "troll", trollAttrs, 6, relation.BeastRelation)
		troll.LocationID = reg
		troll.Faction = "beast"
		troll.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "beast", SleepCycle: "none", HomeLocation: reg}
		troll.MaxHP = 80
		troll.HP = 80
		all = append(all, troll)
		log.Printf("[gen] troll spawned: Stump at %s", reg)
	}
	log.Printf("[gen] generateTrolls: done, total %d", len(all))
	return all
}

func (g *Generator) generateOgres() []*entity.Entity {
	log.Printf("[gen] generateOgres: spawning ogres in highlands and wastelands")
	var all []*entity.Entity
	ogreAttrs := entity.Attributes{STR: 17, DEX: 9, CON: 15, INT: 7, WIS: 8, CHA: 7}
	for i := 0; i < 2; i++ {
		reg := "region_highlands_0"
		ogre := entity.NewEntity("ogre_gorrox_"+fmt.Sprint(i), "Gorrox", "ogre", ogreAttrs, 5, relation.BeastRelation)
		ogre.LocationID = reg
		ogre.Faction = "beast"
		ogre.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "beast", SleepCycle: "diurnal", HomeLocation: reg}
		ogre.MaxHP = 80
		ogre.HP = 80
		all = append(all, ogre)
		log.Printf("[gen] ogre spawned: Gorrox at %s", reg)
	}
	log.Printf("[gen] generateOgres: done, total %d", len(all))
	return all
}

func (g *Generator) generateEttins() []*entity.Entity {
	log.Printf("[gen] generateEttins: spawning ettins in highlands and mountains")
	var all []*entity.Entity
	ettinAttrs := entity.Attributes{STR: 20, DEX: 7, CON: 17, INT: 6, WIS: 10, CHA: 7}
	for i := 0; i < 2; i++ {
		reg := "region_highlands_0"
		ettin := entity.NewEntity("ettin_doublebrain_"+fmt.Sprint(i), "Doublebrain", "ettin", ettinAttrs, 7, relation.BeastRelation)
		ettin.LocationID = reg
		ettin.Faction = "beast"
		ettin.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "beast", SleepCycle: "diurnal", HomeLocation: reg}
		ettin.MaxHP = 100
		ettin.HP = 100
		all = append(all, ettin)
		log.Printf("[gen] ettin spawned: Doublebrain at %s", reg)
	}
	log.Printf("[gen] generateEttins: done, total %d", len(all))
	return all
}

func (g *Generator) generateCyclopses() []*entity.Entity {
	log.Printf("[gen] generateCyclopses: spawning cyclopses in mountains and wastes")
	var all []*entity.Entity
	cyclopsAttrs := entity.Attributes{STR: 22, DEX: 10, CON: 18, INT: 9, WIS: 12, CHA: 9}
	for i := 0; i < 2; i++ {
		reg := "region_waste_0"
		cyclops := entity.NewEntity("cyclops_barga_"+fmt.Sprint(i), "Barga", "cyclops", cyclopsAttrs, 8, relation.BeastRelation)
		cyclops.LocationID = reg
		cyclops.Faction = "beast"
		cyclops.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "beast", SleepCycle: "diurnal", HomeLocation: reg}
		cyclops.MaxHP = 120
		cyclops.HP = 120
		all = append(all, cyclops)
		log.Printf("[gen] cyclops spawned: Barga at %s", reg)
	}
	log.Printf("[gen] generateCyclopses: done, total %d", len(all))
	return all
}

func (g *Generator) generateMedusas() []*entity.Entity {
	log.Printf("[gen] generateMedusas: spawning medusas in caves and wastelands")
	var all []*entity.Entity
	medusaAttrs := entity.Attributes{STR: 18, DEX: 10, CON: 14, INT: 14, WIS: 16, CHA: 14}
	for i := 0; i < 2; i++ {
		reg := "region_waste_0"
		medusa := entity.NewEntity("medusa_gorgara_"+fmt.Sprint(i), "Gorgara", "medusa", medusaAttrs, 10, relation.BeastRelation)
		medusa.LocationID = reg
		medusa.Faction = "beast"
		medusa.AI = entity.EntityAI{Type: "scripted", ScriptIDs: []string{"aggressive"}, FactionID: "beast", SleepCycle: "diurnal", HomeLocation: reg}
		medusa.MaxHP = 100
		medusa.HP = 100
		all = append(all, medusa)
		log.Printf("[gen] medusa spawned: Gorgara at %s", reg)
	}
	log.Printf("[gen] generateMedusas: done, total %d", len(all))
	return all
}

// === HALF-SPECIES GENERATORS ===
// Half-species are placed in towns and settlements where different races interact.

func (g *Generator) generateHalfOrcs() []*entity.Entity {
	log.Printf("[gen] generateHalfOrcs: spawning half-orcs in settlements")
	var all []*entity.Entity
	halfOrcAttrs := entity.Attributes{STR: 14, DEX: 10, CON: 14, INT: 7, WIS: 8, CHA: 7}
	settlements := g.findMatureSettlements()
	if len(settlements) == 0 {
		settlements = []string{"aetheria"}
	}
	for i := 0; i < 3; i++ {
		loc := settlements[i%len(settlements)]
		ent := entity.NewEntity("half_orc_"+fmt.Sprint(i), generateHalfName("half_orc", g.RNG), "half_orc", halfOrcAttrs, 2+g.RNG.Intn(4), relation.HumanRelation)
		ent.LocationID = loc
		ent.Faction = "human"
		ent.AI = entity.EntityAI{Type: "passive", SleepCycle: "diurnal", HomeLocation: loc, ScriptIDs: []string{"human"}}
		equipItem(ent, lookup("leather_armor"))
		equipItem(ent, lookup("leather_boots"))
		if g.RNG.Intn(100) < 50 {
			equipItem(ent, lookup("iron_shield"))
		}
		equipItem(ent, lookup("orc_cleaver"))
		giveCurrency(ent, 5+g.RNG.Intn(15), g.RNG.Intn(5), 0)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] half-orc spawned: %s at %s", ent.Name, loc)
	}
	log.Printf("[gen] generateHalfOrcs: done, total %d", len(all))
	return all
}

func (g *Generator) generateHalfElves() []*entity.Entity {
	log.Printf("[gen] generateHalfElves: spawning half-elves in settlements")
	var all []*entity.Entity
	halfElfAttrs := entity.Attributes{STR: 10, DEX: 11, CON: 10, INT: 11, WIS: 11, CHA: 11}
	settlements := g.findMatureSettlements()
	if len(settlements) == 0 {
		settlements = []string{"aetheria"}
	}
	for i := 0; i < 3; i++ {
		loc := settlements[i%len(settlements)]
		ent := entity.NewEntity("half_elf_"+fmt.Sprint(i), generateHalfName("half_elf", g.RNG), "half_elf", halfElfAttrs, 2+g.RNG.Intn(4), relation.HumanRelation)
		ent.LocationID = loc
		ent.Faction = "human"
		ent.AI = entity.EntityAI{Type: "passive", SleepCycle: "diurnal", HomeLocation: loc, ScriptIDs: []string{"human"}}
		equipItem(ent, lookup("common_clothes"))
		equipItem(ent, lookup("leather_boots"))
		giveCurrency(ent, 5+g.RNG.Intn(15), g.RNG.Intn(5), 0)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] half-elf spawned: %s at %s", ent.Name, loc)
	}
	log.Printf("[gen] generateHalfElves: done, total %d", len(all))
	return all
}

func (g *Generator) generateHalfDwarves() []*entity.Entity {
	log.Printf("[gen] generateHalfDwarves: spawning half-dwarves in settlements")
	var all []*entity.Entity
	halfDwarfAttrs := entity.Attributes{STR: 12, DEX: 8, CON: 14, INT: 8, WIS: 9, CHA: 8}
	settlements := g.findMatureSettlements()
	if len(settlements) == 0 {
		settlements = []string{"aetheria"}
	}
	for i := 0; i < 3; i++ {
		loc := settlements[i%len(settlements)]
		ent := entity.NewEntity("half_dwarf_"+fmt.Sprint(i), generateHalfName("half_dwarf", g.RNG), "half_dwarf", halfDwarfAttrs, 2+g.RNG.Intn(4), relation.HumanRelation)
		ent.LocationID = loc
		ent.Faction = "human"
		ent.AI = entity.EntityAI{Type: "passive", SleepCycle: "diurnal", HomeLocation: loc, ScriptIDs: []string{"human"}}
		equipItem(ent, lookup("leather_armor"))
		equipItem(ent, lookup("leather_boots"))
		equipItem(ent, lookup("axe"))
		giveCurrency(ent, 5+g.RNG.Intn(15), g.RNG.Intn(5), 0)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] half-dwarf spawned: %s at %s", ent.Name, loc)
	}
	log.Printf("[gen] generateHalfDwarves: done, total %d", len(all))
	return all
}

func (g *Generator) generateHalfGoblins() []*entity.Entity {
	log.Printf("[gen] generateHalfGoblins: spawning half-goblins in settlements")
	var all []*entity.Entity
	halfGoblinAttrs := entity.Attributes{STR: 8, DEX: 13, CON: 9, INT: 5, WIS: 7, CHA: 7}
	settlements := g.findMatureSettlements()
	if len(settlements) == 0 {
		settlements = []string{"aetheria"}
	}
	for i := 0; i < 2; i++ {
		loc := settlements[i%len(settlements)]
		ent := entity.NewEntity("half_goblin_"+fmt.Sprint(i), generateHalfName("half_goblin", g.RNG), "half_goblin", halfGoblinAttrs, 2+g.RNG.Intn(4), relation.HumanRelation)
		ent.LocationID = loc
		ent.Faction = "human"
		ent.AI = entity.EntityAI{Type: "passive", SleepCycle: "nocturnal", HomeLocation: loc, ScriptIDs: []string{"human"}}
		equipItem(ent, lookup("common_clothes"))
		giveCurrency(ent, 3+g.RNG.Intn(10), g.RNG.Intn(3), 0)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] half-goblin spawned: %s at %s", ent.Name, loc)
	}
	log.Printf("[gen] generateHalfGoblins: done, total %d", len(all))
	return all
}

func (g *Generator) generateHalfHobgoblins() []*entity.Entity {
	log.Printf("[gen] generateHalfHobgoblins: spawning half-hobgoblins in settlements")
	var all []*entity.Entity
	halfHobgoblinAttrs := entity.Attributes{STR: 12, DEX: 10, CON: 12, INT: 7, WIS: 8, CHA: 7}
	settlements := g.findMatureSettlements()
	if len(settlements) == 0 {
		settlements = []string{"aetheria"}
	}
	for i := 0; i < 2; i++ {
		loc := settlements[i%len(settlements)]
		ent := entity.NewEntity("half_hobgoblin_"+fmt.Sprint(i), generateHalfName("half_hobgoblin", g.RNG), "half_hobgoblin", halfHobgoblinAttrs, 2+g.RNG.Intn(4), relation.HumanRelation)
		ent.LocationID = loc
		ent.Faction = "human"
		ent.AI = entity.EntityAI{Type: "passive", SleepCycle: "diurnal", HomeLocation: loc, ScriptIDs: []string{"human"}}
		equipItem(ent, lookup("leather_armor"))
		equipItem(ent, lookup("leather_boots"))
		equipItem(ent, lookup("short_sword"))
		giveCurrency(ent, 5+g.RNG.Intn(15), g.RNG.Intn(5), 0)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] half-hobgoblin spawned: %s at %s", ent.Name, loc)
	}
	log.Printf("[gen] generateHalfHobgoblins: done, total %d", len(all))
	return all
}

func (g *Generator) generateHalfGnolls() []*entity.Entity {
	log.Printf("[gen] generateHalfGnolls: spawning half-gnolls in settlements")
	var all []*entity.Entity
	halfGnollAttrs := entity.Attributes{STR: 12, DEX: 11, CON: 11, INT: 4, WIS: 7, CHA: 5}
	settlements := g.findMatureSettlements()
	if len(settlements) == 0 {
		settlements = []string{"aetheria"}
	}
	for i := 0; i < 2; i++ {
		loc := settlements[i%len(settlements)]
		ent := entity.NewEntity("half_gnoll_"+fmt.Sprint(i), generateHalfName("half_gnoll", g.RNG), "half_gnoll", halfGnollAttrs, 2+g.RNG.Intn(4), relation.HumanRelation)
		ent.LocationID = loc
		ent.Faction = "human"
		ent.AI = entity.EntityAI{Type: "passive", SleepCycle: "nocturnal", HomeLocation: loc, ScriptIDs: []string{"human"}}
		equipItem(ent, lookup("leather_armor"))
		equipItem(ent, lookup("leather_boots"))
		equipItem(ent, lookup("iron_spear"))
		giveCurrency(ent, g.RNG.Intn(8), 0, 0)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] half-gnoll spawned: %s at %s", ent.Name, loc)
	}
	log.Printf("[gen] generateHalfGnolls: done, total %d", len(all))
	return all
}

func (g *Generator) generateHalfKobolds() []*entity.Entity {
	log.Printf("[gen] generateHalfKobolds: spawning half-kobolds in settlements")
	var all []*entity.Entity
	halfKoboldAttrs := entity.Attributes{STR: 7, DEX: 14, CON: 8, INT: 6, WIS: 7, CHA: 6}
	settlements := g.findMatureSettlements()
	if len(settlements) == 0 {
		settlements = []string{"aetheria"}
	}
	for i := 0; i < 2; i++ {
		loc := settlements[i%len(settlements)]
		ent := entity.NewEntity("half_kobold_"+fmt.Sprint(i), generateHalfName("half_kobold", g.RNG), "half_kobold", halfKoboldAttrs, 2+g.RNG.Intn(4), relation.HumanRelation)
		ent.LocationID = loc
		ent.Faction = "human"
		ent.AI = entity.EntityAI{Type: "passive", SleepCycle: "nocturnal", HomeLocation: loc, ScriptIDs: []string{"human"}}
		equipItem(ent, lookup("common_clothes"))
		giveCurrency(ent, 3+g.RNG.Intn(10), g.RNG.Intn(3), 0)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] half-kobold spawned: %s at %s", ent.Name, loc)
	}
	log.Printf("[gen] generateHalfKobolds: done, total %d", len(all))
	return all
}

func (g *Generator) generateHalfFey() []*entity.Entity {
	log.Printf("[gen] generateHalfFey: spawning half-fey in settlements")
	var all []*entity.Entity
	halfFeyAttrs := entity.Attributes{STR: 8, DEX: 12, CON: 9, INT: 10, WIS: 12, CHA: 13}
	settlements := g.findMatureSettlements()
	if len(settlements) == 0 {
		settlements = []string{"aetheria"}
	}
	for i := 0; i < 2; i++ {
		loc := settlements[i%len(settlements)]
		ent := entity.NewEntity("half_fey_"+fmt.Sprint(i), generateHalfName("half_fey", g.RNG), "half_fey", halfFeyAttrs, 2+g.RNG.Intn(4), relation.HumanRelation)
		ent.LocationID = loc
		ent.Faction = "human"
		ent.AI = entity.EntityAI{Type: "passive", SleepCycle: "nocturnal", HomeLocation: loc, ScriptIDs: []string{"human"}}
		equipItem(ent, lookup("common_clothes"))
		giveCurrency(ent, 5+g.RNG.Intn(15), g.RNG.Intn(5), 0)
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		all = append(all, ent)
		log.Printf("[gen] half-fey spawned: %s at %s", ent.Name, loc)
	}
	log.Printf("[gen] generateHalfFey: done, total %d", len(all))
	return all
}

// findMatureSettlements returns settlement location IDs in the world,
// falling back to aetheria if none found.
func (g *Generator) findMatureSettlements() []string {
	var results []string
	for _, loc := range g.World.AllLocations() {
		if loc.Type == world.LocCity || loc.Type == world.LocBuilding || loc.Type == world.LocWildSite {
			results = append(results, loc.ID)
		}
	}
	return results
}

func generateHalfName(species string, rng *rand.Rand) string {
	switch species {
	case "half_orc":
		names := []string{"Grolf", "Krom", "Thok", "Brak", "Drog", "Krelka", "Draga", "Shaka"}
		return names[rng.Intn(len(names))]
	case "half_elf":
		names := []string{"Thalor", "Elric", "Caelum", "Fenrik", "Thalyra", "Elyra", "Caelia", "Riven"}
		return names[rng.Intn(len(names))]
	case "half_dwarf":
		names := []string{"Dorn", "Brik", "Haldor", "Grund", "Torin", "Dessa", "Brynn", "Torva"}
		return names[rng.Intn(len(names))]
	case "half_goblin":
		names := []string{"Skrit", "Yip", "Klik", "Drak", "Snik", "Rix", "Zik", "Vrik"}
		return names[rng.Intn(len(names))]
	case "half_hobgoblin":
		names := []string{"Karg", "Brak", "Zol", "Vorn", "Ghrak", "Krel", "Shara", "Velka"}
		return names[rng.Intn(len(names))]
	case "half_gnoll":
		names := []string{"Ripper", "Bonepick", "Snapper", "Vex", "Grak", "Mama", "Vexa", "Snapa"}
		return names[rng.Intn(len(names))]
	case "half_kobold":
		names := []string{"Skrit", "Yip", "Klik", "Drak", "Snik", "Skrix", "Yipa", "Klika"}
		return names[rng.Intn(len(names))]
	case "half_fey":
		names := []string{"Thorn", "Bram", "Alder", "Rowan", "Briar", "Thyra", "Sylph", "Faye"}
		return names[rng.Intn(len(names))]
	default:
		return species + "_spawn"
	}
}

func GenerateDefaultWorld() (*world.World, []*entity.Entity) {
	log.Printf("[gen] GenerateDefaultWorld: using seed \"default\"")
	g := NewGenerator("default")
	return g.Generate()
}
