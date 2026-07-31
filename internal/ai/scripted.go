// Package ai contains the AI runtime, script loading, and Lua-facing helpers for entities.
package ai

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"simuz/internal/combat"
	"simuz/internal/economy"
	"simuz/internal/entity"
	"simuz/internal/events"
	"simuz/internal/faction"
	"simuz/internal/items"
	"simuz/internal/quest"
	"simuz/internal/relation"
	"simuz/internal/species"
	"simuz/internal/world"

	lua "github.com/yuin/gopher-lua"
)

type ScriptManager struct {
	scripts     map[string]*lua.FunctionProto
	scriptTypes map[string]string
}

var globalScripts *ScriptManager

func init() {
	globalScripts = &ScriptManager{
		scripts:     make(map[string]*lua.FunctionProto),
		scriptTypes: make(map[string]string),
	}

}
func goValue(value lua.LValue) any {
	switch v := value.(type) {
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case lua.LBool:
		return bool(v)
	default:
		return nil // Handles lua.LNil
	}
}

// MoveRequest is set by the engine to handle instant vs multi-tick travel.
var MoveRequest func(ent *entity.Entity, destID string) bool

// IsEntityTraveling is set by the engine.
var IsEntityTraveling func(entityID string) bool

// setFactionHostility sets the faction-level hostility relation between two factions bidirectionally.
func setFactionHostility(factionA, factionB string, rel relation.HostilityRelation) {
	if factionA == factionB {
		return
	}
	if fac, ok := faction.GetFactionByID(factionA); ok {
		fac.Relation.SetFactionRelation(factionB, rel)
	}
	if fac, ok := faction.GetFactionByID(factionB); ok {
		fac.Relation.SetFactionRelation(factionA, rel)
	}
}

// factionHostility returns the combined faction-level hostility between two factions.
// Positive means friendly, negative means hostile, zero means neutral.
func factionHostility(factionA, factionB string) relation.HostilityRelation {
	if factionA == factionB {
		return 1
	}
	var combined relation.HostilityRelation
	if fac, ok := faction.GetFactionByID(factionA); ok {
		combined += fac.Relation.GetFactionRelation(factionB)
	}
	if fac, ok := faction.GetFactionByID(factionB); ok {
		combined += fac.Relation.GetFactionRelation(factionA)
	}
	return combined
}

// IsHostile checks whether factionA is hostile toward factionB.
func IsHostile(factionA, factionB string) bool {
	return factionHostility(factionA, factionB) < 0
}

// ScriptResult holds the outcome of a single script execution.
type ScriptResult struct {
	DidAct bool
	Events []*events.SimEvent
}

// RunScript executes a loaded Lua AI script and returns whether the script
// performed an action (didAct) along with any events it generated.
// Scripts can return up to three values: didAct (bool), log messages, events (table).
func RunScript(name string, ent *entity.Entity, w *world.World, em *entity.EntityManager, tm *world.GameTime, rng *rand.Rand, qm *quest.Manager) (ScriptResult, error) {
	proto, ok := globalScripts.scripts[name]

	if !ok {
		return ScriptResult{}, fmt.Errorf("script not found: %s", name)
	}

	L := lua.NewState()
	defer L.Close()

	bindEntity(L, ent, tm.Tick)
	bindWorld(L, w, em, tm, rng, ent, qm)
	bindUtils(L, rng, ent)

	closure := L.NewFunctionFromProto(proto)
	err := L.CallByParam(lua.P{
		Fn:      closure,
		NRet:    3,
		Protect: true,
	})
	if err != nil {
		return ScriptResult{}, fmt.Errorf("run script %s: %w", name, err)
	}

	var result ScriptResult

	// 1st return value: didAct (boolean)
	didActVal := L.Get(-3)
	if didActVal != lua.LNil {
		result.DidAct = lua.LVAsBool(didActVal)
	}

	// 3rd return value: events (table)
	val := L.Get(-1)
	if tbl, ok := val.(*lua.LTable); ok {
		result.Events = decodeSimEvents(tbl, tm.Tick)
	}

	return result, nil
}

