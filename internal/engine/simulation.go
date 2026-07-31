package engine

import (
	"fmt"
	"log"
	"math/rand"
	"simuz/internal/ai"
	"simuz/internal/entity"
	"simuz/internal/events"
	"simuz/internal/items"
	"simuz/internal/quest"
	"simuz/internal/species"
	"simuz/internal/world"
	"sync"
	"time"
)

var Sim *Simulation

type Simulation struct {
	mu           sync.RWMutex
	Tick         uint64
	Scheduler    *Scheduler
	World        *world.World
	Entities     *entity.EntityManager
	Quests       *quest.Manager
	Events       *events.Manager
	Time         world.GameTime
	Storage      Storage
	RNG          *rand.Rand
	running      bool
	events       []events.SimEvent
	listeners    []func(events.SimEvent)
	SpawnManager *SpawnManager
	Traveling    map[string]*world.TravelState
}

func (s *Simulation) RLock()   { s.mu.RLock() }
func (s *Simulation) RUnlock() { s.mu.RUnlock() }
func (s *Simulation) Lock()    { s.mu.Lock() }
func (s *Simulation) Unlock()  { s.mu.Unlock() }
func (s *Simulation) GetAllEntities() []*entity.Entity {
	return s.Entities.All()
}
func NewSimulation(w *world.World, em *entity.EntityManager) *Simulation {
	qm := quest.NewManager()
	sim := &Simulation{
		Tick:         0,
		Scheduler:    NewScheduler(),
		World:        w,
		Entities:     em,
		Quests:       qm,
		Events:       events.NewManager(),
		Time:         world.NewGameTime(24),
		RNG:          rand.New(rand.NewSource(time.Now().UnixNano())),
		running:      false,
		events:       make([]events.SimEvent, 0),
		listeners:    make([]func(events.SimEvent), 0),
		SpawnManager: NewSpawnManager(),
		Traveling:    make(map[string]*world.TravelState),
	}
	qm.OnQuestComplete = func(entityID, questID string, rewards *quest.Rewards) {
		if rewards == nil {
			return
		}
		ent := GlobalEntityManagerGet(entityID)
		if ent == nil || !ent.Alive {
			return
		}

		species, ok := species.GetByID(ent.Species)
		if rewards.Experience > 0 && ok && species.CanLevelUp {
			sim.Emit(events.SimEvent{
				Type:   events.EventTypeQuestComplete,
				Source: ent.ID,
				Data: map[string]interface{}{
					"quest_id":   questID,
					"experience": rewards.Experience,
					"gold":       rewards.Gold,
				},
			})
			ent.AddXP(rewards.Experience)
			log.Printf("[quest] %s earned %d XP from quest '%s'", ent.Name, rewards.Experience, questID)
		}
		if rewards.Gold > 0 {
			for i := 0; i < rewards.Gold; i++ {
				ent.AddItem(items.NewItemInstance("gold_"+questID+fmt.Sprint(i), "gp", items.GetDef("gp"), 1))
			}
		}

	}

	GlobalEntityManager = sim.Entities
	ai.MoveRequest = sim.RequestMove
	ai.IsEntityTraveling = sim.IsTraveling
	return sim
}

func (s *Simulation) OnEvent(fn func(events.SimEvent)) {
	s.listeners = append(s.listeners, fn)
}

func (s *Simulation) Emit(event events.SimEvent) {
	s.events = append(s.events, event)
	for _, fn := range s.listeners {
		fn(event)
	}
}

func (s *Simulation) DrainEvents() []events.SimEvent {
	events := s.events
	s.events = nil
	return events
}

func (s *Simulation) EventsCopy() []events.SimEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]events.SimEvent, len(s.events))
	copy(out, s.events)
	return out
}

func (s *Simulation) Start() {
	s.running = true
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	log.Printf("Simulation started at tick 0")
	for range ticker.C {
		if !s.running {
			return
		}
		s.TickOnce()
	}
}

