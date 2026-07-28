// Package events contains the simulation event engine and tick-based event processing helpers.
package events

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"simuz/internal/entity"
	"simuz/internal/world"
)

type Event struct {
	Tick    uint64
	Type    string
	Title   string
	Message string
	LocID   string
}

type Manager struct {
	mu     sync.RWMutex
	events []Event
	rng    *rand.Rand
}

func NewManager() *Manager {
	return &Manager{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *Manager) AddEvent(tick uint64, evType, title, msg, locID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, Event{
		Tick:    tick,
		Type:    evType,
		Title:   title,
		Message: msg,
		LocID:   locID,
	})
	if len(m.events) > 500 {
		m.events = m.events[len(m.events)-500:]
	}
}

func (m *Manager) RecentEvents(n int) []Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.events) == 0 {
		return nil
	}
	if n > len(m.events) {
		n = len(m.events)
	}
	out := make([]Event, n)
	copy(out, m.events[len(m.events)-n:])
	return out
}

func (m *Manager) ProcessTick(tick uint64, w *world.World, rng *rand.Rand, ents []*entity.Entity) {
	if tick%10 != 0 {
		return
	}

	for _, loc := range w.AllLocations() {
		if rng.Float64() > 0.3 {
			continue
		}
		var entsHere []*entity.Entity
		for _, e := range ents {
			if e.Alive && e.LocationID == loc.ID {
				entsHere = append(entsHere, e)
			}
		}
		if len(entsHere) == 0 {
			continue
		}

		maybeEvent := ambientEventFor(loc, rng, entsHere)
		if maybeEvent != nil {
			m.AddEvent(tick, "ambient", maybeEvent.Title, maybeEvent.Message, loc.ID)
		}
	}

	if tick%50 == 0 {
		for _, e := range ents {
			if !e.Alive {
				continue
			}
			switch e.Mood {
			case "angry":
				if rng.Float64() < 0.15 {
					loc := w.Location(e.LocationID)
					locName := e.LocationID
					if loc != nil {
						locName = loc.Name
					}
					m.AddEvent(tick, "mood", "Tension",
						fmt.Sprintf("%s looks ready to explode with rage in %s.", e.Name, locName),
						e.LocationID)
				}
			case "inspired":
				if rng.Float64() < 0.1 {
					m.AddEvent(tick, "mood", "Inspiration",
						fmt.Sprintf("%s feels a surge of creative energy.", e.Name),
						e.LocationID)
				}
			}
		}
	}

	if tick%100 == 0 && rng.Float64() < 0.4 {
		phase := "night"
		locName := "the realm"
		m.AddEvent(tick, "world", "World Event",
			fmt.Sprintf("A mysterious breeze sweeps across %s at %s.", locName, phase),
			"")
	}
}

type ambientTemplate struct {
	Title   string
	Message string
	Check   func(loc *world.Location, rng *rand.Rand, ents []*entity.Entity) bool
}

var ambientEvents = []ambientTemplate{
	{
		Title:   "Wind",
		Message: "The wind howls through the area, rattling loose branches.",
		Check:   func(loc *world.Location, rng *rand.Rand, ents []*entity.Entity) bool { return loc.IsOutside },
	},
	{
		Title:   "Footsteps",
		Message: "Distant footsteps echo from somewhere nearby.",
		Check:   func(loc *world.Location, rng *rand.Rand, ents []*entity.Entity) bool { return true },
	},
	{
		Title:   "Shadows",
		Message: "Shadows flicker at the edge of your vision.",
		Check:   func(loc *world.Location, rng *rand.Rand, ents []*entity.Entity) bool { return loc.IsOutside },
	},
	{
		Title:   "Animal Call",
		Message: "An animal calls out in the distance.",
		Check:   func(loc *world.Location, rng *rand.Rand, ents []*entity.Entity) bool { return loc.IsOutside },
	},
	{
		Title:   "Firelight",
		Message: "Firelight dances warmly, casting long shadows on the walls.",
		Check: func(loc *world.Location, rng *rand.Rand, ents []*entity.Entity) bool {
			return !loc.IsOutside && hasTag(loc, "inn")
		},
	},
	{
		Title:   "Murmurs",
		Message: "Quiet murmurs fill the room as people go about their business.",
		Check:   func(loc *world.Location, rng *rand.Rand, ents []*entity.Entity) bool { return !loc.IsOutside },
	},
}

func hasTag(loc *world.Location, tag string) bool {
	for _, t := range loc.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func ambientEventFor(loc *world.Location, rng *rand.Rand, ents []*entity.Entity) *Event {
	var candidates []ambientTemplate
	for _, t := range ambientEvents {
		if t.Check(loc, rng, ents) {
			candidates = append(candidates, t)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	t := candidates[rng.Intn(len(candidates))]
	return &Event{
		Type:    "ambient",
		Title:   t.Title,
		Message: t.Message,
		LocID:   loc.ID,
	}
}
