package engine

import (
	"math/rand"
	"simuz/internal/entity"
	"simuz/internal/events"
	"simuz/internal/quest"
	"simuz/internal/world"
	"sync"
)

var Sim *Simulation

type Simulation struct {
	mu           sync.RWMutex
	Tick         uint64
	Scheduler    *Scheduler
	World        *world.World
	Entities     *entity.Manager
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
