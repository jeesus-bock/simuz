package engine

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"simuz/internal/ai"
	"simuz/internal/combat"
	"simuz/internal/entity"
	"simuz/internal/events"
	"simuz/internal/items"
	"simuz/internal/quest"
	"simuz/internal/world"
)

const saveInterval = 300

type Storage interface {
	Save(sim *Simulation) error
	Load() (*Simulation, error)
	Enabled() bool
}

type EventType int

const (
	EventEntityKilled EventType = iota
	EventEntityTalked
	EventLocationEntered
	EventItemCollected
	EventItemDelivered
	EventItemUsed
	EventCraftCompleted
	EventTravelCompleted
	EventTick
	EventTimePassed
)

type SimEvent struct {
	Type   EventType
	Tick   uint64
	Source string
	Data   map[string]any
}

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
	events       []SimEvent
	listeners    []func(SimEvent)
	SpawnManager *SpawnManager
	Traveling    map[string]*world.TravelState
}

func (s *Simulation) RLock()    { s.mu.RLock() }
func (s *Simulation) RUnlock()  { s.mu.RUnlock() }
func (s *Simulation) Lock()     { s.mu.Lock() }
func (s *Simulation) Unlock()   { s.mu.Unlock() }

func NewSimulation(w *world.World) *Simulation {
	qm := quest.NewManager()
	qm.OnQuestComplete = func(entityID, questID string, rewards *quest.Rewards) {
		if rewards == nil {
			return
		}
		ent := globalEntityManagerGet(entityID)
		if ent == nil || !ent.Alive {
			return
		}
		if rewards.Experience > 0 && entity.CanLevelUp(ent.Species) {
			ent.AddXP(rewards.Experience)
			log.Printf("[quest] %s earned %d XP from quest '%s'", ent.Name, rewards.Experience, questID)
		}
		if rewards.Gold > 0 {
			for i := 0; i < rewards.Gold; i++ {
				ent.AddItem(items.NewItemInstance("gold_"+questID+fmt.Sprint(i), "gp", items.GetDef("gp"), 1))
			}
		}
	}
	sim := &Simulation{
		Tick:         0,
		Scheduler:    NewScheduler(),
		World:        w,
		Entities:     entity.NewManager(),
		Quests:       qm,
		Events:       events.NewManager(),
		Time:         world.NewGameTime(24),
		RNG:          rand.New(rand.NewSource(time.Now().UnixNano())),
		running:      false,
		events:       make([]SimEvent, 0),
		listeners:    make([]func(SimEvent), 0),
		SpawnManager: NewSpawnManager(),
		Traveling:    make(map[string]*world.TravelState),
	}
	globalEntityManager = sim.Entities
	ai.MoveRequest = sim.RequestMove
	ai.IsEntityTraveling = sim.IsTraveling
	return sim
}

var globalEntityManager *entity.Manager

func globalEntityManagerGet(id string) *entity.Entity {
	if globalEntityManager == nil {
		return nil
	}
	return globalEntityManager.Get(id)
}

func (s *Simulation) OnEvent(fn func(SimEvent)) {
	s.listeners = append(s.listeners, fn)
}

func (s *Simulation) Emit(event SimEvent) {
	s.events = append(s.events, event)
	for _, fn := range s.listeners {
		fn(event)
	}
}

func (s *Simulation) DrainEvents() []SimEvent {
	events := s.events
	s.events = nil
	return events
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
		if !ent.Alive {
			continue
		}
		if s.IsTraveling(ent.ID) {
			continue
		}
		processEntityAI(ent, s)
	}

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
			if other.ID != ent.ID && other.Alive && combat.Relation(ent.Faction, other.Faction) == combat.Hostile {
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

	if s.SpawnManager != nil {
		s.SpawnManager.ProcessSpawns(s.World, s.Entities, int(s.Tick), s.RNG)
	}

	if s.Events != nil {
		s.Events.ProcessTick(s.Tick, s.World, s.RNG, s.Entities.All())
	}

	s.Emit(SimEvent{
		Type: EventTick,
		Tick: s.Tick,
	})

	if s.Tick%saveInterval == 0 && s.Storage != nil && s.Storage.Enabled() {
		if err := s.Storage.Save(s); err != nil {
			log.Printf("Save error: %v", err)
		}
	}
}

