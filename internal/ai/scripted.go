package ai

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"simuz/internal/combat"
	"simuz/internal/economy"
	"simuz/internal/entity"
	"simuz/internal/items"
	"simuz/internal/quest"
	"simuz/internal/world"
)

type ScriptManager struct {
	mu      sync.RWMutex
	scripts map[string]*lua.FunctionProto
}

var globalScripts *ScriptManager

func init() {
	globalScripts = &ScriptManager{
		scripts: make(map[string]*lua.FunctionProto),
	}
}

func LoadScript(name, source string) error {
	L := lua.NewState()
	defer L.Close()

	fn, err := L.LoadString(source)
	if err != nil {
		return fmt.Errorf("load script %s: %w", name, err)
	}

	proto := fn.Proto

	globalScripts.mu.Lock()
	defer globalScripts.mu.Unlock()
	globalScripts.scripts[name] = proto
	log.Printf("Loaded AI script: %s", name)
	return nil
}

// MoveRequest is set by the engine to handle instant vs multi-tick travel.
var MoveRequest func(ent *entity.Entity, destID string) bool

// IsEntityTraveling is set by the engine.
var IsEntityTraveling func(entityID string) bool

func RunScript(name string, ent *entity.Entity, w *world.World, em *entity.Manager, tm *world.GameTime, rng *rand.Rand, qm *quest.Manager) error {
	globalScripts.mu.RLock()
	proto, ok := globalScripts.scripts[name]
	globalScripts.mu.RUnlock()
	if !ok {
		return fmt.Errorf("script not found: %s", name)
	}

	L := lua.NewState()
	defer L.Close()

	bindEntity(L, ent, tm.Tick)
	bindWorld(L, w, em, tm, rng, ent, qm)
	bindUtils(L, rng, ent)

	closure := L.NewFunctionFromProto(proto)
	err := L.CallByParam(lua.P{
		Fn:      closure,
		NRet:    0,
		Protect: true,
	})
	if err != nil {
		return fmt.Errorf("run script %s: %w", name, err)
	}

	return nil
}

func bindEntity(L *lua.LState, e *entity.Entity, tick uint64) {
	entTbl := L.NewTable()
	entTbl.RawSetString("id", lua.LString(e.ID))
	entTbl.RawSetString("name", lua.LString(e.Name))
	entTbl.RawSetString("species", lua.LString(e.Species))
	entTbl.RawSetString("faction", lua.LString(e.Faction))
	entTbl.RawSetString("loc_id", lua.LString(e.LocationID))
	entTbl.RawSetString("hp", lua.LNumber(e.HP))
	entTbl.RawSetString("max_hp", lua.LNumber(e.MaxHP))
	entTbl.RawSetString("fp", lua.LNumber(e.FP))
	entTbl.RawSetString("max_fp", lua.LNumber(e.MaxFP))
	entTbl.RawSetString("level", lua.LNumber(e.Level))
	entTbl.RawSetString("xp", lua.LNumber(e.XP))
	entTbl.RawSetString("age", lua.LNumber(e.Age))
	if e.AI.HomeLocation != "" {
		entTbl.RawSetString("home", lua.LString(e.AI.HomeLocation))
	} else {
		entTbl.RawSetString("home", lua.LNil)
	}
	invTbl := L.NewTable()
	for _, inst := range e.Inventory {
		if inst.Def != nil && !inst.Equipped && inst.Def.Type != items.TypeCurrency {
			invTbl.Append(lua.LString(inst.DefID))
		}
	}
	entTbl.RawSetString("inventory", invTbl)
	stateTbl := L.NewTable()
	stateTbl.RawSetString("activity", lua.LString(e.Activity.Type))
	stateTbl.RawSetString("since_tick", lua.LNumber(e.Activity.SinceTick))
	stateTbl.RawSetString("until_tick", lua.LNumber(e.Activity.UntilTick))
	stateTbl.RawSetString("location_id", lua.LString(e.LocationID))
	entTbl.RawSetString("state", stateTbl)
	entTbl.RawSetString("mood", lua.LString(e.Mood))
	entTbl.RawSetString("hunger", lua.LNumber(computeHunger(e, tick)))
	skillsTbl := L.NewTable()
	for _, sname := range entity.SkillNames {
		skillsTbl.RawSetString(sname, lua.LNumber(e.SkillLevel(sname)))
	}
	entTbl.RawSetString("skills", skillsTbl)
	L.SetGlobal("self", entTbl)
}

func computeHunger(e *entity.Entity, tick uint64) float64 {
	if !entity.ShouldAutoFeed(e.Species) && e.LastMealTick > 0 {
		threshold := entity.SpeciesStarvationThreshold(e.Species)
		if threshold > 0 {
			ticksSince := int(tick) - e.LastMealTick
			if ticksSince >= threshold {
				return 1.0
			}
			return float64(ticksSince) / float64(threshold)
		}
	}
	return 0.0
}

func combatPowerScoreLua(e *entity.Entity) int {
	if e == nil {
		return 0
	}
	attrs := e.EffectiveAttrs()
	score := e.Level*4 + attrs.STR*2 + attrs.DEX + attrs.CON
	if e.MaxHP > 0 {
		score += e.HP * 10 / e.MaxHP
	}
	if e.AI.Brave {
		score += 6
	}
	return score
}

