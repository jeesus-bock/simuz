package gen

import (
	"fmt"
	"math"
	"simuz/internal/entity"
	"simuz/internal/species"
	"simuz/internal/world"
)

// Structural template mapping out cultural settlement parameters
type settlementTemplate struct {
	namePrefixes []string
	nameSuffixes []string
	cityType     string // "keep", "shire", "canopy", "metropolis", "bastion"
}

// Global data matrix matching settlement themes to your species definitions
var cultureMatrix = map[string]settlementTemplate{
	"human": {
		namePrefixes: []string{"Kuningas", "Kivi", "Oru", "Põllu", "Ranna", "Uue"},
		nameSuffixes: []string{"linn", "kindlus", "alevik", "mõisa", "sadam"},
		cityType:     "metropolis",
	},
	"orc": {
		namePrefixes: []string{"Raud", "Tust", "Tahma", "Mürina", "Veri", "Krogu"},
		nameSuffixes: []string{"foundry", "garrison", "crag", "hold", "clanstead"},
		cityType:     "bastion",
	},
	"dwarf": {
		namePrefixes: []string{"Mäe", "Sepa", "Kaevur", "Raudne", "Graniit"},
		nameSuffixes: []string{"Keep", "Deep", "Vault", "Delve", "Forge"},
		cityType:     "keep",
	},
	"elf": {
		namePrefixes: []string{"Tähe", "Valguse", "Kuu", "Sireli", "Pilve"},
		nameSuffixes: []string{"Sanctuary", "Canopy", "Grove", "Spires", "Haven"},
		cityType:     "canopy",
	},
	"hobbit": {
		namePrefixes: []string{"Mätta", "Aasa", "Kodu", "Pätsi", "Uru"},
		nameSuffixes: []string{"Shire", "Burrows", "Hollow", "Meadow"},
		cityType:     "shire",
	},
}

// generateSettlements iterates over all active regions and builds dynamic permanent settlements
func (g *Generator) generateSettlements() []*entity.Entity {
	var totalWorldPop []*entity.Entity

	// 1. Gather all procedurally generated regions from your active world registry
	var regions []*world.Location
	for _, loc := range g.World.AllLocations() {
		if loc.Type == world.LocRegion {
			regions = append(regions, loc)
		}
	}

	for _, reg := range regions {
		// Determine the biome type archetype by parsing the region ID string prefix
		// (e.g., extracting "swamp" out of "region_swamp_1")
		var biomeKey string
		for key := range biomeMatrix {
			if len(reg.ID) >= 7+len(key) && reg.ID[7:7+len(key)] == key {
				biomeKey = key
				break
			}
		}
		if biomeKey == "" {
			biomeKey = "plains"
		}

		// 2. Proximity Matching: Scan our cultural profiles to find a species that loves this biome
		var eligibleCultures []string
		for speciesID, profile := range species.Registry {
			for _, preferred := range profile.PreferredBiomes {
				if preferred == biomeKey {
					eligibleCultures = append(eligibleCultures, speciesID)
				}
			}
		}

		// Fallback safe mapping defaults if an outlier region is encountered
		if len(eligibleCultures) == 0 {
			eligibleCultures = append(eligibleCultures, "human")
		}

		// Pick one random cultural archetype from the matching choices
		cultureKey := eligibleCultures[g.RNG.Intn(len(eligibleCultures))]
		culture := cultureMatrix[cultureKey]

		// 3. Procedural Language Mixing: Assemble a localized, authentic name descriptor
		pfx := culture.namePrefixes[g.RNG.Intn(len(culture.namePrefixes))]
		sfx := culture.nameSuffixes[g.RNG.Intn(len(culture.nameSuffixes))]
		settlementName := fmt.Sprintf("%s%s", pfx, sfx)
		settlementID := fmt.Sprintf("%s_city_%s", reg.ID, cultureKey)

		// 4. Spatial Offsetting: Place the city center close to the center coordinates of the region
		// We add a minor random offset to give a slight variance across layout zones
		sx := math.Round(reg.Position.X + (g.RNG.Float64()*10.0 - 5.0))
		sy := math.Round(reg.Position.Y + (g.RNG.Float64()*10.0 - 5.0))

		// 5. Instantiate using your standard city constructors
		cityLoc := world.NewLocation(settlementID, settlementName, world.LocCity, reg.ID, world.Position{X: sx, Y: sy})
		cityLoc.IsOutside = false
		cityLoc.Tags = []string{cultureKey, culture.cityType, "settlement"}
		g.World.AddLocation(cityLoc)

		// 6. Branch into your nested procedural building factory method we built earlier
		// We pass 'cultureKey' down so the building generator knows what kind of houses to build!
		g.generateThematicBuildings(settlementID, cultureKey)

		// 7. Queue population parameters to spawn members belonging to this culture later
		// settlementNPCs := g.generateNPCs(settlementID, cultureKey)
		// totalWorldPop = append(totalWorldPop, settlementNPCs...)
	}

	return totalWorldPop
}

// generateThematicBuildings maps unique structural setups depending on who populates the city
func (g *Generator) generateThematicBuildings(cityID string, culture string) {
	// Base foundational structures common to all urban nodes
	innID := cityID + "_inn"
	innName := "The Sleeping Dragon" // Can be randomized using our previous inn name arrays

	innLoc := world.NewLocation(innID, innName, world.LocBuilding, cityID, world.Position{X: 0, Y: 0})
	innLoc.Tags = []string{"inn"}
	g.World.AddLocation(innLoc)
	g.generateRooms(innID)

	// Cultural Architectural Branching
	switch culture {
	case "dwarf":
		// Dwarves get deep mechanical engineering setups
		forge := world.NewLocation(cityID+"_great_forge", "The Grand Foundry Core", world.LocBuilding, cityID, world.Position{X: -4, Y: 4})
		forge.Tags = []string{"forge", "blacksmith", "mechanical_workshop"}
		g.World.AddLocation(forge)

	case "orc":
		// Orcs get industrial weapon factories and armories
		scrapYard := world.NewLocation(cityID+"_smog_foundry", "Smog-Iron Smelter", world.LocBuilding, cityID, world.Position{X: 4, Y: -4})
		scrapYard.Tags = []string{"forge", "blacksmith", "cartel_scrap"}
		g.World.AddLocation(scrapYard)

	case "elf":
		// Elves get botanical laboratories and star towers
		observatory := world.NewLocation(cityID+"_star_spire", "Astrological Sky-Spire", world.LocBuilding, cityID, world.Position{X: 3, Y: 3})
		observatory.Tags = []string{"observatory", "temple", "astrology_node"}
		g.World.AddLocation(observatory)

	case "human":
		// Humans get bustling open-market bazaars
		market := world.NewLocation(cityID+"_bazaar", "Market Square Common", world.LocBuilding, cityID, world.Position{X: 5, Y: 5})
		market.Tags = []string{"market", "trade"}
		g.World.AddLocation(market)
	}
}