func processEntityAI(ent *entity.Entity, sim *Simulation) {
	phase := sim.Time.Phase()
	combat.SetTick(sim.Tick)

	setActivity(ent, sim.Tick, phase)

	if ent.AI.Type == "scripted" || ent.AI.Type == "dormant" {
		if ent.AI.Type == "scripted" {
			for _, sid := range ent.AI.ScriptIDs {
				if sid == "" {
					continue
				}
				if err := ai.RunScript(sid, ent, sim.World, sim.Entities, &sim.Time, sim.RNG, sim.Quests); err != nil {
					log.Printf("AI script error for %s (%s): %v", ent.Name, sid, err)
				}
			}
		}
		return
	}

	home := ent.AI.HomeLocation
	isHome := home != "" && ent.LocationID == home

	if isSleepTime(ent, phase) {
		trySleep(ent, sim, home, phase)
		return
	}

	if sim.Tick%20 == 0 && home == "" && sim.RNG.Intn(100) < 10 {
		roamPassive(ent, sim)
	}

	if sim.Tick%30 == 0 && home != "" {
		if isHome && sim.RNG.Intn(100) < 30 {
			roamPassive(ent, sim)
		} else if !isHome && sim.RNG.Intn(100) < 40 {
			ent.LocationID = home
			log.Printf("[ai] %s returned home to %s", ent.Name, home)
			setActivity(ent, sim.Tick, phase)
		}
	}

	switch ent.AI.Type {
	case "aggressive":
		factionHash := 0
		for _, c := range ent.Faction {
			factionHash += int(c)
		}
		rate := 5 + (factionHash % 5)
		if sim.Tick%uint64(rate) != 0 {
			return
		}
		nearby := sim.Entities.ByLocation(ent.LocationID)
		var target *entity.Entity
		for _, other := range nearby {
			if other.ID == ent.ID || !other.Alive || other.Immortal {
				continue
			}
			if combat.Relation(ent.Faction, other.Faction) == combat.Hostile {
				target = other
				break
			}
		}
		if target != nil {
			hit := simpleAttackAt(sim, ent, target)
			applyCombatMoods(ent, target, hit, sim.Tick)
			if !target.Alive {
				combat.LootCorpse(ent, target)
				sim.Emit(SimEvent{
					Type:   EventEntityKilled,
					Tick:   sim.Tick,
					Source: ent.ID,
					Data:   map[string]any{"target": target.ID, "attacker": ent.ID},
				})
				rewardXP(ent, target)
				questKilled(sim, target)
			}
			return
		}

		if sim.Tick%30 == 0 && sim.RNG.Intn(100) < 40 {
			wanderAggressive(ent, sim)
		}

	case "defensive":
		if sim.Tick%3 != 0 {
			return
		}
		nearby := sim.Entities.ByLocation(ent.LocationID)
		var target *entity.Entity
		for _, other := range nearby {
			if other.ID == ent.ID || !other.Alive || other.Immortal {
				continue
			}
			if combat.Relation(ent.Faction, other.Faction) == combat.Hostile {
				target = other
				break
			}
		}
		if target != nil {
			hit := simpleAttackAt(sim, ent, target)
			applyCombatMoods(ent, target, hit, sim.Tick)
			if !target.Alive {
				combat.LootCorpse(ent, target)
				sim.Emit(SimEvent{
					Type:   EventEntityKilled,
					Tick:   sim.Tick,
					Source: ent.ID,
					Data:   map[string]any{"target": target.ID, "attacker": ent.ID},
				})
				rewardXP(ent, target)
				questKilled(sim, target)
			}
			return
		}
		if home != "" && !isHome && sim.RNG.Intn(100) < 40 {
			ent.LocationID = home
		}

	case "hunting":
		if sim.Tick%3 != 0 {
			return
		}
		nearby := sim.Entities.ByLocation(ent.LocationID)
		var target *entity.Entity
		for _, other := range nearby {
			if other.ID == ent.ID || !other.Alive || other.Immortal {
				continue
			}
			if combat.Relation(ent.Faction, other.Faction) == combat.Hostile {
				target = other
				break
			}
		}
		if target != nil {
			hit := simpleAttackAt(sim, ent, target)
			applyCombatMoods(ent, target, hit, sim.Tick)
			if !target.Alive {
				combat.LootCorpse(ent, target)
				sim.Emit(SimEvent{
					Type:   EventEntityKilled,
					Tick:   sim.Tick,
					Source: ent.ID,
					Data:   map[string]any{"target": target.ID, "attacker": ent.ID},
				})
				rewardXP(ent, target)
				questKilled(sim, target)
			}
			return
		}
		// Chase prey toward hostile entities in nearby locations
		if sim.Tick%30 == 0 {
			nearbyLocs := sim.World.ChildLocations(ent.LocationID)
			for _, loc := range nearbyLocs {
				locEntities := sim.Entities.ByLocation(loc.ID)
				for _, e := range locEntities {
					if e.Alive && !e.Immortal && combat.Relation(ent.Faction, e.Faction) == combat.Hostile {
						ent.LocationID = loc.ID
						return
					}
				}
			}
			wanderAggressive(ent, sim)
		}

	case "gathering":
		if sim.Tick%10 != 0 {
			return
		}
		// Avoid combat: flee from hostiles
		nearby := sim.Entities.ByLocation(ent.LocationID)
		for _, other := range nearby {
			if other.ID == ent.ID || !other.Alive || other.Immortal {
				continue
			}
			if combat.Relation(ent.Faction, other.Faction) == combat.Hostile {
				if home != "" && ent.LocationID != home {
					ent.LocationID = home
				} else {
					siblings := sim.World.ChildLocations(ent.LocationID)
					if len(siblings) > 0 {
						safe := siblings[sim.RNG.Intn(len(siblings))]
						ent.LocationID = safe.ID
					}
				}
				return
			}
		}
		// Gather: stay in resource-rich outdoor locations, occasionally pick up items
		parent := sim.World.Location(ent.LocationID)
		if parent != nil && parent.IsOutside {
			if sim.RNG.Intn(100) < 15 {
				log.Printf("[ai] %s gathered resources at %s", ent.Name, ent.LocationID)
			}
		} else if home != "" && ent.LocationID != home {
			ent.LocationID = home
		}

	case "healing":
		if sim.Tick%5 != 0 {
			return
		}
		// Heal nearby injured non-hostile entities
		nearby := sim.Entities.ByLocation(ent.LocationID)
		for _, other := range nearby {
			if other.ID == ent.ID || !other.Alive || other.Immortal {
				continue
			}
			if other.HP < other.MaxHP && combat.Relation(ent.Faction, other.Faction) != combat.Hostile {
				healAmt := 2 + ent.Level
				other.Heal(healAmt)
				log.Printf("[ai] %s healed %s for %d HP", ent.Name, other.Name, healAmt)
			}
		}
		// Stay near the home location (temple/sanctuary)
		if home != "" && ent.LocationID != home && sim.RNG.Intn(100) < 20 {
			ent.LocationID = home
		}

	case "scouting":
		if sim.Tick%2 != 0 {
			return
		}
		// Scout: wander nearby locations, flee from hostiles
		nearby := sim.Entities.ByLocation(ent.LocationID)
		hasHostile := false
		for _, other := range nearby {
			if other.ID == ent.ID || !other.Alive || other.Immortal {
				continue
			}
			if combat.Relation(ent.Faction, other.Faction) == combat.Hostile {
				hasHostile = true
				break
			}
		}
		if hasHostile {
			// Flee home
			if home != "" && ent.LocationID != home {
				ent.LocationID = home
				log.Printf("[ai] %s fled to %s", ent.Name, home)
			} else {
				// Retreat to parent location
				parent := sim.World.Location(ent.LocationID)
				if parent != nil && parent.ParentID != "" {
					ent.LocationID = parent.ParentID
					log.Printf("[ai] %s retreated to %s", ent.Name, parent.ParentID)
				}
			}
			return
		}
		// Explore child locations
		children := sim.World.ChildLocations(ent.LocationID)
		if len(children) > 0 {
			dest := children[sim.RNG.Intn(len(children))]
			ent.LocationID = dest.ID
		} else if sim.RNG.Intn(100) < 30 {
			wanderAggressive(ent, sim)
		}

	case "guarding":
		if sim.Tick%2 != 0 {
			return
		}
		// Stay at home location, attack any hostile that approaches
		if home != "" && ent.LocationID != home {
			ent.LocationID = home
			return
		}
		nearby := sim.Entities.ByLocation(ent.LocationID)
		var target *entity.Entity
		for _, other := range nearby {
			if other.ID == ent.ID || !other.Alive || other.Immortal {
				continue
			}
			if combat.Relation(ent.Faction, other.Faction) == combat.Hostile {
				target = other
				break
			}
		}
		if target != nil {
			hit := simpleAttackAt(sim, ent, target)
			applyCombatMoods(ent, target, hit, sim.Tick)
			if !target.Alive {
				combat.LootCorpse(ent, target)
				sim.Emit(SimEvent{
					Type:   EventEntityKilled,
					Tick:   sim.Tick,
					Source: ent.ID,
					Data:   map[string]any{"target": target.ID, "attacker": ent.ID},
				})
				rewardXP(ent, target)
				questKilled(sim, target)
			}
		}
	}
}