func (s *Simulation) Stop() {
	s.running = false
}

func (s *Simulation) TickOnce() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Tick++
	s.Quests.SetTick(s.Tick)

	s.Scheduler.ProcessDue(s.Tick)

	s.Time.Advance()

	for _, loc := range s.World.AllLocations() {
		if loc.IsOutside && loc.Weather != nil {
			if s.Tick%240 == 0 {
				loc.Weather = world.GenerateWeatherFor(s.Time.Season(), loc.ID, s.RNG)
			}
		}
	}

	processTravel(s)
	processTerritory(s)

	for _, ent := range s.Entities.All() {
		if ent.Alive {
			processAging(ent, s)
		}
	}

	for _, ent := range s.Entities.All() {
		if !ent.Alive || !ent.Conscious {
			continue
		}
		if s.IsTraveling(ent.ID) {
			continue
		}
		processEntityAI(ent, s)
	}

	// Quest definitions describe offers made by source NPCs. Once an entity
	// reaches the source's location, make the offer part of the simulation so
	// quest ownership and subsequent activity are observable in the UI.
	offerQuestsAtSources(s)

	for _, ent := range s.Entities.All() {
		if ent.Alive {
			ent.TickEffects()
		}
	}

	for _, ent := range s.Entities.All() {
		if ent.Alive {
			ent.TickMoods(s.Tick)
		}
	}

	for _, ent := range s.Entities.All() {
		if !ent.Alive || ent.MaxHP <= 0 {
			continue
		}
		// Skip passive regen if hostiles are nearby
		nearby := s.Entities.ByLocation(ent.LocationID)
		hasHostile := false
		for _, other := range nearby {
			if other.ID != ent.ID && other.Alive && ent.GetFactionRelation(other.Faction).String() == "hostile" {
				hasHostile = true
				break
			}
		}
		if hasHostile {
			continue
		}
		// Harsh outdoor weather slows regen and stresses civilians
		loc := s.World.Location(ent.LocationID)
		outdoorHarsh := false
		if loc != nil && loc.IsOutside {
			if wth := s.World.EffectiveWeather(ent.LocationID); wth != nil && wth.IsHarsh() {
				outdoorHarsh = true
				if s.Tick%60 == 0 && (ent.Faction == "civilian" || ent.Faction == "merchant") {
					ent.AddMoodModifier("weather", "stressed", 30)
				}
			}
		}
		switch ent.Activity.Type {
		case entity.ActivitySleep:
			if ent.HP < ent.MaxHP {
				ent.Heal(1)
			}
			if ent.FP < ent.MaxFP {
				ent.RestFP(1)
			}
		case entity.ActivityIdle, entity.ActivityMeditate:
			interval := uint64(3)
			if outdoorHarsh {
				interval = 6
			}
			if s.Tick%interval == 0 && ent.HP < ent.MaxHP {
				ent.Heal(1)
			}
			if s.Tick%2 == 0 && ent.FP < ent.MaxFP {
				ent.RestFP(1)
			}
		}
	}

	processTimeLimitQuests(s)

	// Natural reproduction: adult male/female pairs of the same species
	// at the same location have a small chance to produce offspring each tick.
	processReproduction(s)
	processCrossbreeding(s)

	if s.SpawnManager != nil {
		s.SpawnManager.ProcessSpawns(s.World, s.Entities, int(s.Tick), s.RNG)
	}

	if s.Events != nil {
		s.Events.ProcessTick(s.Tick, s.World, s.RNG, s.Entities.All())
	}

	s.Emit(events.SimEvent{
		Type: events.EventTick,
		Tick: s.Tick,
	})

	if s.Tick%saveInterval == 0 && s.Storage != nil && s.Storage.Enabled() {
		if err := s.Storage.Save(s); err != nil {
			log.Printf("Save error: %v", err)
		}
	}
}
