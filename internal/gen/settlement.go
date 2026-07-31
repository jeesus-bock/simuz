package gen

import (
	"fmt"
	"log"
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
	log.Printf("[gen] generateSettlements: starting settlement generation")
	var totalWorldPop []*entity.Entity

	var regions []*world.Location
	for _, loc := range g.World.AllLocations() {
		if loc.Type == world.LocRegion {
			regions = append(regions, loc)
		}
	}

	log.Printf("[gen] generateSettlements: found %d regions, generating cities", len(regions))

	for _, reg := range regions {
		biomeKey := biomeFromRegionID(reg.ID)

		var eligibleCultures []string
		for speciesID, profile := range species.Registry {
			for _, preferred := range profile.PreferredBiomes {
				if preferred == biomeKey {
					eligibleCultures = append(eligibleCultures, speciesID)
				}
			}
		}

		if len(eligibleCultures) == 0 {
			eligibleCultures = append(eligibleCultures, "human")
		}

		cultureKey := eligibleCultures[g.RNG.Intn(len(eligibleCultures))]
		culture, ok := cultureMatrix[cultureKey]
		if !ok {
			culture = cultureMatrix["human"]
		}

		pfx := culture.namePrefixes[g.RNG.Intn(len(culture.namePrefixes))]
		sfx := culture.nameSuffixes[g.RNG.Intn(len(culture.nameSuffixes))]
		settlementName := fmt.Sprintf("%s%s", pfx, sfx)
		settlementID := fmt.Sprintf("%s_city_%s", reg.ID, cultureKey)

		sx := math.Round(reg.Position.X + (g.RNG.Float64()*10.0 - 5.0))
		sy := math.Round(reg.Position.Y + (g.RNG.Float64()*10.0 - 5.0))

		cityLoc := world.NewLocation(settlementID, settlementName, world.LocCity, reg.ID, world.Position{X: sx, Y: sy})
		cityLoc.IsOutside = false
		cityLoc.Tags = []string{cultureKey, culture.cityType, "settlement"}
		g.World.AddLocation(cityLoc)
		log.Printf("[gen] settlement: %q id=%s culture=%s biome=%s pos=(%.0f,%.0f)", settlementName, settlementID, cultureKey, biomeKey, sx, sy)

		g.generateThematicBuildings(settlementID, cultureKey)

		settlementNPCs := g.generateNPCs(settlementID, cultureKey)
		totalWorldPop = append(totalWorldPop, settlementNPCs...)
		log.Printf("[gen] settlement %q: spawned %d NPCs", settlementName, len(settlementNPCs))
	}

	log.Printf("[gen] generateSettlements: done, total %d settlement NPCs", len(totalWorldPop))
	return totalWorldPop
}

// generateThematicBuildings maps unique structural setups depending on who populates the city
func (g *Generator) generateThematicBuildings(cityID string, culture string) {
	log.Printf("[gen] generateThematicBuildings: city=%s culture=%s", cityID, culture)
	innID := cityID + "_inn"
	innName := "The Sleeping Dragon"

	innLoc := world.NewLocation(innID, innName, world.LocBuilding, cityID, world.Position{X: 0, Y: 0})
	innLoc.Tags = []string{"inn"}
	g.World.AddLocation(innLoc)
	log.Printf("[gen] building: %q (inn) at %s", innName, cityID)
	g.generateRooms(innID)

	switch culture {
	case "dwarf":
		forge := world.NewLocation(cityID+"_great_forge", "The Grand Foundry Core", world.LocBuilding, cityID, world.Position{X: -4, Y: 4})
		forge.Tags = []string{"forge", "blacksmith", "mechanical_workshop"}
		g.World.AddLocation(forge)
		log.Printf("[gen] building: The Grand Foundry Core (forge) at %s", cityID)

	case "orc":
		scrapYard := world.NewLocation(cityID+"_smog_foundry", "Smog-Iron Smelter", world.LocBuilding, cityID, world.Position{X: 4, Y: -4})
		scrapYard.Tags = []string{"forge", "blacksmith", "cartel_scrap"}
		g.World.AddLocation(scrapYard)
		log.Printf("[gen] building: Smog-Iron Smelter (forge) at %s", cityID)

	case "elf":
		observatory := world.NewLocation(cityID+"_star_spire", "Astrological Sky-Spire", world.LocBuilding, cityID, world.Position{X: 3, Y: 3})
		observatory.Tags = []string{"observatory", "temple", "astrology_node"}
		g.World.AddLocation(observatory)
		log.Printf("[gen] building: Astrological Sky-Spire (observatory) at %s", cityID)

	case "human":
		market := world.NewLocation(cityID+"_bazaar", "Market Square Common", world.LocBuilding, cityID, world.Position{X: 5, Y: 5})
		market.Tags = []string{"market", "trade"}
		g.World.AddLocation(market)
		log.Printf("[gen] building: Market Square Common (market) at %s", cityID)
	}
}