func shouldFleeCombatLua(ent *entity.Entity, hostiles []*entity.Entity) bool {
	if ent == nil || len(hostiles) == 0 || ent.AI.Brave {
		return false
	}
	if ent.MaxHP > 0 && ent.HP*100 <= ent.MaxHP*35 {
		return true
	}
	selfScore := combatPowerScoreLua(ent)
	strongest := 0
	total := 0
	for _, other := range hostiles {
		if other == nil || !other.Alive || other.Immortal {
			continue
		}
		score := combatPowerScoreLua(other)
		total += score
		if score > strongest {
			strongest = score
		}
	}
	if strongest >= selfScore+8 {
		return true
	}
	if total >= selfScore*2 {
		return true
	}
	return false
}

func chooseFleeDestinationLua(w *world.World, ent *entity.Entity, rng *rand.Rand) string {
	if w == nil || ent == nil {
		return ""
	}
	loc := w.Location(ent.LocationID)
	if loc == nil {
		return ""
	}
	if loc.ParentID != "" {
		parent := w.Location(loc.ParentID)
		if parent != nil {
			siblings := w.ChildLocations(parent.ID)
			candidates := make([]*world.Location, 0, len(siblings))
			for _, sib := range siblings {
				if sib != nil && sib.ID != ent.LocationID {
					candidates = append(candidates, sib)
				}
			}
			if len(candidates) > 0 {
				return candidates[rng.Intn(len(candidates))].ID
			}
		}
		return loc.ParentID
	}
	children := w.ChildLocations(ent.LocationID)
	if len(children) > 0 {
		return children[rng.Intn(len(children))].ID
	}
	return ""
}

func scriptCombatAttack(w *world.World, em *entity.Manager, qm *quest.Manager, rng *rand.Rand, tick uint64, attacker, target *entity.Entity) bool {
	if attacker == nil || target == nil || !attacker.Alive || !attacker.Conscious || !target.Alive {
		return false
	}
	combat.SetTick(tick)
	combat.ResetWeatherVisibility()
	if loc := w.Location(target.LocationID); loc != nil && loc.IsOutside {
		if wth := w.EffectiveWeather(target.LocationID); wth != nil {
			combat.SetWeatherVisibility(wth.VisibilityModifier())
		}
	}
	hit := combat.SimpleAttack(attacker, target, rng)
	combat.ResetWeatherVisibility()
	if !target.Alive {
		combat.LootCorpse(attacker, target)
		if attacker.Faction != target.Faction {
			combat.ShiftRelation(attacker.Faction, target.Faction, -1)
		}
		if loc := w.Location(target.LocationID); loc != nil && attacker.Faction != "" && attacker.Faction != "civilian" && attacker.Faction != "deity" {
			if loc.ControllingFaction == attacker.Faction {
				loc.ControlStrength += 3
				if loc.ControlStrength > 100 {
					loc.ControlStrength = 100
				}
			} else {
				loc.ControlStrength -= 5
				if loc.ControlStrength <= 0 {
					loc.ControllingFaction = attacker.Faction
					loc.ControlStrength = 5
				}
			}
		}
		if entity.CanLevelUp(attacker.Species) {
			xp := 5 + target.Level*3 + target.MaxHP/10
			if xp < 1 {
				xp = 1
			}
			attacker.AddXP(xp)
		}
		questKilledLua(qm, em, target)
	}
	return hit
}

func scriptFleeCombat(w *world.World, em *entity.Manager, qm *quest.Manager, rng *rand.Rand, tick uint64, ent *entity.Entity, hostiles []*entity.Entity) bool {
	if w == nil || ent == nil || len(hostiles) == 0 {
		return false
	}
	attacker := hostiles[0]
	bestScore := -1
	for _, other := range hostiles {
		if other == nil || !other.Alive || other.Immortal {
			continue
		}
		score := combatPowerScoreLua(other)
		if score > bestScore {
			bestScore = score
			attacker = other
		}
	}
	if attacker != nil {
		if !scriptCombatAttack(w, em, qm, rng, tick, attacker, ent) && !ent.Alive {
			return true
		}
		if !ent.Alive {
			return true
		}
	}
	destID := chooseFleeDestinationLua(w, ent, rng)
	if destID == "" {
		return true
	}
	if ent.MaxHP > 0 && ent.HP*100 <= ent.MaxHP*35 {
		if !ent.HasMoodModifierSource("combat_flee") {
			ent.AddMoodModifier("combat_flee", "fearful", 5)
		}
	} else if !ent.HasMoodModifierSource("combat_flee") {
		ent.AddMoodModifier("combat_flee", "stressed", 5)
	}
	if MoveRequest != nil {
		MoveRequest(ent, destID)
		return true
	}
	ent.LocationID = destID
	return true
}

func questKilledLua(qm *quest.Manager, em *entity.Manager, target *entity.Entity) {
	if qm == nil || em == nil || target == nil {
		return
	}
	species := target.Species
	if species == "" {
		species = "unknown"
	}
	for _, ent := range em.All() {
		if !ent.Alive {
			continue
		}
		states := qm.EntityStates(ent.ID)
		for _, state := range states {
			if state.State != quest.StateActive {
				continue
			}
			def := qm.GetDef(state.QuestID)
			if def == nil {
				continue
			}
			for _, stage := range def.Stages {
				if stage.ID != state.CurrentStage {
					continue
				}
				for _, obj := range stage.Objectives {
					if obj.Type == "kill_entities" && (obj.EntityTemplate == species || obj.EntityTemplate == "*") {
						qm.ProgressObjective(ent.ID, state.QuestID, obj.ID, 1)
					}
				}
			}
		}
	}
}

