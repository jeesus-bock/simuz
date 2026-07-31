package gen

import (
	"fmt"
	"simuz/internal/entity"
	"simuz/internal/relation"
	"simuz/internal/species"
	"simuz/internal/world"
	"strings"
)

// generateNPCs procedurally generates a living population customized to the town's culture and structural layout nodes.
func (g *Generator) generateNPCs(townID string, dominantCulture string) []*entity.Entity {
	var entities []*entity.Entity

	type architecturalTask struct {
		id, profession, targetRoom string
	}
	var targetBuildings []architecturalTask

	// Fix: Read locations via a safe thread-locked method if available, or straight map loop
	for _, loc := range g.World.AllLocations() {
		if loc.ParentID == townID && loc.Type == world.LocBuilding {

			// FIX: Suffixes updated to perfectly match your generateTown building slice identifiers
			if strings.HasSuffix(loc.ID, "_inn") {
				// Safety: Ensure room fallback defaults to parent building if room generation fails
				targetBuildings = append(targetBuildings, architecturalTask{
					id: loc.ID, profession: "innkeeper", targetRoom: loc.ID,
				})
			} else if strings.HasSuffix(loc.ID, "_blacksmith") {
				targetBuildings = append(targetBuildings, architecturalTask{
					id: loc.ID, profession: "blacksmith", targetRoom: loc.ID,
				})
			} else if strings.HasSuffix(loc.ID, "_temple") {
				targetBuildings = append(targetBuildings, architecturalTask{
					id: loc.ID, profession: "wizard", targetRoom: loc.ID,
				})
			} else if strings.HasSuffix(loc.ID, "_market") {
				targetBuildings = append(targetBuildings, architecturalTask{
					id: loc.ID, profession: "merchant", targetRoom: loc.ID,
				})
			}
		}
	}

	// Always guarantee at least one municipal guard post exists mapped safely to town base
	targetBuildings = append(targetBuildings, architecturalTask{
		id: townID, profession: "guard", targetRoom: townID,
	})

	for _, b := range targetBuildings {
		workerSpecies := dominantCulture
		if g.RNG.Intn(100) < 15 {
			allPossibleSpecies := []string{"human", "orc", "elf", "dwarf", "hobbit"}
			workerSpecies = allPossibleSpecies[g.RNG.Intn(len(allPossibleSpecies))]
		}

		specDef, exists := species.Registry[workerSpecies]
		if !exists {
			specDef = species.Registry["human"]
		}

		gender := "male"
		if g.RNG.Intn(100) < 50 {
			gender = "female"
		}

		// Pulling procedural names dynamically from your defined species models
		npcName := specDef.GetRandomName(gender, g.RNG)
		npcID := fmt.Sprintf("%s_%s_%d", b.id, b.profession, g.RNG.Intn(1000))

		// FIX: Safe instantiation syntax
		ent := entity.NewEntity(npcID, npcName, workerSpecies, entity.Attributes{STR: 10, DEX: 10}, 1, relation.CivilianRelation)
		ent.LocationID = b.targetRoom
		ent.Gender = gender

		ent.BioProfile = &specDef
		ent.Age = specDef.AdultAge + g.RNG.Intn(10)
		ent.MaxAge = specDef.MaxAge
		ent.Alive = true
		ent.Conscious = true

		ent.AI = entity.EntityAI{
			Type:         "scripted",
			FactionID:    townID,
			HomeLocation: b.targetRoom,
			SleepCycle:   specDef.DefaultSleepCycle,
			ScriptIDs:    []string{b.profession},
		}
		ent.Faction = "civilian"
		ent.Profession = b.profession

		// Inventory Assignment Engine
		switch b.profession {
		case "innkeeper":
			equipItem(ent, lookup("common_clothes"))
			equipItem(ent, lookup("tankard"))
			addItem(ent, lookup("beer"))
			giveCurrency(ent, 15+g.RNG.Intn(30), 5+g.RNG.Intn(10), 0)

		case "blacksmith":
			equipItem(ent, lookup("work_tunic"))
			equipItem(ent, lookup("smith_hammer"))
			addItem(ent, lookup("iron_ore"))
			if workerSpecies == "orc" {
				addItem(ent, lookup("orc_cleaver"))
			}
			giveCurrency(ent, 5+g.RNG.Intn(15), 5+g.RNG.Intn(15), 1+g.RNG.Intn(5))

		case "guard":
			ent.AI.Brave = true
			equipItem(ent, lookup("leather_gloves"))
			giveCurrency(ent, 5, 3, 0)

			equipItem(ent, lookup("iron_sword"))

		case "wizard":
			equipItem(ent, lookup("priest_robe"))
			addItem(ent, lookup("charged_quartz"))
			giveCurrency(ent, 10, 2, 1)

		default:
			equipItem(ent, lookup("common_clothes"))
			giveCurrency(ent, 1+g.RNG.Intn(5), 0, 0)
		}
		// Contextual combat loadouts refactored into a clean switch statement
		switch workerSpecies {
		case "dwarf":
			ent.AI.Brave = true
			equipItem(ent, lookup("heavy_iron_plate"))
			equipItem(ent, lookup("iron_helmet"))
			equipItem(ent, lookup("iron_shield"))
			equipItem(ent, lookup("double_bit_axe"))
			addItem(ent, lookup("whetstone"))
			addItem(ent, lookup("dwarven_ale"))
			addItem(ent, lookup("iron_ore"))

		case "orc":
			ent.AI.Brave = true
			equipItem(ent, lookup("spiked_leather_harness"))
			equipItem(ent, lookup("orc_cleaver"))
			equipItem(ent, lookup("barbed_javelin"))
			addItem(ent, lookup("smoked_meat"))
			addItem(ent, lookup("crude_bandage"))

		case "elf":
			equipItem(ent, lookup("elven_chainmail"))
			equipItem(ent, lookup("recurve_longbow"))
			equipItem(ent, lookup("silver_shortsword"))
			addItem(ent, lookup("quiver_of_arrows"))
			addItem(ent, lookup("lembas_bread"))
			addItem(ent, lookup("glow_stone"))

		case "hobbit":
			equipItem(ent, lookup("padded_tunic"))
			equipItem(ent, lookup("hunting_slingshot"))
			equipItem(ent, lookup("sturdy_dagger"))
			addItem(ent, lookup("pouch_of_pebbles"))
			addItem(ent, lookup("pipeweed"))
			addItem(ent, lookup("pocket_watch"))

		default: // Default fallback handles regular humans
			equipItem(ent, lookup("chainmail"))
			equipItem(ent, lookup("iron_helmet"))
			equipItem(ent, lookup("iron_sword"))
			equipItem(ent, lookup("wooden_shield"))
			addItem(ent, lookup("iron_rations"))
			addItem(ent, lookup("torch"))
		}
		entities = append(entities, ent)
	}

	// Ambient population builder loops
	numCommoners := 1 + g.RNG.Intn(3)
	for i := 0; i < numCommoners; i++ {
		ent := entity.NewEntity(fmt.Sprintf("%s_peasant_%d", townID, i), "Town Folk", dominantCulture, entity.Attributes{STR: 10}, 1, relation.CivilianRelation)
		ent.LocationID = townID
		ent.Profession = "traveler"
		ent.AI = entity.EntityAI{Type: "passive", FactionID: townID, HomeLocation: townID, SleepCycle: "diurnal", ScriptIDs: []string{"traveler"}}
		equipItem(ent, lookup("common_clothes"))
		entities = append(entities, ent)
	}

	return entities
}