func isSleepTime(ent *entity.Entity, phase world.DayPhase) bool {
	cycle := ent.AI.SleepCycle
	if cycle == "none" {
		return false
	}
	if cycle == "nocturnal" {
		return phase != world.Night && phase != world.Dusk
	}
	return phase == world.Night || phase == world.Dusk
}

func trySleep(ent *entity.Entity, sim *Simulation, home string, phase world.DayPhase) {
	cycle := ent.AI.SleepCycle
	if cycle == "none" {
		return
	}
	isNocturnal := cycle == "nocturnal"
	if phase == world.Dusk || phase == world.Dawn {
		if home != "" && ent.LocationID != home {
			ent.LocationID = home
			log.Printf("[ai] %s went home to sleep", ent.Name)
		}
		if !isNocturnal {
			ent.Activity = entity.EntityActivity{
				Type:      entity.ActivityIdle,
				SinceTick: sim.Tick,
				UntilTick: sim.Tick + 40,
			}
		}
		return
	}

	if isNocturnal {
		if phase == world.Night {
			goActive(ent, sim)
		} else {
			ent.Activity = entity.EntityActivity{
				Type:      entity.ActivitySleep,
				SinceTick: sim.Tick,
				UntilTick: sim.Tick + 80,
			}
		}
		return
	}

	ent.Activity = entity.EntityActivity{
		Type:      entity.ActivitySleep,
		SinceTick: sim.Tick,
		UntilTick: sim.Tick + 80,
	}
	if home != "" && ent.LocationID != home {
		ent.LocationID = home
		log.Printf("[ai] %s went home to sleep", ent.Name)
	}
}