// decodeSimEvents converts a Lua table of event tables into []*events.SimEvent.
// Each entry in the Lua table should have keys: type (int), tick (uint64), source (string), data (table).
func decodeSimEvents(tbl *lua.LTable, defaultTick uint64) []*events.SimEvent {
	var luaEvents []*events.SimEvent
	tbl.ForEach(func(k, v lua.LValue) {
		eventTbl, ok := v.(*lua.LTable)
		if !ok {
			return
		}
		ev := &events.SimEvent{
			Tick: defaultTick,
		}

		// type (EventType int)
		if typeVal := eventTbl.RawGetString("type"); typeVal != lua.LNil {
			typeInt, err := strconv.Atoi(typeVal.String())
			if err != nil {
				log.Printf("Error converting type to int: %v", err)
				return
			}
			ev.Type = events.EventType(typeInt)
		}

		// tick (uint64)
		if tickVal := eventTbl.RawGetString("tick"); tickVal != lua.LNil {
			tickInt, err := strconv.Atoi(tickVal.String())
			if err != nil {
				log.Printf("Error converting tick to int: %v", err)
				return
			}
			ev.Tick = uint64(tickInt)
		}

		// source (string)
		if srcVal := eventTbl.RawGetString("source"); srcVal != lua.LNil {
			ev.Source = srcVal.String()
		}

		// data (table -> map[string]any)
		if dataVal := eventTbl.RawGet(lua.LString("data")); dataVal != lua.LNil {
			result := make(map[string]any)
			table, ok := dataVal.(*lua.LTable)
			if !ok {
				log.Printf("Error: data is not a table")
				return
			}
			// Iterate through the Lua table keys and values
			table.ForEach(func(key lua.LValue, val lua.LValue) {
				// Ensure the key is a string (ignores numeric array indices if necessary)
				if strKey, ok := key.(lua.LString); ok {
					result[string(strKey)] = goValue(val)
				}
			})
		}

		luaEvents = append(luaEvents, ev)
	})
	return luaEvents
}

// luaTableToMap converts a Lua table with string keys to a Go map[string]any.
func luaTableToMap(L *lua.LState, tbl *lua.LTable) map[string]any {
	m := make(map[string]any)
	tbl.ForEach(func(k, v lua.LValue) {
		key, ok := k.(lua.LString)
		if !ok {
			return
		}
		m[string(key)] = luaValueToGo(L, v)
	})
	return m
}

