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

	// Always guarantee guard posts, a bard, and a diplomat
	for i := 0; i < 2+g.RNG.Intn(3); i++ {
		targetBuildings = append(targetBuildings, architecturalTask{
			id: townID, profession: "guard", targetRoom: townID,
		})
	}
	targetBuildings = append(targetBuildings, architecturalTask{
		id: townID, profession: "bard", targetRoom: townID,
	})
	targetBuildings = append(targetBuildings, architecturalTask{
		id: townID, profession: "diplomat", targetRoom: townID,
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

		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)

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

		case "bard":
			equipItem(ent, lookup("common_clothes"))
			addItem(ent, lookup("beer"))
			giveCurrency(ent, 5+g.RNG.Intn(10), 2+g.RNG.Intn(5), 0)

		case "diplomat":
			equipItem(ent, lookup("fine_clothes"))
			addItem(ent, lookup("dagger"))
			addItem(ent, lookup("herb_pouch"))
			giveCurrency(ent, 10+g.RNG.Intn(20), 5+g.RNG.Intn(10), 1+g.RNG.Intn(3))

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
	numCommoners := 10 + g.RNG.Intn(15)
	for i := 0; i < numCommoners; i++ {

		workerSpecies := dominantCulture
		var professions []string
		if workerSpecies == "orc" {
			professions = []string{
				"flesh_carver",      // Brutal alternative to herbalist/doctor (harvests specimens)
				"blood_trapper",     // Aggressive hunter tracking lethal territorial beasts
				"raider_vanguard",   // Direct spearhead courier that loots settlements on route
				"pit_brawler",       // Heavy replacement for bar_patron (constantly testing strength)
				"bone_scraper",      // Scavenger variant of thief (strips armor and weapons from corpses)
				"clan_drummer",      // Intimidating replacement for traveler/bard (sounds war cues)
				"skull_cleaver",     // Elite combat guard enforcing tribal laws
				"tribute_collector", // Aggressive merchant shaking down locals for resource taxes
			}
		} else {
			// Fallback array for civilized human/hobbit profiles
			professions = []string{"farmer", "fisherman", "herbalist", "courier", "bar_patron", "thief", "ranger", "traveler"}
		}
		prof := professions[g.RNG.Intn(len(professions))]
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
		npcName := specDef.GetRandomName(gender, g.RNG)
		npcID := fmt.Sprintf("%s_%s_%d", townID, prof, i)

		ent := entity.NewEntity(npcID, npcName, workerSpecies, entity.Attributes{STR: 10, DEX: 10}, 1, relation.CivilianRelation)
		ent.LocationID = townID
		ent.Gender = gender
		ent.BioProfile = &specDef
		ent.Age = specDef.AdultAge + g.RNG.Intn(10)
		ent.MaxAge = specDef.MaxAge
		ent.Alive = true
		ent.Conscious = true
		ent.AI = entity.EntityAI{
			Type:         "scripted",
			FactionID:    townID,
			HomeLocation: townID,
			SleepCycle:   specDef.DefaultSleepCycle,
			ScriptIDs:    []string{prof},
		}
		ent.Faction = "civilian"
		ent.Profession = prof
		AssignWorship(ent, g.RNG)
		AssignLanguages(ent, g.RNG)
		equipItem(ent, lookup("common_clothes"))
		giveCurrency(ent, 1+g.RNG.Intn(5), 0, 0)
		entities = append(entities, ent)
	}

	// Farm animals: sheep and dogs near settlements
	numSheep := 3 + g.RNG.Intn(5)
	for i := 0; i < numSheep; i++ {
		sheepID := fmt.Sprintf("%s_sheep_%d", townID, i)
		sheep := entity.NewEntity(sheepID, "Sheep", "sheep", entity.Attributes{STR: 4, DEX: 6, CON: 8}, 1, relation.Relation{})
		sheep.LocationID = townID
		sheep.Gender = "female"
		if g.RNG.Intn(100) < 30 {
			sheep.Gender = "male"
		}
		sheep.Faction = "civilian"
		sheep.AI = entity.EntityAI{Type: "passive", HomeLocation: townID, SleepCycle: "diurnal"}
		entities = append(entities, sheep)
	}

	numDogs := 1 + g.RNG.Intn(2)
	for i := 0; i < numDogs; i++ {
		dogID := fmt.Sprintf("%s_dog_%d", townID, i)
		dog := entity.NewEntity(dogID, "Dog", "dog", entity.Attributes{STR: 8, DEX: 12, CON: 10}, 1, relation.Relation{})
		dog.LocationID = townID
		dog.Gender = "male"
		if g.RNG.Intn(100) < 50 {
			dog.Gender = "female"
		}
		dog.Faction = "civilian"
		dog.AI = entity.EntityAI{Type: "scripted", HomeLocation: townID, SleepCycle: "diurnal", ScriptIDs: []string{"dog"}}
		entities = append(entities, dog)
	}

	return entities
}
