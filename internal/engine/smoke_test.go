package engine

import (
	"math/rand"
	"testing"
	"time"

	"simuz/internal/ai"
	"simuz/internal/gen"
)

func TestSimulationStabilizeSmoke(t *testing.T) {
	ai.InitScripts()

	g := gen.New("smoke-stab-" + time.Now().Format("150405"))
	w, ents := g.Generate()

	sim := NewSimulation(w)
	for _, e := range ents {
		sim.Entities.Add(e)
	}
	deities, _ := gen.GenerateDeities(w)
	for _, d := range deities {
		sim.Entities.Add(d)
	}
	for _, q := range gen.SeedQuests() {
		sim.Quests.Register(q)
	}
	sim.RNG = rand.New(rand.NewSource(42))

	// advance a number of ticks
	for i := 0; i < 120; i++ {
		sim.TickOnce()
	}

	// dens exist
	dens := []string{"orc_camp", "wolf_den", "bandit_camp", "kobold_warren", "fey_glade"}
	for _, d := range dens {
		if loc := sim.World.Location(d); loc == nil {
			t.Errorf("missing den %s", d)
		}
	}

	// at least some spawns or entities at dens
	hasLife := false
	for _, d := range dens {
		if len(sim.Entities.ByLocation(d)) > 0 {
			hasLife = true
			break
		}
	}
	if !hasLife {
		t.Log("no entities at dens yet (may be ok early)")
	}

	// traveling map exists and doesn't panic
	_ = sim.IsTraveling("nonexistent")

	// territory fields on a region or site
	for _, loc := range sim.World.AllLocations() {
		if loc.ID == "northern_highlands" || loc.ID == "orc_camp" {
			_ = loc.ControllingFaction
			_ = loc.ControlStrength
		}
	}

	// weather present on regions
	reg := sim.World.Location("golden_plains")
	if reg == nil || reg.Weather == nil {
		t.Error("region missing weather")
	}

	t.Logf("smoke ok: tick=%d entities=%d locations=%d traveling=%d",
		sim.Tick, len(sim.Entities.All()), len(sim.World.AllLocations()), len(sim.Traveling))
}