func goActive(ent *entity.Entity, sim *Simulation) {
	home := ent.AI.HomeLocation
	if home != "" && ent.LocationID != home {
		ent.LocationID = home
	}
	setActivity(ent, sim.Tick, sim.Time.Phase())
}

func setActivity(ent *entity.Entity, tick uint64, phase world.DayPhase) {
	cycle := ent.AI.SleepCycle
	if cycle == "none" {
		ent.Activity = entity.EntityActivity{
			Type:      entity.ActivityMeditate,
			SinceTick: tick,
			UntilTick: tick + 10,
		}
		return
	}
	if cycle == "nocturnal" {
		if phase == world.Night {
			ent.Activity = entity.EntityActivity{
				Type:      entity.ActivityHunt,
				SinceTick: tick,
				UntilTick: tick + 10,
			}
		} else {
			ent.Activity = entity.EntityActivity{
				Type:      entity.ActivitySleep,
				SinceTick: tick,
				UntilTick: tick + 10,
			}
		}
		return
	}
	switch phase {
	case world.Dawn:
		ent.Activity = entity.EntityActivity{
			Type:      entity.ActivityEat,
			SinceTick: tick,
			UntilTick: tick + 10,
		}
	case world.Day:
		if ent.AI.Type == "guard" {
			ent.Activity = entity.EntityActivity{
				Type:      entity.ActivityPatrol,
				SinceTick: tick,
				UntilTick: tick + 10,
			}
		} else {
			acts := []entity.ActivityType{
				entity.ActivityWork, entity.ActivityWork, entity.ActivitySit,
				entity.ActivityCraft, entity.ActivityPaint, entity.ActivityGather,
			}
			idx := int(tick+uint64(ent.Level)*7) % len(acts)
			ent.Activity = entity.EntityActivity{
				Type:      acts[idx],
				SinceTick: tick,
				UntilTick: tick + 10,
			}
		}
	case world.Dusk:
		ent.Activity = entity.EntityActivity{
			Type:      entity.ActivityDrink,
			SinceTick: tick,
			UntilTick: tick + 10,
		}
	case world.Night:
		ent.Activity = entity.EntityActivity{
			Type:      entity.ActivitySleep,
			SinceTick: tick,
			UntilTick: tick + 10,
		}
	}
}

func roamPassive(ent *entity.Entity, sim *Simulation) {
	home := ent.AI.HomeLocation
	if home == "" {
		return
	}
	parent := sim.World.Location(ent.LocationID)
	if parent == nil {
		return
	}
	siblings := sim.World.ChildLocations(parent.ParentID)
	var targets []string
	for _, sib := range siblings {
		if sib.ID != ent.LocationID {
			targets = append(targets, sib.ID)
		}
	}
	if len(targets) == 0 {
		return
	}
	dest := targets[sim.RNG.Intn(len(targets))]
	if dest == home && sim.RNG.Intn(100) < 50 {
		return
	}
	ent.LocationID = dest
	log.Printf("[ai] %s wandered to %s", ent.Name, dest)
}

