// Package engine contains the simulation engine, tick processing, and related systems.
package engine

import (
	"log"

	"simuz/internal/entity"
	"simuz/internal/events"
	"simuz/internal/world"
)

func (s *Simulation) IsTraveling(entityID string) bool {
	if s.Traveling == nil {
		return false
	}
	ts, ok := s.Traveling[entityID]
	return ok && ts != nil && ts.Status == world.TravelInProgress
}

func (s *Simulation) TravelState(entityID string) *world.TravelState {
	if s.Traveling == nil {
		return nil
	}
	return s.Traveling[entityID]
}

// RequestMove moves instantly within region/city tree, or starts multi-tick travel across regions.
func (s *Simulation) RequestMove(ent *entity.Entity, destID string) bool {
	if ent == nil || !ent.Alive {
		return false
	}
	dest := s.World.Location(destID)
	if dest == nil {
		return false
	}
	if ent.LocationID == destID {
		return true
	}

	// Mortals cannot move TO divine realms, but can move FROM them (escape).
	// This allows mortals who somehow end up in divine realms to leave.
	toDivine := s.World.IsDivineRealm(destID)
	if toDivine && ent.Species != "deity" && ent.Faction != "deity" {
		return false
	}

	// If already in a divine realm, allow movement to any mortal location
	fromDivine := s.World.IsDivineRealm(ent.LocationID)
	if fromDivine && ent.Species != "deity" && ent.Faction != "deity" {
		// Allow escape to mortal locations
		if !toDivine {
			ent.LocationID = destID
			s.moveLeashedEntities(ent, destID)
			if s.Quests != nil {
				s.Quests.CheckVisitLocation(ent.ID, destID)
			}
			log.Printf("[travel] %s escaped from divine realm to %s", ent.Name, destID)
			return true
		}
		// Still block movement to other divine realms
		return false
	}

	route := s.World.Route(ent.LocationID, destID)
	if len(route) >= 2 {
		// Mortals cannot travel through divine realms at all.
		// Block the request if any intermediate node in the route is a divine realm.
		if ent.Species != "deity" && ent.Faction != "deity" {
			for _, locID := range route {
				if s.World.IsDivineRealm(locID) {
					log.Printf("[travel] %s route %s → %s blocked: passes through divine realm %s", ent.Name, ent.LocationID, destID, locID)
					return false
				}
			}
		}

		if len(route) == 2 && s.World.CanInstantMove(ent.LocationID, destID) {
			ent.LocationID = destID
			s.moveLeashedEntities(ent, destID)
			if s.Quests != nil {
				s.Quests.CheckVisitLocation(ent.ID, destID)
			}
			return true
		}
		if ts := s.TravelState(ent.ID); ts != nil && ts.Status == world.TravelInProgress {
			if ts.ToID == destID {
				return true
			}
			delete(s.Traveling, ent.ID)
		}
		if s.Traveling == nil {
			s.Traveling = make(map[string]*world.TravelState)
		}
		s.Traveling[ent.ID] = world.NewTravel(ent.ID, ent.LocationID, destID, route, len(route)-1, world.TravelWalk)
		ent.Activity = entity.EntityActivity{
			Type:      entity.ActivityTravel,
			SinceTick: s.Tick,
			UntilTick: s.Tick + uint64(len(route)-1),
		}
		log.Printf("[travel] %s plotting route %s → %s (%d steps)", ent.Name, ent.LocationID, destID, len(route)-1)
		return true
	}

	// Already traveling to this dest
	if ts := s.TravelState(ent.ID); ts != nil && ts.Status == world.TravelInProgress {
		if ts.ToID == destID {
			return true
		}
		// Reroute
		delete(s.Traveling, ent.ID)
	}

	if s.World.CanInstantMove(ent.LocationID, destID) {
		ent.LocationID = destID
		s.moveLeashedEntities(ent, destID)
		if s.Quests != nil {
			s.Quests.CheckVisitLocation(ent.ID, destID)
		}
		return true
	}

	fromReg := s.World.RegionOf(ent.LocationID)
	toReg := s.World.RegionOf(destID)
	if fromReg == nil || toReg == nil {
		// Fallback instant if no region context (realms)
		ent.LocationID = destID
		s.moveLeashedEntities(ent, destID)
		if s.Quests != nil {
			s.Quests.CheckVisitLocation(ent.ID, destID)
		}
		return true
	}
	if fromReg.ID == toReg.ID {
		ent.LocationID = destID
		s.moveLeashedEntities(ent, destID)
		if s.Quests != nil {
			s.Quests.CheckVisitLocation(ent.ID, destID)
		}
		return true
	}

	// Multi-tick cross-region travel
	weather := s.World.EffectiveWeather(ent.LocationID)
	ticks := world.TravelTimeWithWeather(fromReg, toReg, world.TravelWalk, weather)
	if s.Traveling == nil {
		s.Traveling = make(map[string]*world.TravelState)
	}
	s.Traveling[ent.ID] = world.NewTravel(ent.ID, ent.LocationID, destID, nil, ticks, world.TravelWalk)
	ent.Activity = entity.EntityActivity{
		Type:      entity.ActivityTravel,
		SinceTick: s.Tick,
		UntilTick: s.Tick + uint64(ticks),
	}
	log.Printf("[travel] %s departing %s → %s (%d ticks)", ent.Name, ent.LocationID, destID, ticks)
	return true
}