func bindEntity(L *lua.LState, e *entity.Entity, tick uint64) {
	entTbl := L.NewTable()
	entTbl.RawSetString("id", lua.LString(e.ID))
	entTbl.RawSetString("name", lua.LString(e.Name))
	entTbl.RawSetString("species", lua.LString(e.Species))
	entTbl.RawSetString("faction", lua.LString(e.Faction))
	entTbl.RawSetString("profession", lua.LString(e.Profession))
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

	// Relationships
	entTbl.RawSetString("get_relationship", L.NewFunction(func(L *lua.LState) int {
		otherID := L.ToString(1)
		if otherID == "" {
			L.Push(lua.LNil)
			return 1
		}
		rel, ok := e.GetRelationship(otherID)
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		tbl.RawSetString("other_id", lua.LString(rel.OtherID))
		tbl.RawSetString("type", lua.LString(string(rel.Type)))
		tbl.RawSetString("since_tick", lua.LNumber(rel.SinceTick))
		L.Push(tbl)
		return 1
	}))

	entTbl.RawSetString("get_relationship_to", L.NewFunction(func(L *lua.LState) int {
		relationshipType := L.ToString(1)
		if relationshipType == "" {
			L.Push(lua.LNil)
			return 1
		}
		rel, ok := e.GetRelationshipTo(entity.RelationshipType(relationshipType))
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		tbl.RawSetString("other_id", lua.LString(rel.OtherID))
		tbl.RawSetString("type", lua.LString(string(rel.Type)))
		tbl.RawSetString("since_tick", lua.LNumber(rel.SinceTick))
		L.Push(tbl)
		return 1
	}))

	entTbl.RawSetString("get_relationships", L.NewFunction(func(L *lua.LState) int {
		tbl := L.NewTable()
		for otherID, rel := range e.Relationships {
			row := L.NewTable()
			row.RawSetString("other_id", lua.LString(rel.OtherID))
			row.RawSetString("type", lua.LString(string(rel.Type)))
			row.RawSetString("since_tick", lua.LNumber(rel.SinceTick))
			tbl.RawSetString(otherID, row)
		}
		L.Push(tbl)
		return 1
	}))

	entTbl.RawSetString("get_children", L.NewFunction(func(L *lua.LState) int {
		tbl := L.NewTable()
		for _, rel := range e.GetChildren() {
			tbl.Append(lua.LString(rel.OtherID))
		}
		L.Push(tbl)
		return 1
	}))

	entTbl.RawSetString("get_parents", L.NewFunction(func(L *lua.LState) int {
		tbl := L.NewTable()
		for _, rel := range e.GetParents() {
			tbl.Append(lua.LString(rel.OtherID))
		}
		L.Push(tbl)
		return 1
	}))

	entTbl.RawSetString("get_partner", L.NewFunction(func(L *lua.LState) int {
		rel, ok := e.GetPartner()
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(rel.OtherID))
		return 1
	}))

	entTbl.RawSetString("get_relationship_type", L.NewFunction(func(L *lua.LState) int {
		otherID := L.ToString(1)
		if otherID == "" {
			L.Push(lua.LNil)
			return 1
		}
		rel, ok := e.GetRelationship(otherID)
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(string(rel.Type)))
		return 1
	}))

	entTbl.RawSetString("get_relationship_since", L.NewFunction(func(L *lua.LState) int {
		otherID := L.ToString(1)
		if otherID == "" {
			L.Push(lua.LNil)
			return 1
		}
		rel, ok := e.GetRelationship(otherID)
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LNumber(rel.SinceTick))
		return 1
	}))

	entTbl.RawSetString("has_relationship", L.NewFunction(func(L *lua.LState) int {
		otherID := L.ToString(1)
		if otherID == "" {
			L.Push(lua.LFalse)
			return 1
		}
		_, ok := e.GetRelationship(otherID)
		L.Push(lua.LBool(ok))
		return 1
	}))

	entTbl.RawSetString("has_relationship_type", L.NewFunction(func(L *lua.LState) int {
		otherID := L.ToString(1)
		relType := entity.RelationshipType(L.ToString(2))
		if otherID == "" || relType == "" {
			L.Push(lua.LFalse)
			return 1
		}
		rel, ok := e.GetRelationship(otherID)
		if !ok {
			L.Push(lua.LFalse)
			return 1
		}
		L.Push(lua.LBool(rel.Type == relType))
		return 1
	}))

	entTbl.RawSetString("add_relationship", L.NewFunction(func(L *lua.LState) int {
		otherID := L.ToString(1)
		relTypeStr := L.ToString(2)
		sinceTick := uint64(L.ToInt(3))
		if otherID == "" || relTypeStr == "" {
			L.Push(lua.LFalse)
			return 1
		}
		relType := entity.RelationshipType(relTypeStr)
		e.AddRelationship(otherID, relType, sinceTick)
		L.Push(lua.LTrue)
		return 1
	}))

	entTbl.RawSetString("remove_relationship", L.NewFunction(func(L *lua.LState) int {
		otherID := L.ToString(1)
		if otherID == "" {
			L.Push(lua.LFalse)
			return 1
		}
		delete(e.Relationships, otherID)
		L.Push(lua.LTrue)
		return 1
	}))

	entTbl.RawSetString("num_relationships", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(len(e.Relationships)))
		return 1
	}))

	entTbl.RawSetString("is_related", L.NewFunction(func(L *lua.LState) int {
		otherID := L.ToString(1)
		if otherID == "" {
			L.Push(lua.LFalse)
			return 1
		}
		rel, ok := e.GetRelationship(otherID)
		if !ok {
			L.Push(lua.LFalse)
			return 1
		}
		// Check if the relationship is a family bond (parent, child, mate, sibling)
		family := rel.Type == entity.RelationshipParent ||
			rel.Type == entity.RelationshipChild ||
			rel.Type == entity.RelationshipMate ||
			rel.Type == entity.RelationshipSibling
		L.Push(lua.LBool(family))
		return 1
	}))

	// Faction and profession setters
	entTbl.RawSetString("set_faction", L.NewFunction(func(L *lua.LState) int {
		newFaction := L.ToString(1)
		if newFaction == "" {
			L.Push(lua.LFalse)
			return 1
		}
		e.Faction = newFaction
		L.Push(lua.LTrue)
		return 1
	}))

	entTbl.RawSetString("set_profession", L.NewFunction(func(L *lua.LState) int {
		newProfession := L.ToString(1)
		e.Profession = newProfession
		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("self", entTbl)
}

func computeHunger(e *entity.Entity, tick uint64) float64 {
	species, ok := species.GetByID(e.Species)
	if !ok {
		return 0
	}
	if !species.AutoFeed && e.LastMealTick > 0 {
		threshold := species.StarvationThreshold
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

type luaNearbyCombatSite struct {
	LocationID         string
	Hostiles           int
	Allies             int
	TotalFactions      int
	ControllingFaction string
	ControlStrength    int
}

func nearbyCombatSitesLua(w *world.World, em *entity.EntityManager, ent *entity.Entity) []luaNearbyCombatSite {
	if w == nil || em == nil || ent == nil {
		return nil
	}
	adjacent := w.AdjacentLocations(ent.LocationID)
	sites := make([]luaNearbyCombatSite, 0, len(adjacent))
	for _, loc := range adjacent {
		if loc == nil {
			continue
		}
		site := luaNearbyCombatSite{
			LocationID:         loc.ID,
			ControllingFaction: loc.ControllingFaction,
			ControlStrength:    loc.ControlStrength,
		}
		factions := make(map[string]struct{})
		for _, other := range em.ByLocation(loc.ID) {
			if other == nil || other.ID == ent.ID || !other.Alive || other.Immortal {
				continue
			}
			factions[other.Faction] = struct{}{}
			fac1, ok := faction.GetFactionByID(ent.Faction)
			if !ok {
				continue
			}
			fac2, ok := faction.GetFactionByID(other.Faction)
			if !ok {
				continue
			}
			fac1h := fac1.Relation.GetFactionRelation(other.Faction)
			fac2h := fac2.Relation.GetFactionRelation(ent.Faction)
			combined := fac1h + fac2h
			if combined > 5 {
				site.Hostiles++
			} else {
				site.Allies++
			}
		}
		site.TotalFactions = len(factions)
		if site.Hostiles > 0 && site.TotalFactions >= 2 {
			sites = append(sites, site)
		}
	}
	return sites
}

func choosePassiveCombatDestinationLua(w *world.World, ent *entity.Entity, sites []luaNearbyCombatSite) string {
	if w == nil || ent == nil {
		return ""
	}
	current := w.Location(ent.LocationID)
	bestID := ""
	bestScore := -1 << 30
	consider := func(loc *world.Location) {
		if loc == nil || loc.ID == ent.LocationID {
			return
		}
		score := 0
		siteFound := false
		for _, site := range sites {
			if site.LocationID != loc.ID {
				continue
			}
			siteFound = true
			score -= site.Hostiles * 8
			score += site.Allies * 4
			if site.ControllingFaction == ent.Faction {
				score += 8
			} else if site.ControllingFaction != "" {
				score -= 8
			}
			break
		}
		if !siteFound {
			score += 12
		}
		if loc.ControllingFaction == ent.Faction {
			score += 10
		} else if loc.ControllingFaction != "" {
			score -= 10
		}
		if current != nil && loc.ParentID == current.ID {
			score += 2
		}
		if home := ent.AI.HomeLocation; home != "" && loc.ID == home {
			score += 15
		}
		if score > bestScore {
			bestScore = score
			bestID = loc.ID
		}
	}
	if home := ent.AI.HomeLocation; home != "" {
		consider(w.Location(home))
	}
	if current != nil && current.ParentID != "" {
		consider(w.Location(current.ParentID))
	}
	for _, loc := range w.AdjacentLocations(ent.LocationID) {
		consider(loc)
	}
	if bestScore <= 0 {
		return ""
	}
	return bestID
}

func shouldAssistNearbyCombatLua(rng *rand.Rand, ent *entity.Entity, site luaNearbyCombatSite) bool {
	if rng == nil || ent == nil || site.Hostiles == 0 || site.Allies == 0 {
		return false
	}
	chance := 8 + site.Allies*8
	if ent.AI.Brave {
		chance += 18
	}
	if site.ControllingFaction == ent.Faction {
		chance += 12
	}
	if site.Allies >= site.Hostiles {
		chance += 10
	}
	if chance > 60 {
		chance = 60
	}
	return rng.Intn(chance) == 0
}

func passiveCombatResponseLua(w *world.World, em *entity.EntityManager, ent *entity.Entity, rng *rand.Rand) bool {
	if w == nil || em == nil || ent == nil {
		return false
	}
	sites := nearbyCombatSitesLua(w, em, ent)
	if len(sites) == 0 {
		return false
	}
	current := w.Location(ent.LocationID)
	lostLocation := current != nil && current.ControllingFaction != "" && current.ControllingFaction != ent.Faction
	for _, site := range sites {
		if shouldAssistNearbyCombatLua(rng, ent, site) {
			if MoveRequest != nil {
				MoveRequest(ent, site.LocationID)
			} else {
				ent.LocationID = site.LocationID
			}
			return true
		}
	}
	if lostLocation || len(sites) > 0 {
		if destID := choosePassiveCombatDestinationLua(w, ent, sites); destID != "" {
			if MoveRequest != nil {
				MoveRequest(ent, destID)
			} else {
				ent.LocationID = destID
			}
			return true
		}
		return false
	}
	return false
}

func retreatCatchChanceLua(attacker, defender *entity.Entity) int {
	if attacker == nil || defender == nil {
		return 0
	}
	chance := 15 + (combatPowerScoreLua(attacker)-combatPowerScoreLua(defender))/2
	if attacker.AI.Brave {
		chance += 10
	}
	if defender.AI.Brave {
		chance -= 10
	}
	if defender.MaxHP > 0 && defender.HP*100 <= defender.MaxHP*35 {
		chance += 10
	}
	if chance < 0 {
		return 0
	}
	if chance > 80 {
		return 80
	}
	return chance
}

func scriptCombatAttack(w *world.World, em *entity.EntityManager, qm *quest.Manager, rng *rand.Rand, tick uint64, attacker, target *entity.Entity) bool {
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
	if attacker.Faction != "" && attacker.Faction != "civilian" && attacker.Faction != "deity" &&
		target.Faction != "" && target.Faction != "civilian" && target.Faction != "deity" {
		if loc := w.Location(target.LocationID); loc != nil {
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
	}
	if hit {
		target.ChangeEntityRelation(attacker.ID, -2)
		target.ChangeFactionRelation(attacker.Faction, -1)
		if aFac, ok := faction.GetFactionByID(attacker.Faction); ok {
			aFac.Relation.ChangeFactionRelation(target.Faction, -1)
		}
		if tFac, ok := faction.GetFactionByID(target.Faction); ok {
			tFac.Relation.ChangeFactionRelation(attacker.Faction, -1)
		}
	}
	if !target.Alive {
		combat.LootCorpse(attacker, target)
		if attacker.Faction != target.Faction {
			target.ChangeEntityRelation(attacker.ID, -10)
			target.ChangeFactionRelation(attacker.Faction, -7)
			if aFac, ok := faction.GetFactionByID(attacker.Faction); ok {
				aFac.Relation.ChangeFactionRelation(target.Faction, -7)
			}
			if tFac, ok := faction.GetFactionByID(target.Faction); ok {
				tFac.Relation.ChangeFactionRelation(attacker.Faction, -7)
			}
		}
		species, ok := species.GetByID(attacker.Species)
		if !ok {
			return false
		}
		if species.CanLevelUp {
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

func scriptRetreatOpportunityAttack(w *world.World, em *entity.EntityManager, qm *quest.Manager, rng *rand.Rand, tick uint64, attacker, defender *entity.Entity) bool {
	if w == nil || attacker == nil || defender == nil || !attacker.Alive || !defender.Alive {
		return false
	}
	chance := retreatCatchChanceLua(attacker, defender)
	if chance <= 0 || rng.Intn(100) >= chance {
		return false
	}
	return scriptCombatAttack(w, em, qm, rng, tick, attacker, defender)
}

func scriptFleeCombat(w *world.World, em *entity.EntityManager, qm *quest.Manager, rng *rand.Rand, tick uint64, ent *entity.Entity, hostiles []*entity.Entity) bool {
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
		if ent.LocationID == destID {
			scriptRetreatOpportunityAttack(w, em, qm, rng, tick, attacker, ent)
		}
		return true
	}
	ent.LocationID = destID
	scriptRetreatOpportunityAttack(w, em, qm, rng, tick, attacker, ent)
	return true
}

func questKilledLua(qm *quest.Manager, em *entity.EntityManager, target *entity.Entity) {
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

func bindWorld(L *lua.LState, w *world.World, em *entity.EntityManager, tm *world.GameTime, rng *rand.Rand, ent *entity.Entity, qm *quest.Manager) {
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
		if w.Location(id) == nil || ((w.IsDivineRealm(ent.LocationID) || w.IsDivineRealm(id)) && ent.Species != "deity" && ent.Faction != "deity") {
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
		tbl.RawSetString("profession", lua.LString(e.Profession))
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
		att := em.Get(a)
		def := em.Get(b)

		// If both resolve as entities, use entity-level relation.
		if att != nil && def != nil {
			rel := att.Relation.Relation(def)
			L.Push(lua.LBool(rel < 0))
			return 1
		}
		// Otherwise treat as faction strings and use faction-level hostility.
		L.Push(lua.LBool(IsHostile(a, b)))
		return 1
	}))

	worldTbl.RawSetString("get_faction", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		e := em.Get(id)
		if e == nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(e.Faction))
		return 1
	}))

	worldTbl.RawSetString("set_faction", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		newFaction := L.ToString(2)
		e := em.Get(id)
		if e == nil || newFaction == "" {
			L.Push(lua.LFalse)
			return 1
		}
		e.Faction = newFaction
		log.Printf("[lua] %s set faction of %s to %s", ent.Name, id, newFaction)
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("get_profession", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		e := em.Get(id)
		if e == nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(e.Profession))
		return 1
	}))

	worldTbl.RawSetString("set_profession", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		newProfession := L.ToString(2)
		e := em.Get(id)
		if e == nil {
			L.Push(lua.LFalse)
			return 1
		}
		e.Profession = newProfession
		log.Printf("[lua] %s set profession of %s to %s", ent.Name, id, newProfession)
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("get_relation", L.NewFunction(func(L *lua.LState) int {
		factionA := L.ToString(1)
		factionB := L.ToString(2)
		rel := factionHostility(factionA, factionB)
		var relStr string
		if rel < 0 {
			relStr = "hostile"
		} else if rel > 0 {
			relStr = "friendly"
		} else {
			relStr = "neutral"
		}
		L.Push(lua.LString(relStr))
		return 1
	}))

	worldTbl.RawSetString("set_relation", L.NewFunction(func(L *lua.LState) int {
		a := L.ToString(1)
		b := L.ToString(2)
		rel := L.ToString(3)
		var r relation.HostilityRelation
		switch rel {
		case "hostile":
			r = -5
		case "friendly":
			r = 5
		default:
			r = 0
		}
		setFactionHostility(a, b, r)
		log.Printf("[lua] %s set relation %s <-> %s = %s", ent.Name, a, b, rel)
		return 0
	}))

	worldTbl.RawSetString("defend_self", L.NewFunction(func(L *lua.LState) int {
		nearby := em.ByLocation(ent.LocationID)
		hostiles := make([]*entity.Entity, 0, len(nearby))
		for _, other := range nearby {
			if other == nil || other.ID == ent.ID || !other.Alive || other.Immortal {
				continue
			}
			if ent.Relation.Relation(other) >= 0 {
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

	worldTbl.RawSetString("avoid_combat", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LBool(passiveCombatResponseLua(w, em, ent, rng)))
		return 1
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

	worldTbl.RawSetString("loot_item", L.NewFunction(func(L *lua.LState) int {
		sourceID := L.ToString(1)
		itemDefID := L.ToString(2)
		source := em.Get(sourceID)
		if source == nil {
			L.Push(lua.LFalse)
			return 1
		}
		idx, found := economy.HasItem(source, itemDefID)
		if !found {
			L.Push(lua.LFalse)
			return 1
		}
		ok := economy.TransferItem(source, ent, idx)
		if ok {
			log.Printf("[lua] %s looted %s from %s", ent.Name, itemDefID, source.Name)
		}
		L.Push(lua.LBool(ok))
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
			if attacker.Relation.Relation(other) >= 0 {
				continue
			}
			other.TakeDamage(amount)
			count++
			if !other.Alive {
				combat.LootCorpse(attacker, other)
				if attacker.Faction != other.Faction {
					other.ChangeFactionRelation(attacker.Faction, -10)
				}
				species, ok := species.GetByID(attacker.Species)
				if !ok {
					continue
				}
				if species.CanLevelUp {
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

	// Relationship query functions on the world table
	worldTbl.RawSetString("get_relationship", L.NewFunction(func(L *lua.LState) int {
		entityID := L.ToString(1)
		otherID := L.ToString(2)
		e := em.Get(entityID)
		if e == nil {
			L.Push(lua.LNil)
			return 1
		}
		if otherID == "" {
			L.Push(lua.LNil)
			return 1
		}
		rel, ok := e.GetRelationship(otherID)
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		tbl.RawSetString("other_id", lua.LString(rel.OtherID))
		tbl.RawSetString("type", lua.LString(string(rel.Type)))
		tbl.RawSetString("since_tick", lua.LNumber(rel.SinceTick))
		L.Push(tbl)
		return 1
	}))

	worldTbl.RawSetString("get_children", L.NewFunction(func(L *lua.LState) int {
		entityID := L.ToString(1)
		e := em.Get(entityID)
		if e == nil {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		for _, rel := range e.GetChildren() {
			tbl.Append(lua.LString(rel.OtherID))
		}
		L.Push(tbl)
		return 1
	}))

	worldTbl.RawSetString("get_parents", L.NewFunction(func(L *lua.LState) int {
		entityID := L.ToString(1)
		e := em.Get(entityID)
		if e == nil {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		for _, rel := range e.GetParents() {
			tbl.Append(lua.LString(rel.OtherID))
		}
		L.Push(tbl)
		return 1
	}))

	worldTbl.RawSetString("get_partner", L.NewFunction(func(L *lua.LState) int {
		entityID := L.ToString(1)
		e := em.Get(entityID)
		if e == nil {
			L.Push(lua.LNil)
			return 1
		}
		rel, ok := e.GetPartner()
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(rel.OtherID))
		return 1
	}))

	worldTbl.RawSetString("get_relationship_type", L.NewFunction(func(L *lua.LState) int {
		entityID := L.ToString(1)
		otherID := L.ToString(2)
		e := em.Get(entityID)
		if e == nil {
			L.Push(lua.LNil)
			return 1
		}
		if otherID == "" {
			L.Push(lua.LNil)
			return 1
		}
		rel, ok := e.GetRelationship(otherID)
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(string(rel.Type)))
		return 1
	}))

	worldTbl.RawSetString("get_relationship_since", L.NewFunction(func(L *lua.LState) int {
		entityID := L.ToString(1)
		otherID := L.ToString(2)
		e := em.Get(entityID)
		if e == nil {
			L.Push(lua.LNil)
			return 1
		}
		if otherID == "" {
			L.Push(lua.LNil)
			return 1
		}
		rel, ok := e.GetRelationship(otherID)
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LNumber(rel.SinceTick))
		return 1
	}))

	worldTbl.RawSetString("has_relationship", L.NewFunction(func(L *lua.LState) int {
		entityID := L.ToString(1)
		otherID := L.ToString(2)
		e := em.Get(entityID)
		if e == nil {
			L.Push(lua.LFalse)
			return 1
		}
		if otherID == "" {
			L.Push(lua.LFalse)
			return 1
		}
		_, ok := e.GetRelationship(otherID)
		L.Push(lua.LBool(ok))
		return 1
	}))

	worldTbl.RawSetString("has_relationship_type", L.NewFunction(func(L *lua.LState) int {
		entityID := L.ToString(1)
		otherID := L.ToString(2)
		relType := entity.RelationshipType(L.ToString(3))
		e := em.Get(entityID)
		if e == nil {
			L.Push(lua.LFalse)
			return 1
		}
		if otherID == "" || relType == "" {
			L.Push(lua.LFalse)
			return 1
		}
		rel, ok := e.GetRelationship(otherID)
		if !ok {
			L.Push(lua.LFalse)
			return 1
		}
		L.Push(lua.LBool(rel.Type == relType))
		return 1
	}))

	worldTbl.RawSetString("add_relationship", L.NewFunction(func(L *lua.LState) int {
		entityID := L.ToString(1)
		otherID := L.ToString(2)
		relTypeStr := L.ToString(3)
		sinceTick := uint64(L.ToInt(4))
		e := em.Get(entityID)
		if e == nil || otherID == "" || relTypeStr == "" {
			L.Push(lua.LFalse)
			return 1
		}
		relType := entity.RelationshipType(relTypeStr)
		e.AddRelationship(otherID, relType, sinceTick)
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("remove_relationship", L.NewFunction(func(L *lua.LState) int {
		entityID := L.ToString(1)
		otherID := L.ToString(2)
		e := em.Get(entityID)
		if e == nil || otherID == "" {
			L.Push(lua.LFalse)
			return 1
		}
		delete(e.Relationships, otherID)
		L.Push(lua.LTrue)
		return 1
	}))

	worldTbl.RawSetString("num_relationships", L.NewFunction(func(L *lua.LState) int {
		entityID := L.ToString(1)
		e := em.Get(entityID)
		if e == nil {
			L.Push(lua.LNumber(0))
			return 1
		}
		L.Push(lua.LNumber(len(e.Relationships)))
		return 1
	}))

	worldTbl.RawSetString("is_related", L.NewFunction(func(L *lua.LState) int {
		entityID := L.ToString(1)
		otherID := L.ToString(2)
		e := em.Get(entityID)
		if e == nil || otherID == "" {
			L.Push(lua.LFalse)
			return 1
		}
		rel, ok := e.GetRelationship(otherID)
		if !ok {
			L.Push(lua.LFalse)
			return 1
		}
		family := rel.Type == entity.RelationshipParent ||
			rel.Type == entity.RelationshipChild ||
			rel.Type == entity.RelationshipMate ||
			rel.Type == entity.RelationshipSibling
		L.Push(lua.LBool(family))
		return 1
	}))
	// Drop this directly into your existing bindWorld(L, w, em, tm, rng, ent, qm) function
	L.SetField(worldTbl, "damage_location", L.NewFunction(func(L *lua.LState) int {
		// 1. Parse arguments passed down from the Lua script execution layer
		attackerID := L.CheckString(1)
		rawAmount := L.CheckNumber(2)
		damageAmount := int(rawAmount)

		// 2. Fetch the current location of the executing entity
		// (Assuming your entity or manager has a clean spatial map registry)
		currentLocationID := ent.LocationID

		// 3. Scan all live entities in the exact same location node map sector
		// Use your em (*entity.EntityManager) to safely query the world registry
		allEntities := em.GetEntitiesInLocation(currentLocationID)

		// 4. Iterate over the group and apply damage vectors cleanly
		for _, target := range allEntities {
			// Guard: Do not let an explosion damage deceased entities or the attacker themselves!
			if target.ID == attackerID || !target.Alive {
				continue
			}

			// Subtract health points directly on the target struct model memory layer
			target.HP -= damageAmount

			// Handle sudden fatality checks
			if target.HP <= 0 {
				target.HP = 0
				target.Alive = false
				target.TimeOfDeath = uint64(tm.Tick) // Log history timestamps accurately

				// Set their mood modifier context to dead state
				target.Mood = "dead"

				// Optional: Trigger your event system return pipeline array
				// simEvents = append(simEvents, events.NewDeathEvent(target.ID, attackerID))
			}
		}

		return 0 // Returns nothing back to the Lua script execution stack
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
