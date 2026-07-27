package engine

import (
	"log"

	"simuz/internal/combat"
	"simuz/internal/entity"
	"simuz/internal/world"
)

const territoryInterval = 30

// processTerritory updates ControllingFaction on locations based on living presence.
func processTerritory(s *Simulation) {
	if s.Tick%territoryInterval != 0 {
		return
	}
	// Count force per location by faction (level-weighted)
	type force struct {
		faction string
		power   int
	}
	locPower := make(map[string]map[string]int)
	for _, e := range s.Entities.All() {
		if !e.Alive || e.Faction == "" || e.Faction == "deity" {
			continue
		}
		if locPower[e.LocationID] == nil {
			locPower[e.LocationID] = make(map[string]int)
		}
		pwr := 1 + e.Level
		if e.Faction == "civilian" || e.Faction == "merchant" {
			pwr = 1
		}
		locPower[e.LocationID][e.Faction] += pwr
	}

	// Roll presence up to parent region for wild control
	regionAgg := make(map[string]map[string]int)
	for locID, factions := range locPower {
		region := s.World.RegionOf(locID)
		if region == nil {
			continue
		}
		if regionAgg[region.ID] == nil {
			regionAgg[region.ID] = make(map[string]int)
		}
		for f, p := range factions {
			regionAgg[region.ID][f] += p
		}
	}

	for _, loc := range s.World.AllLocations() {
		powers := locPower[loc.ID]
		// Sites (camps/dens) use local presence; regions use aggregate
		if loc.Type == world.LocRegion {
			powers = regionAgg[loc.ID]
		}
		if len(powers) == 0 {
			if loc.ControlStrength > 0 {
				loc.ControlStrength--
				if loc.ControlStrength <= 0 {
					loc.ControllingFaction = ""
					loc.ControlStrength = 0
				}
			}
			continue
		}
		bestF, bestP := "", 0
		second := 0
		for f, p := range powers {
			if p > bestP {
				second = bestP
				bestP = p
				bestF = f
			} else if p > second {
				second = p
			}
		}
		// Civilian/merchant only hold towns (city/building under city), not wild
		if bestF == "civilian" || bestF == "merchant" {
			if loc.Type == world.LocRegion || loc.HasTag("camp") || loc.HasTag("den") || loc.HasTag("ruins") || loc.HasTag("hostile") {
				// pick strongest non-civilian
				bestF, bestP = "", 0
				for f, p := range powers {
					if f == "civilian" || f == "merchant" {
						continue
					}
					if p > bestP {
						bestP = p
						bestF = f
					}
				}
			}
		}
		if bestF == "" {
			continue
		}
		margin := bestP - second
		if loc.ControllingFaction == bestF {
			loc.ControlStrength += 1 + margin/3
			if loc.ControlStrength > 100 {
				loc.ControlStrength = 100
			}
		} else if loc.ControllingFaction == "" || margin >= 2 || bestP >= loc.ControlStrength {
			old := loc.ControllingFaction
			loc.ControllingFaction = bestF
			loc.ControlStrength = bestP
			if old != bestF {
				log.Printf("[territory] %s now controlled by %s (str %d)", loc.ID, bestF, loc.ControlStrength)
			}
		} else {
			loc.ControlStrength--
			if loc.ControlStrength <= 0 {
				loc.ControllingFaction = bestF
				loc.ControlStrength = bestP
			}
		}
	}
}

// NudgeTerritoryOnKill shifts control at the kill location toward the killer's faction.
func NudgeTerritoryOnKill(w *world.World, locID, killerFaction string) {
	if w == nil || killerFaction == "" || killerFaction == "civilian" || killerFaction == "deity" {
		return
	}
	loc := w.Location(locID)
	if loc == nil {
		return
	}
	if loc.ControllingFaction == killerFaction {
		loc.ControlStrength += 3
		if loc.ControlStrength > 100 {
			loc.ControlStrength = 100
		}
		return
	}
	loc.ControlStrength -= 5
	if loc.ControlStrength <= 0 {
		loc.ControllingFaction = killerFaction
		loc.ControlStrength = 5
	}
}

// ApplyTerritoryKillHooks after a kill.
func ApplyTerritoryKillHooks(s *Simulation, killer, victim *entity.Entity) {
	if killer == nil || victim == nil {
		return
	}
	NudgeTerritoryOnKill(s.World, victim.LocationID, killer.Faction)
	if killer.Faction != victim.Faction {
		combat.ShiftRelation(killer.Faction, victim.Faction, -1)
	}
}