func (s *Simulation) moveLeashedEntities(dragger *entity.Entity, destID string) {
	if dragger == nil || s.Entities == nil {
		return
	}
	for _, e := range s.Entities.All() {
		if e.Alive && e.LeashedBy == dragger.ID {
			e.LocationID = destID
			log.Printf("[leash] %s dragged %s to %s", dragger.Name, e.Name, destID)
		}
	}
}

func processTravel(s *Simulation) {
	if len(s.Traveling) == 0 {
		return
	}
	for id, ts := range s.Traveling {
		if ts == nil || ts.Status != world.TravelInProgress {
			delete(s.Traveling, id)
			continue
		}
		ent := s.Entities.Get(id)
		if ent == nil || !ent.Alive {
			delete(s.Traveling, id)
			continue
		}
		ts.Tick()
		if len(ts.Route) > 1 && ts.RouteIndex < len(ts.Route)-1 {
			nextIdx := ts.RouteIndex + 1
			nextID := ts.Route[nextIdx]

			// Mortals cannot enter divine realms mid-travel.
			// If the next step is a divine realm, abort the travel and let them stay put.
			if ent.Species != "divine" && ent.Species != "deity" && ent.Faction != "deity" && s.World.IsDivineRealm(nextID) {
				ts.Status = world.TravelArrived
				ent.Activity = entity.EntityActivity{
					Type:      entity.ActivityIdle,
					SinceTick: s.Tick,
				}
				log.Printf("[travel] %s blocked from entering divine realm %s, travel aborted", ent.Name, nextID)
				s.Emit(events.SimEvent{
					Type:   events.EventTravelCompleted,
					Tick:   s.Tick,
					Source: ent.ID,
					Data:   map[string]any{"from": ts.FromID, "to": ts.ToID, "blocked": true},
				})
				delete(s.Traveling, id)
				continue
			}

			ent.LocationID = nextID
			s.moveLeashedEntities(ent, nextID)
			ts.RouteIndex = nextIdx
			if ts.RouteIndex >= len(ts.Route)-1 {
				ts.Status = world.TravelArrived
			}
			log.Printf("[travel] %s moved along route to %s (%d/%d)", ent.Name, nextID, ts.RouteIndex, len(ts.Route)-1)
		}
		if ts.Status == world.TravelArrived {
			ent.Activity = entity.EntityActivity{
				Type:      entity.ActivityIdle,
				SinceTick: s.Tick,
			}
			if s.Quests != nil {
				s.Quests.CheckVisitLocation(ent.ID, ts.ToID)
			}
			log.Printf("[travel] %s arrived at %s", ent.Name, ts.ToID)
			s.Emit(events.SimEvent{
				Type:   events.EventTravelCompleted,
				Tick:   s.Tick,
				Source: ent.ID,
				Data:   map[string]any{"from": ts.FromID, "to": ts.ToID, "route": ts.Route},
			})
			delete(s.Traveling, id)
		} else if len(ts.Route) == 0 && ts.Status == world.TravelArrived {
			ent.LocationID = ts.ToID
			s.moveLeashedEntities(ent, ts.ToID)
			ent.Activity = entity.EntityActivity{
				Type:      entity.ActivityIdle,
				SinceTick: s.Tick,
			}
			if s.Quests != nil {
				s.Quests.CheckVisitLocation(ent.ID, ts.ToID)
			}
			log.Printf("[travel] %s arrived at %s", ent.Name, ts.ToID)
			s.Emit(events.SimEvent{
				Type:   events.EventTravelCompleted,
				Tick:   s.Tick,
				Source: ent.ID,
				Data:   map[string]any{"from": ts.FromID, "to": ts.ToID},
			})
			delete(s.Traveling, id)
		}
	}
}
