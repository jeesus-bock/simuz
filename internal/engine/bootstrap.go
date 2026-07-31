package engine

import (
	"math/rand"
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
	sim := NewSimulation(w)
	sim.Entities = em

	// 1. Initialize Deities into the divided Divine Real Estate
	wb.initDeities(em)

	// 2. Populate starting mortal demographics
	wb.spawnInitialPopulations(em)

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
			Gender:     "none",
			Alive:      true,
			Immortal:   true,
			Conscious:  true,
			LocationID: def.RealmRoomID,
			Attributes: def.Attributes,

			// FIX 2: Instantiate as map[string]string to match your exact struct signature
			Memory: make(map[string]string),
			Skills: make(map[string]int),
		}

		// FIX 3: Store values as clean strings instead of booleans
		for _, need := range def.Needs {
			d.Memory["need_"+need] = "true"
		}
		d.Memory["divine_shape"] = def.Shape

		// FIX 4: Dynamically append the deity directly to your entity manager's active slice tracking field.
		// If your manager stores entities in a field called Entities or List, append it directly:

		em.Add(d)

		// NOTE: If em.Entities throws an error, comment that line out and replace it with your
		// manager's actual addition method signature (e.g., em.Add(d) or em.Entities[d.ID] = d).
	}
}

func (wb *WorldBuilder) spawnInitialPopulations(em *entity.EntityManager) {
	// Use your separate generation loops to populate world cells directly into the manager
}
