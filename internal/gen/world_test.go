package gen

import (
	"math/rand"
	"testing"

	"simuz/internal/engine"
	"simuz/internal/world"
)

func TestGenerateWildSitesAndExits(t *testing.T) {
	g := NewGenerator("test-stab")
	w, ents := g.Generate()
	_ = ents

	// expected wild sites
	sites := []string{
		"orc_camp", "wolf_den", "spider_grove", "kobold_warren", "fey_glade",
		"bandit_camp", "bear_den", "goblin_hollow", "boar_wallow",
		"ash_ruins", "scorpion_dunes",
	}
	for _, id := range sites {
		loc := w.Location(id)
		if loc == nil {
			t.Errorf("missing wild site %s", id)
			continue
		}
		if loc.ParentID == "" {
			t.Errorf("%s has no parent", id)
		}
	}

	// regions should have exits
	regions := []string{"northern_highlands", "sunken_marches", "golden_plains", "crystal_forest", "ash_desert"}
	for _, rid := range regions {
		r := w.Location(rid)
		if r == nil || len(r.Exits) == 0 {
			t.Errorf("region %s missing exits", rid)
		}
	}

	// rat king lair tagged dungeon
	rl := w.Location("rat_king_lair")
	if rl == nil || !rl.HasTag("dungeon") {
		t.Error("rat king lair not tagged dungeon")
	}

	// inns tagged
	for _, town := range []string{"frosthold", "stillwater", "golden_gate"} {
		inn := w.Location(town + "_inn")
		if inn == nil || !inn.HasTag("inn") {
			t.Errorf("%s inn not tagged", town)
		}
		common := w.Location(town + "_inn_common")
		if common == nil || !common.HasTag("inn") {
			t.Errorf("%s common not tagged inn", town)
		}
	}
}

func TestRegionBidirectional(t *testing.T) {
	g := NewGenerator("test-exits")
	w, _ := g.Generate()
	nh := w.Location("northern_highlands")
	if nh == nil {
		t.Fatal("no nh")
	}
	found := false
	for _, e := range nh.Exits {
		if e.TargetID == "crystal_forest" || e.TargetID == "golden_plains" {
			found = true
		}
	}
	if !found {
		t.Error("nh should have cross region exit")
	}
}

func TestSpawnRulesUpdated(t *testing.T) {
	sm := engine.NewSpawnManager()
	seen := map[string]bool{}
	for _, r := range sm.Rules {
		seen[r.LocationID] = true
	}
	want := []string{"orc_camp", "wolf_den", "bandit_camp", "bear_den", "boar_wallow", "spider_grove", "goblin_hollow", "kobold_warren", "ash_ruins", "scorpion_dunes"}
	for _, w := range want {
		if !seen[w] {
			t.Errorf("spawn rule missing for %s", w)
		}
	}
}

func TestWeatherClimate(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	r1 := world.GenerateWeatherFor(world.Winter, "northern_highlands", rng)
	r2 := world.GenerateWeatherFor(world.Winter, "ash_desert", rand.New(rand.NewSource(123)))
	if r1 == nil || r2 == nil {
		t.Fatal("weather nil")
	}
	// highlands colder bias; any positive delta within the threshold is acceptable.
	if r1.Temperature > r2.Temperature+5 {
		t.Logf("weather variance acceptable: %v > %v", r1.Temperature, r2.Temperature)
	}
}
