package engine

import (
	"log"

	"simuz/internal/entity"
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
	s.Traveling[ent.ID] = world.NewTravel(ent.ID, ent.LocationID, destID, ticks, world.TravelWalk)
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
	if s.Traveling == nil || len(s.Traveling) == 0 {
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
		if ts.Status == world.TravelArrived {
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
			s.Emit(SimEvent{
				Type:   EventTravelCompleted,
				Tick:   s.Tick,
				Source: ent.ID,
				Data:   map[string]any{"from": ts.FromID, "to": ts.ToID},
			})
			delete(s.Traveling, id)
		}
	}
}
