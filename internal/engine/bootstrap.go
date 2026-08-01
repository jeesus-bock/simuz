package engine

import (
	"math/rand"
	"simuz/internal/ai"
	"simuz/internal/entity"
	"simuz/internal/gen"
	"simuz/internal/world"
)

// WorldBuilder handles the orchestration of raw templates into running state maps.
type WorldBuilder struct {
	seed string
	rng  *rand.Rand
}

func NewWorldBuilder(seed string, rng *rand.Rand) *WorldBuilder {
	return &WorldBuilder{
		seed: seed,
		rng:  rng,
	}
}

// BootstrapWorld allocates, links, and builds the entire world simulation.
func (wb *WorldBuilder) BootstrapWorld(w *world.World, em *entity.EntityManager) (*Simulation, error) {
	// 0. Load all Lua AI scripts from disk
	ai.InitScripts()

	// 1. Generate the world (locations, exits, weather) and mortal entities
	g := gen.NewGenerator(wb.seed)
	g.World = w
	genWorld, genEntities := g.Generate()

	// 2. Add generated entities to the entity manager
	for _, ent := range genEntities {
		em.Add(ent)
	}

	// 3. Create the simulation with the populated world and entity manager
	sim := NewSimulation(genWorld, em)

	// 4. Initialize Deities into the divided Divine Real Estate
	wb.initDeities(em)

	return sim, nil
}

func (wb *WorldBuilder) initDeities(em *entity.EntityManager) {
	// deityScriptIDs returns the script list for a deity.
	// deity_core runs for all deities; the per-deity script is included only if it exists.
	deityScriptIDs := func(id string) []string {
		known := map[string]bool{
			"seus_crackbolt":      true,
			"posse_eidon":         true,
			"othena_the_pedantic": true,
			"oriz_the_bloodshot":  true,
			"odd_in":              true,
			"thurn_the_thumper":   true,
			"low_key":             true,
			"froyda_the_thistle":  true,
			"ooh_huang":           true,
			"groan_yin":           true,
			"wukong_the_mangy":    true,
			"amater_ashes":        true,
			"snoozanoo":           true,
			"raijin_the_rattler":  true,
			"haydes_the_hoarder":  true,
			"tie_o_mat":           true,
			"baa_hamut":           true,
			"vaicna_the_unwashed": true,
		}
		if known[id] {
			return []string{"deity_core", id}
		}
		return []string{"deity_core"}
	}

	// Global package arrays like deityDefs are visible when compiling the entire package.
	// We read the definitions array we scrubbed previously.
	for _, def := range gen.DeityDefs {
		// Convert the deity definition into your exact matching Entity struct shape
		d := &entity.Entity{
			ID:         def.ID,
			Name:       def.Name,
			Species:    "divine",
			Profession: "deity",
			Gender:     "other",
			Alive:      true,
			Immortal:   true,
			Conscious:  true,
			LocationID: def.RealmRoomID,
			Attributes: def.Attributes,
			AI: entity.EntityAI{
				Type:       "scripted",
				ScriptIDs:  deityScriptIDs(def.ID),
				SleepCycle: "none",
			},

			// FIX 2: Instantiate as map[string]string to match your exact struct signature
			Memory: make(map[string]string),
			Skills: make(map[string]int),
			LanguageSkills: make(map[string]int),
		}

		// FIX 3: Store values as clean strings instead of booleans
		for _, need := range def.Needs {
			d.Memory["need_"+need] = "true"
		}
		d.Memory["divine_shape"] = def.Shape
		d.Memory["pantheon"] = def.Pantheon
		d.Memory["domain"] = def.Domain

		gen.AssignLanguages(d, wb.rng)

		// FIX 4: Dynamically append the deity directly to your entity manager's active slice tracking field.
		// If your manager stores entities in a field called Entities or List, append it directly:

		em.Add(d)

		// NOTE: If em.Entities throws an error, comment that line out and replace it with your
		// manager's actual addition method signature (e.g., em.Add(d) or em.Entities[d.ID] = d).
	}
}