func wanderAggressive(ent *entity.Entity, sim *Simulation) {
	home := ent.AI.HomeLocation
	parent := sim.World.Location(ent.LocationID)
	if parent == nil {
		return
	}
	var targets []string
	siblings := sim.World.ChildLocations(parent.ParentID)
	for _, sib := range siblings {
		if sib.ID != ent.LocationID && sib.ID != home {
			targets = append(targets, sib.ID)
		}
	}
	if len(targets) == 0 {
		siblings = sim.World.ChildLocations(parent.ID)
		for _, sib := range siblings {
			if sib.ID != ent.LocationID {
				targets = append(targets, sib.ID)
			}
		}
	}
	if len(targets) == 0 {
		return
	}
	dest := targets[sim.RNG.Intn(len(targets))]
	ent.LocationID = dest
	log.Printf("[ai] %s prowled to %s", ent.Name, dest)
}

func formatTick(tick uint64) string {
	return fmt.Sprintf("%d", tick)
}

func simpleAttackAt(sim *Simulation, attacker, defender *entity.Entity) bool {
	combat.SetTick(sim.Tick)
	combat.ResetWeatherVisibility()
	if loc := sim.World.Location(defender.LocationID); loc != nil && loc.IsOutside {
		if wth := sim.World.EffectiveWeather(defender.LocationID); wth != nil {
			combat.SetWeatherVisibility(wth.VisibilityModifier())
		}
	}
	hit := combat.SimpleAttack(attacker, defender, sim.RNG)
	combat.ResetWeatherVisibility()
	if !defender.Alive {
		NudgeTerritoryOnKill(sim.World, defender.LocationID, attacker.Faction)
	}
	return hit
}

func rewardXP(killer, target *entity.Entity) {
	if !entity.CanLevelUp(killer.Species) {
		return
	}
	xp := 5 + target.Level*3 + target.MaxHP/10
	if xp < 1 {
		xp = 1
	}
	log.Printf("[xp] %s gained %d XP for killing %s", killer.Name, xp, target.Name)
	killer.AddXP(xp)
	killer.AddMoodModifier("combat_kill", "happy", 30)
	if killer.Faction != target.Faction {
		combat.ShiftRelation(killer.Faction, target.Faction, -1)
	}
}

func applyCombatMoods(attacker, defender *entity.Entity, hit bool, tick uint64) {
	if hit {
		attacker.AddMoodModifier("combat_hit", "angry", 10)
		defender.AddMoodModifier("combat_take_damage", "fearful", 10)
	} else {
		attacker.AddMoodModifier("combat_miss", "stressed", 5)
	}
}

func questKilled(sim *Simulation, target *entity.Entity) {
	species := target.Species
	if species == "" {
		species = "unknown"
	}
	for _, ent := range sim.Entities.All() {
		if !ent.Alive {
			continue
		}
		states := sim.Quests.EntityStates(ent.ID)
		for _, state := range states {
			if state.State != quest.StateActive {
				continue
			}
			def := sim.Quests.GetDef(state.QuestID)
			if def == nil {
				continue
			}
			for _, stage := range def.Stages {
				if stage.ID != state.CurrentStage {
					continue
				}
				for _, obj := range stage.Objectives {
					if obj.Type == "kill_entities" {
						if obj.EntityTemplate == species || obj.EntityTemplate == "*" {
							sim.Quests.ProgressObjective(ent.ID, state.QuestID, obj.ID, 1)
						}
					}
				}
			}
		}
	}
}

func processTimeLimitQuests(sim *Simulation) {
	if sim.Tick%120 != 0 {
		return
	}
	totalHours := (sim.Tick * uint64(sim.Time.Speed)) / 60
	for _, ent := range sim.Entities.All() {
		if !ent.Alive {
			continue
		}
		states := sim.Quests.EntityStates(ent.ID)
		for _, state := range states {
			if state.State != quest.StateActive {
				continue
			}
			def := sim.Quests.GetDef(state.QuestID)
			if def == nil || def.FailConditions == nil {
				continue
			}
			for _, fc := range def.FailConditions {
				if fc.Type == "time" && fc.Hours > 0 {
					elapsedHours := int(totalHours) - int(state.AcceptedTick*uint64(sim.Time.Speed)/60)
					if elapsedHours >= fc.Hours {
						state.State = quest.StateFailed
						log.Printf("[quest] '%s' failed for %s — time limit exceeded", state.QuestID, ent.Name)
					}
				}
			}
		}
	}
}