func bindWorld(L *lua.LState, w *world.World, em *entity.Manager, tm *world.GameTime, rng *rand.Rand, ent *entity.Entity, qm *quest.Manager) {
	worldTbl := L.NewTable()
	worldTbl.RawSetString("time", lua.LString(tm.String()))
	worldTbl.RawSetString("tick", lua.LNumber(tm.Tick))
	worldTbl.RawSetString("phase", lua.LString(tm.Phase().String()))
	worldTbl.RawSetString("day", lua.LNumber(tm.Day))
	worldTbl.RawSetString("day_of_week", lua.LNumber(tm.DayOfWeek()))
	worldTbl.RawSetString("day_name", lua.LString(tm.DayOfWeekName()))

	worldTbl.RawSetString("location_name", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		loc := w.Location(id)
		if loc == nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(loc.Name))
		return 1
	}))

	worldTbl.RawSetString("exits_from", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		tbl := L.NewTable()
		for _, child := range w.ChildLocations(id) {
			tbl.Append(lua.LString(child.ID))
		}
		L.Push(tbl)
		return 1
	}))

	worldTbl.RawSetString("travel_exits", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		loc := w.Location(id)
		tbl := L.NewTable()
		if loc != nil {
			for _, e := range loc.Exits {
				row := L.NewTable()
				row.RawSetString("target_id", lua.LString(e.TargetID))
				row.RawSetString("direction", lua.LString(e.Direction))
				row.RawSetString("distance", lua.LNumber(e.Distance))
				tbl.Append(row)
			}
		}
		// Also expose region exits when querying a non-region location
		if reg := w.RegionOf(id); reg != nil && (loc == nil || loc.Type != world.LocRegion) {
			for _, e := range reg.Exits {
				row := L.NewTable()
				row.RawSetString("target_id", lua.LString(e.TargetID))
				row.RawSetString("direction", lua.LString(e.Direction))
				row.RawSetString("distance", lua.LNumber(e.Distance))
				tbl.Append(row)
			}
		}
		L.Push(tbl)
		return 1
	}))

	worldTbl.RawSetString("weather", L.NewFunction(func(L *lua.LState) int {
		locID := ent.LocationID
		if L.GetTop() >= 1 && L.Get(1) != lua.LNil {
			locID = L.ToString(1)
		}
		wth := w.EffectiveWeather(locID)
		if wth == nil {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		tbl.RawSetString("type", lua.LString(wth.Type.String()))
		tbl.RawSetString("temperature", lua.LNumber(wth.Temperature))
		tbl.RawSetString("visibility", lua.LNumber(wth.Visibility))
		tbl.RawSetString("wind_speed", lua.LNumber(wth.WindSpeed))
		tbl.RawSetString("wind_dir", lua.LString(wth.WindDir))
		tbl.RawSetString("humidity", lua.LNumber(wth.Humidity))
		tbl.RawSetString("harsh", lua.LBool(wth.IsHarsh()))
		tbl.RawSetString("stormy", lua.LBool(wth.IsStormy()))
		tbl.RawSetString("vis_mod", lua.LNumber(wth.VisibilityModifier()))
		tbl.RawSetString("travel_mod", lua.LNumber(wth.TravelSpeedModifier()))
		L.Push(tbl)
		return 1
	}))

	worldTbl.RawSetString("location_control", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		if id == "" {
			id = ent.LocationID
		}
		loc := w.Location(id)
		if loc == nil {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		tbl.RawSetString("faction", lua.LString(loc.ControllingFaction))
		tbl.RawSetString("strength", lua.LNumber(loc.ControlStrength))
		L.Push(tbl)
		return 1
	}))

	worldTbl.RawSetString("is_traveling", L.NewFunction(func(L *lua.LState) int {
		traveling := false
		if IsEntityTraveling != nil {
			traveling = IsEntityTraveling(ent.ID)
		}
		L.Push(lua.LBool(traveling))
		return 1
	}))

	worldTbl.RawSetString("move_to", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		if w.Location(id) == nil {
			L.Push(lua.LBool(false))
			return 1
		}
		ok := false
		if MoveRequest != nil {
			ok = MoveRequest(ent, id)
		} else {
			ent.LocationID = id
			ok = true
			if qm != nil {
				qm.CheckVisitLocation(ent.ID, id)
			}
		}
		lv := L.GetGlobal("self")
		if tbl, ok2 := lv.(*lua.LTable); ok2 {
			tbl.RawSetString("loc_id", lua.LString(ent.LocationID))
			state := tbl.RawGetString("state")
			if stbl, ok3 := state.(*lua.LTable); ok3 {
				stbl.RawSetString("location_id", lua.LString(ent.LocationID))
			}
		}
		if ok && ent.LocationID == id {
			log.Printf("[lua] moved %s to %s", ent.Name, id)
			if qm != nil {
				qm.RecordActivity(ent.ID, "Moved to '"+id+"'")
			}
		}
		L.Push(lua.LBool(ok))
		return 1
	}))

	worldTbl.RawSetString("parent_location", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		loc := w.Location(id)
		if loc == nil || loc.ParentID == "" {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(loc.ParentID))
		return 1
	}))

	worldTbl.RawSetString("entities_at", L.NewFunction(func(L *lua.LState) int {
		locID := L.ToString(1)
		all := em.ByLocation(locID)
		tbl := L.NewTable()
		for _, other := range all {
			if other.Alive {
				tbl.Append(lua.LString(other.ID))
			}
		}
		L.Push(tbl)
		return 1
	}))

	worldTbl.RawSetString("nearby_entities", L.NewFunction(func(L *lua.LState) int {
		nearby := em.ByLocation(ent.LocationID)
		tbl := L.NewTable()
		for _, other := range nearby {
			if other.ID != ent.ID && other.Alive {
				tbl.Append(lua.LString(other.ID))
			}
		}
		L.Push(tbl)
		return 1
	}))

	worldTbl.RawSetString("entity_info", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		e := em.Get(id)
		if e == nil {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		tbl.RawSetString("name", lua.LString(e.Name))
		tbl.RawSetString("species", lua.LString(e.Species))
		tbl.RawSetString("faction", lua.LString(e.Faction))
		tbl.RawSetString("hp", lua.LNumber(e.HP))
		tbl.RawSetString("max_hp", lua.LNumber(e.MaxHP))
		tbl.RawSetString("level", lua.LNumber(e.Level))
		tbl.RawSetString("xp", lua.LNumber(e.XP))
		tbl.RawSetString("age", lua.LNumber(e.Age))
		tbl.RawSetString("alive", lua.LBool(e.Alive))
		tbl.RawSetString("conscious", lua.LBool(e.Conscious))
		tbl.RawSetString("location_id", lua.LString(e.LocationID))
		tbl.RawSetString("hunger", lua.LNumber(computeHunger(e, tm.Tick)))
		L.Push(tbl)
		return 1
	}))

	worldTbl.RawSetString("is_hostile", L.NewFunction(func(L *lua.LState) int {
		a := L.ToString(1)
		b := L.ToString(2)
		L.Push(lua.LBool(combat.Relation(a, b) == combat.Hostile))
		return 1
	}))

	worldTbl.RawSetString("defend_self", L.NewFunction(func(L *lua.LState) int {
		nearby := em.ByLocation(ent.LocationID)
		hostiles := make([]*entity.Entity, 0, len(nearby))
		for _, other := range nearby {
			if other == nil || other.ID == ent.ID || !other.Alive || other.Immortal {
				continue
			}
			if combat.Relation(ent.Faction, other.Faction) != combat.Hostile {
				continue
			}
			hostiles = append(hostiles, other)
		}
		if len(hostiles) == 0 {
			L.Push(lua.LFalse)
			return 1
		}
		if shouldFleeCombatLua(ent, hostiles) {
			L.Push(lua.LBool(scriptFleeCombat(w, em, qm, rng, tm.Tick, ent, hostiles)))
			return 1
		}
		target := hostiles[0]
		L.Push(lua.LBool(scriptCombatAttack(w, em, qm, rng, tm.Tick, ent, target)))
		return 1
	}))

	worldTbl.RawSetString("set_relation", L.NewFunction(func(L *lua.LState) int {
		a := L.ToString(1)
		b := L.ToString(2)
		rel := L.ToString(3)
		var r combat.FactionRelation
		switch rel {
		case "hostile":
			r = combat.Hostile
		case "friendly":
			r = combat.Friendly
		default:
			r = combat.Neutral
		}
		combat.SetRelation(a, b, r)
		log.Printf("[lua] %s set relation %s <-> %s = %s", ent.Name, a, b, rel)
		return 0
	}))

	worldTbl.RawSetString("attack", L.NewFunction(func(L *lua.LState) int {
		attackerID := L.ToString(1)
		targetID := L.ToString(2)
		attacker := em.Get(attackerID)
		target := em.Get(targetID)
		if attacker == nil || target == nil || !attacker.Alive || !attacker.Conscious || !target.Alive {
			L.Push(lua.LFalse)
			return 1
		}
		hit := scriptCombatAttack(w, em, qm, rng, tm.Tick, attacker, target)
		L.Push(lua.LBool(hit))
		return 1
	}))

	worldTbl.RawSetString("divine_intervention", L.NewFunction(func(L *lua.LState) int {
		deityID := L.ToString(1)
		targetID := L.ToString(2)
		itype := L.ToString(3)

		deity := em.Get(deityID)
		target := em.Get(targetID)
		result := L.NewTable()

		if deity == nil || target == nil || !deity.Alive || !target.Alive {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}

		switch itype {
		case "heal":
			deityAttrs := deity.EffectiveAttrs()
			amt := 10 + deity.Level + (deityAttrs.WIS / 2)
			target.Heal(amt)
			result.RawSetString("done", lua.LTrue)
			result.RawSetString("amount", lua.LNumber(amt))
			log.Printf("[divine] %s healed %s for %d HP", deity.Name, target.Name, amt)

		case "bless":
			deityAttrs := deity.EffectiveAttrs()
			dur := 30 + deityAttrs.WIS
			target.Flags["blessed"] = tm.Tick + uint64(dur)
			result.RawSetString("done", lua.LTrue)
			result.RawSetString("duration", lua.LNumber(dur))
			log.Printf("[divine] %s blessed %s for %d ticks", deity.Name, target.Name, dur)

		case "smite":
			nearby := em.ByLocation(target.LocationID)
			count := 0
			for _, other := range nearby {
				if other.ID == target.ID || other.ID == deityID || !other.Alive || other.Immortal {
					continue
				}
				dmg := 5 + deity.Level + (deity.Attributes.STR / 3)
				other.TakeDamage(dmg)
				if !other.Alive {
					log.Printf("[divine] %s smote %s for %d — slain!", deity.Name, other.Name, dmg)
				} else {
					log.Printf("[divine] %s smote %s for %d (%d HP left)", deity.Name, other.Name, dmg, other.HP)
				}
				count++
			}
			result.RawSetString("done", lua.LTrue)
			result.RawSetString("targets", lua.LNumber(count))
			log.Printf("[divine] %s smote %d hostiles near %s", deity.Name, count, target.Name)

		case "scare":
			nearby := em.ByLocation(target.LocationID)
			count := 0
			for _, other := range nearby {
				if other.ID == target.ID || other.ID == deityID || !other.Alive || other.Immortal {
					continue
				}
				if other.Species == "deity" || other.Faction == "deity" {
					continue
				}
				fearDmg := 3 + deity.Level/2
				other.TakeDamage(fearDmg)
				count++
				log.Printf("[divine] %s frightened %s for %d fear damage", deity.Name, other.Name, fearDmg)
			}
			result.RawSetString("done", lua.LTrue)
			result.RawSetString("targets", lua.LNumber(count))
			log.Printf("[divine] %s scared %d mortals near %s", deity.Name, count, target.Name)

		case "quest":
			questID := L.ToString(4)
			if qm != nil {
				accepted := qm.Accept(targetID, questID, target.Level, uint64(tm.Tick))
				result.RawSetString("done", lua.LBool(accepted))
				result.RawSetString("quest_id", lua.LString(questID))
				if accepted {
					log.Printf("[divine] %s granted quest '%s' to %s", deity.Name, questID, target.Name)
				}
			} else {
				result.RawSetString("done", lua.LFalse)
				result.RawSetString("error", lua.LString("quest manager unavailable"))
			}

		default:
			result.RawSetString("done", lua.LFalse)
			result.RawSetString("error", lua.LString("unknown intervention type: "+itype))
		}

		L.Push(result)
		return 1
	}))

	worldTbl.RawSetString("entity_name", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		e := em.Get(id)
		if e == nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(e.Name))
		return 1
	}))

	worldTbl.RawSetString("entity_items", L.NewFunction(func(L *lua.LState) int {
		targetID := L.ToString(1)
		target := em.Get(targetID)
		if target == nil {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		for _, inst := range target.Inventory {
			if inst.Def != nil && !inst.Equipped && inst.Def.Type != items.TypeCurrency {
				tbl.Append(lua.LString(inst.DefID))
			}
		}
		L.Push(tbl)
		return 1
	}))

	worldTbl.RawSetString("try_buy", L.NewFunction(func(L *lua.LState) int {
		sellerID := L.ToString(1)
		itemDefID := L.ToString(2)
		result := L.NewTable()
		seller := em.Get(sellerID)
		if seller == nil || !seller.Alive {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		idx, found := economy.HasItem(seller, itemDefID)
		if !found || idx < 0 || idx >= len(seller.Inventory) {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		inst := seller.Inventory[idx]
		if inst.Def == nil {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		basePrice := inst.Def.Value
		if basePrice <= 0 {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		price, _ := economy.BuyPrice(basePrice, seller, ent)
		if !economy.CanAfford(ent, price) {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		economy.RemoveCurrency(ent, price)
		economy.AddCurrency(seller, price)
		economy.TransferItem(seller, ent, idx)
		result.RawSetString("done", lua.LTrue)
		result.RawSetString("price", lua.LNumber(price))
		log.Printf("[lua] %s bought %s from %s (%s) for %d", ent.Name, itemDefID, seller.Name, sellerID, price)
		L.Push(result)
		return 1
	}))

	worldTbl.RawSetString("try_sell", L.NewFunction(func(L *lua.LState) int {
		buyerID := L.ToString(1)
		itemDefID := L.ToString(2)
		result := L.NewTable()
		buyer := em.Get(buyerID)
		if buyer == nil || !buyer.Alive {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		idx, found := economy.HasItem(ent, itemDefID)
		if !found || idx < 0 || idx >= len(ent.Inventory) {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		inst := ent.Inventory[idx]
		if inst.Def == nil {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		basePrice := inst.Def.Value
		if basePrice <= 0 {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		price, _ := economy.SellPrice(basePrice, ent, buyer)
		if !economy.CanAfford(buyer, price) {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		economy.RemoveCurrency(buyer, price)
		economy.AddCurrency(ent, price)
		economy.TransferItem(ent, buyer, idx)
		result.RawSetString("done", lua.LTrue)
		result.RawSetString("price", lua.LNumber(price))
		log.Printf("[lua] %s sold %s to %s (%s) for %d", ent.Name, itemDefID, buyer.Name, buyerID, price)
		L.Push(result)
		return 1
	}))

	worldTbl.RawSetString("add_item", L.NewFunction(func(L *lua.LState) int {
		itemDefID := L.ToString(1)
		itemDef := items.GetDef(itemDefID)
		if itemDef == nil {
			L.Push(lua.LFalse)
			return 1
		}
		inst := items.NewItemInstance(itemDefID+"_"+strconv.Itoa(len(ent.Inventory)), itemDefID, itemDef, 1)
		ent.AddItem(inst)
		if qm != nil {
			qm.CheckCollectItem(ent.ID, itemDefID)
		}
		log.Printf("[lua] %s added %s to inventory", ent.Name, itemDefID)
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("steal", L.NewFunction(func(L *lua.LState) int {
		targetID := L.ToString(1)
		itemDefID := L.ToString(2)
		result := L.NewTable()
		target := em.Get(targetID)
		if target == nil || !target.Alive {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		idx, found := economy.HasItem(target, itemDefID)
		if !found {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		inst := target.Inventory[idx]
		if inst.Def == nil {
			result.RawSetString("done", lua.LFalse)
			L.Push(result)
			return 1
		}
		if inst.Count > 1 {
			target.Inventory[idx].Count--
		} else {
			target.Inventory = append(target.Inventory[:idx], target.Inventory[idx+1:]...)
		}
		newInst := items.NewItemInstance(itemDefID+"_"+ent.ID+"_"+strconv.Itoa(len(ent.Inventory)), itemDefID, inst.Def, 1)
		ent.AddItem(newInst)
		result.RawSetString("done", lua.LTrue)
		log.Printf("[lua] %s stole %s from %s", ent.Name, itemDefID, target.Name)
		L.Push(result)
		return 1
	}))

	worldTbl.RawSetString("damage_location", L.NewFunction(func(L *lua.LState) int {
		attackerID := L.ToString(1)
		amount := L.ToInt(2)
		result := L.NewTable()
		attacker := em.Get(attackerID)
		if attacker == nil || !attacker.Alive {
			result.RawSetString("targets", lua.LNumber(0))
			L.Push(result)
			return 1
		}
		nearby := em.ByLocation(attacker.LocationID)
		count := 0
		for _, other := range nearby {
			if other.ID == attacker.ID || !other.Alive || other.Immortal {
				continue
			}
			if combat.Relation(attacker.Faction, other.Faction) == combat.Friendly {
				continue
			}
			other.TakeDamage(amount)
			count++
			if !other.Alive {
				combat.LootCorpse(attacker, other)
				if attacker.Faction != other.Faction {
					combat.ShiftRelation(attacker.Faction, other.Faction, -1)
				}
				if entity.CanLevelUp(attacker.Species) {
					xp := 5 + other.Level*3 + other.MaxHP/10
					if xp < 1 {
						xp = 1
					}
					attacker.AddXP(xp)
				}
			}
		}
		result.RawSetString("targets", lua.LNumber(count))
		log.Printf("[lua] %s dealt %d area damage to %d targets", attacker.Name, amount, count)
		L.Push(result)
		return 1
	}))

	worldTbl.RawSetString("use_item", L.NewFunction(func(L *lua.LState) int {
		itemDefID := L.ToString(1)
		idx := -1
		for i, inst := range ent.Inventory {
			if inst.DefID == itemDefID && !inst.Equipped {
				idx = i
				break
			}
		}
		if idx < 0 {
			L.Push(lua.LFalse)
			return 1
		}
		inst := ent.Inventory[idx]
		if inst.Def == nil || inst.Def.Substance == nil {
			L.Push(lua.LFalse)
			return 1
		}
		se := inst.Def.Substance
		boostMod := entity.Attributes{
			STR: se.STR, DEX: se.DEX, CON: se.CON,
			INT: se.INT, WIS: se.WIS, CHA: se.CHA,
		}
		crashMod := entity.Attributes{
			STR: se.CrashSTR, DEX: se.CrashDEX, CON: se.CrashCON,
			INT: se.CrashINT, WIS: se.CrashWIS, CHA: se.CrashCHA,
		}
		ent.ApplySubstance(se.Name, inst.DefID, se.Duration, se.CrashDuration, boostMod, crashMod, se.HealHP, se.HealFP, se.HealPerTick, se.FPPerTick, uint64(tm.Tick))
		ent.Inventory = append(ent.Inventory[:idx], ent.Inventory[idx+1:]...)
		log.Printf("[lua] %s used %s — %s", ent.Name, inst.Def.Name, se.Name)
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("talk_to", L.NewFunction(func(L *lua.LState) int {
		npcID := L.ToString(1)
		if qm != nil {
			qm.RecordActivity(ent.ID, "Tried to talk to '"+npcID+"'")
			qm.CheckTalkToNPC(ent.ID, npcID)
		}
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("deliver_item", L.NewFunction(func(L *lua.LState) int {
		npcID := L.ToString(1)
		itemDefID := L.ToString(2)
		if qm != nil {
			qm.RecordActivity(ent.ID, "Tried to deliver '"+itemDefID+"' to '"+npcID+"'")
			qm.CheckDeliverItem(ent.ID, npcID, itemDefID)
			for i, inst := range ent.Inventory {
				if inst.DefID == itemDefID && !inst.Equipped {
					if inst.Count > 1 {
						ent.Inventory[i].Count--
					} else {
						ent.Inventory = append(ent.Inventory[:i], ent.Inventory[i+1:]...)
					}
					break
				}
			}
		}
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("feed", L.NewFunction(func(L *lua.LState) int {
		targetID := L.ToString(1)
		target := em.Get(targetID)
		if target == nil || !target.Alive {
			L.Push(lua.LFalse)
			return 1
		}
		target.LastMealTick = int(tm.Tick)
		log.Printf("[lua] %s fed %s", ent.Name, target.Name)
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("heal", L.NewFunction(func(L *lua.LState) int {
		targetID := L.ToString(1)
		amount := L.ToInt(2)
		target := em.Get(targetID)
		if target == nil || !target.Alive {
			L.Push(lua.LFalse)
			return 1
		}
		target.Heal(amount)
		log.Printf("[lua] %s healed %s for %d HP", ent.Name, target.Name, amount)
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("give_quest", L.NewFunction(func(L *lua.LState) int {
		targetID := L.ToString(1)
		questID := L.ToString(2)
		target := em.Get(targetID)
		if target == nil || !target.Alive {
			L.Push(lua.LFalse)
			return 1
		}
		if qm == nil {
			L.Push(lua.LFalse)
			return 1
		}
		accepted := qm.Accept(targetID, questID, target.Level, uint64(tm.Tick))
		if accepted {
			log.Printf("[lua] %s gave quest '%s' to %s", ent.Name, questID, target.Name)
		}
		L.Push(lua.LBool(accepted))
		return 1
	}))

	worldTbl.RawSetString("quest_progress", L.NewFunction(func(L *lua.LState) int {
		questID := L.ToString(1)
		objID := L.ToString(2)
		delta := L.ToInt(3)
		if qm == nil {
			L.Push(lua.LFalse)
			return 1
		}
		qm.ProgressObjective(ent.ID, questID, objID, delta)
		log.Printf("[lua] %s progressed quest '%s' objective '%s' by %d", ent.Name, questID, objID, delta)
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("quest_set", L.NewFunction(func(L *lua.LState) int {
		questID := L.ToString(1)
		objID := L.ToString(2)
		value := L.ToInt(3)
		if qm == nil {
			L.Push(lua.LFalse)
			return 1
		}
		qm.SetObjective(ent.ID, questID, objID, value)
		log.Printf("[lua] %s set quest '%s' objective '%s' to %d", ent.Name, questID, objID, value)
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("craft", L.NewFunction(func(L *lua.LState) int {
		recipeID := L.ToString(1)
		var recipe *items.Recipe
		for _, r := range items.Recipes {
			if r.ID == recipeID {
				recipe = r
				break
			}
		}
		if recipe == nil {
			L.Push(lua.LFalse)
			return 1
		}
		if !items.HasMaterials(ent.Inventory, recipe.Inputs) {
			L.Push(lua.LFalse)
			return 1
		}
		if recipe.Station != "" {
			loc := w.Location(ent.LocationID)
			if loc == nil {
				L.Push(lua.LFalse)
				return 1
			}
			hasStation := false
			for _, tag := range loc.Tags {
				if tag == recipe.Station {
					hasStation = true
					break
				}
			}
			if !hasStation {
				L.Push(lua.LFalse)
				return 1
			}
		}
		ent.Inventory = items.RemoveInputs(ent.Inventory, recipe.Inputs)
		outputDef := items.GetDef(recipe.Output.DefID)
		if outputDef != nil {
			inst := items.NewItemInstance("craft_"+ent.ID+"_"+recipe.ID, outputDef.ID, outputDef, recipe.Output.Count)
			ent.Inventory = append(ent.Inventory, inst)
		}
		log.Printf("[lua] %s crafted %s", ent.Name, recipe.Name)
		skillMap := map[string]string{"forge": "smithing", "cauldron": "alchemy", "campfire": "cooking", "workbench": "tailoring"}
		if sname, ok := skillMap[recipe.Station]; ok {
			ent.AddSkillXP(sname, 20)
		}
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("has_material", L.NewFunction(func(L *lua.LState) int {
		defID := L.ToString(1)
		for _, inst := range ent.Inventory {
			if inst.DefID == defID && !inst.Equipped && inst.Count > 0 {
				L.Push(lua.LTrue)
				return 1
			}
		}
		L.Push(lua.LFalse)
		return 1
	}))

	worldTbl.RawSetString("recipe_info", L.NewFunction(func(L *lua.LState) int {
		recipeID := L.ToString(1)
		var recipe *items.Recipe
		for _, r := range items.Recipes {
			if r.ID == recipeID {
				recipe = r
				break
			}
		}
		if recipe == nil {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		tbl.RawSetString("id", lua.LString(recipe.ID))
		tbl.RawSetString("name", lua.LString(recipe.Name))
		tbl.RawSetString("station", lua.LString(recipe.Station))
		tbl.RawSetString("output_def", lua.LString(recipe.Output.DefID))
		tbl.RawSetString("output_count", lua.LNumber(recipe.Output.Count))
		inputTbl := L.NewTable()
		for i, inp := range recipe.Inputs {
			it := L.NewTable()
			it.RawSetString("def_id", lua.LString(inp.DefID))
			it.RawSetString("count", lua.LNumber(inp.Count))
			inputTbl.RawSetInt(i+1, it)
		}
		tbl.RawSetString("inputs", inputTbl)
		L.Push(tbl)
		return 1
	}))

	worldTbl.RawSetString("drag_entity", L.NewFunction(func(L *lua.LState) int {
		targetID := L.ToString(1)
		target := em.Get(targetID)
		if target == nil || !target.Alive {
			L.Push(lua.LFalse)
			return 1
		}
		target.LeashedBy = ent.ID
		log.Printf("[lua] %s started dragging/leashing %s", ent.Name, target.Name)
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("undrag_entity", L.NewFunction(func(L *lua.LState) int {
		targetID := L.ToString(1)
		target := em.Get(targetID)
		if target == nil {
			L.Push(lua.LFalse)
			return 1
		}
		if target.LeashedBy == ent.ID || targetID == "" {
			target.LeashedBy = ""
			log.Printf("[lua] %s released drag on %s", ent.Name, target.Name)
			L.Push(lua.LTrue)
			return 1
		}
		L.Push(lua.LFalse)
		return 1
	}))

	worldTbl.RawSetString("is_leashed", L.NewFunction(func(L *lua.LState) int {
		targetID := ent.ID
		if L.GetTop() >= 1 && L.Get(1) != lua.LNil {
			targetID = L.ToString(1)
		}
		target := em.Get(targetID)
		if target == nil || target.LeashedBy == "" {
			L.Push(lua.LFalse)
			L.Push(lua.LNil)
			return 2
		}
		L.Push(lua.LTrue)
		L.Push(lua.LString(target.LeashedBy))
		return 2
	}))

	worldTbl.RawSetString("start_rescue", L.NewFunction(func(L *lua.LState) int {
		targetID := L.ToString(1)
		target := em.Get(targetID)
		if target == nil || !target.Alive {
			L.Push(lua.LFalse)
			return 1
		}
		target.RescueState = "in_progress"
		log.Printf("[lua] %s started rescue of %s", ent.Name, target.Name)
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("complete_rescue", L.NewFunction(func(L *lua.LState) int {
		targetID := L.ToString(1)
		target := em.Get(targetID)
		if target == nil {
			L.Push(lua.LFalse)
			return 1
		}
		target.RescueState = "completed"
		target.LeashedBy = ""
		log.Printf("[lua] %s completed rescue of %s", ent.Name, target.Name)
		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("world", worldTbl)
}

func bindUtils(L *lua.LState, rng *rand.Rand, e *entity.Entity) {
	utilTbl := L.NewTable()

	utilTbl.RawSetString("rand_int", L.NewFunction(func(L *lua.LState) int {
		n := L.ToInt(1)
		if n <= 0 {
			n = 1
		}
		L.Push(lua.LNumber(rng.Intn(n)))
		return 1
	}))

	utilTbl.RawSetString("log", L.NewFunction(func(L *lua.LState) int {
		msg := L.ToString(1)
		log.Printf("[lua] %s", msg)
		return 0
	}))

	utilTbl.RawSetString("mem_set", L.NewFunction(func(L *lua.LState) int {
		key := L.ToString(1)
		val := luaValueToGo(L, L.Get(2))
		e.Flags[key] = val
		return 0
	}))

	utilTbl.RawSetString("mem_get", L.NewFunction(func(L *lua.LState) int {
		key := L.ToString(1)
		val, ok := e.Flags[key]
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		goValueToLua(L, val)
		return 1
	}))

	utilTbl.RawSetString("set_mood", L.NewFunction(func(L *lua.LState) int {
		mood := L.ToString(1)
		duration := 30
		if L.GetTop() >= 2 {
			duration = L.ToInt(2)
			if duration < 1 {
				duration = 30
			}
		}
		e.AddMoodModifier("lua", mood, uint64(duration))
		lv := L.GetGlobal("self")
		if tbl, ok := lv.(*lua.LTable); ok {
			tbl.RawSetString("mood", lua.LString(mood))
		}
		return 0
	}))

	utilTbl.RawSetString("json_encode", L.NewFunction(func(L *lua.LState) int {
		val := luaValueToGo(L, L.Get(1))
		b, err := json.Marshal(val)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(string(b)))
		return 1
	}))

	utilTbl.RawSetString("json_decode", L.NewFunction(func(L *lua.LState) int {
		s := L.ToString(1)
		var val any
		if err := json.Unmarshal([]byte(s), &val); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		goValueToLua(L, val)
		return 1
	}))

	L.SetGlobal("util", utilTbl)
}

func luaValueToGo(L *lua.LState, lv lua.LValue) any {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case *lua.LTable:
		if v.Len() > 0 {
			arr := make([]any, 0, v.Len())
			v.ForEach(func(_ lua.LValue, val lua.LValue) {
				arr = append(arr, luaValueToGo(L, val))
			})
			return arr
		}
		m := make(map[string]any)
		v.ForEach(func(key lua.LValue, val lua.LValue) {
			if k, ok := key.(lua.LString); ok {
				m[string(k)] = luaValueToGo(L, val)
			}
		})
		return m
	default:
		return lv.String()
	}
}

func goValueToLua(L *lua.LState, gv any) {
	switch val := gv.(type) {
	case nil:
		L.Push(lua.LNil)
	case bool:
		L.Push(lua.LBool(val))
	case string:
		L.Push(lua.LString(val))
	case float64:
		L.Push(lua.LNumber(val))
	case json.Number:
		f, _ := strconv.ParseFloat(string(val), 64)
		L.Push(lua.LNumber(f))
	case map[string]any:
		tbl := L.NewTable()
		for k, vv := range val {
			tbl.RawSetString(k, toLuaValue(L, vv))
		}
		L.Push(tbl)
	case []any:
		tbl := L.NewTable()
		for _, vv := range val {
			tbl.Append(toLuaValue(L, vv))
		}
		L.Push(tbl)
	default:
		L.Push(lua.LString(fmt.Sprintf("%v", val)))
	}
}

func toLuaValue(L *lua.LState, v any) lua.LValue {
	switch val := v.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(val)
	case string:
		return lua.LString(val)
	case float64:
		return lua.LNumber(val)
	case json.Number:
		f, _ := strconv.ParseFloat(string(val), 64)
		return lua.LNumber(f)
	case map[string]any:
		tbl := L.NewTable()
		for k, vv := range val {
			tbl.RawSetString(k, toLuaValue(L, vv))
		}
		return tbl
	case []any:
		tbl := L.NewTable()
		for _, vv := range val {
			tbl.Append(toLuaValue(L, vv))
		}
		return tbl
	default:
		return lua.LString(fmt.Sprintf("%v", val))
	}
}
