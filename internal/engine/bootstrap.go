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
				ScriptIDs:  []string{"deity_core", def.ID},
				SleepCycle: "none",
			},

			// FIX 2: Instantiate as map[string]string to match your exact struct signature
			Memory: make(map[string]string),
			Skills: make(map[string]int),
		}

		// FIX 3: Store values as clean strings instead of booleans
		for _, need := range def.Needs {
			d.Memory["need_"+need] = "true"
		}
		d.Memory["divine_shape"] = def.Shape
		d.Memory["pantheon"] = def.Pantheon
		d.Memory["domain"] = def.Domain

		// FIX 4: Dynamically append the deity directly to your entity manager's active slice tracking field.
		// If your manager stores entities in a field called Entities or List, append it directly:

		em.Add(d)

		// NOTE: If em.Entities throws an error, comment that line out and replace it with your
		// manager's actual addition method signature (e.g., em.Add(d) or em.Entities[d.ID] = d).
	}
}
