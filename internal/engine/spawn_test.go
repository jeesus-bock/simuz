package engine

import (
	"testing"
)

func TestSpawnRulesUpdated(t *testing.T) {
	sm := NewSpawnManager()
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
