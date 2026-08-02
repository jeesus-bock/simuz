// Package quest defines quest definitions, state handling, and quest progression logic.
package quest

import (
	"embed"
	"fmt"
	"log"
	"strconv"
	"strings"

	"simuz/internal/events"

	lua "github.com/yuin/gopher-lua"
)

//go:embed scripts/*.lua
var questScriptFS embed.FS

// LoadScripts loads every .lua file under scripts/ and returns registered quest defs
// and any simulation events the scripts produced. Each script must call quest.define({ ... })
// once (or more). Scripts may optionally return a table of events as their last expression.
func LoadScripts() ([]*QuestDef, []*events.SimEvent, error) {
	entries, err := questScriptFS.ReadDir("scripts")
	if err != nil {
		return nil, nil, fmt.Errorf("read quest scripts: %w", err)
	}

	var defs []*QuestDef
	var allEvents []*events.SimEvent
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".lua") {
			continue
		}
		data, err := questScriptFS.ReadFile("scripts/" + name)
		if err != nil {
			return nil, nil, fmt.Errorf("read quest script %s: %w", name, err)
		}
		loaded, evts, err := execQuestScript(name, string(data))
		if err != nil {
			return nil, nil, fmt.Errorf("quest script %s: %w", name, err)
		}
		defs = append(defs, loaded...)
		allEvents = append(allEvents, evts...)
		for _, d := range loaded {
			log.Printf("Loaded quest script: %s (%s)", name, d.ID)
		}
	}
	return defs, allEvents, nil
}

// MustLoadScripts is LoadScripts that panics on error (for startup).
func MustLoadScripts() []*QuestDef {
	defs, _, err := LoadScripts()
	if err != nil {
		panic(err)
	}
	return defs
}

// execQuestScript runs a single quest Lua script and returns the quest definitions
// and any simulation events the script produced.
func execQuestScript(name, source string) ([]*QuestDef, []*events.SimEvent, error) {
	L := lua.NewState()
	defer L.Close()

	var defined []*QuestDef

	questTbl := L.NewTable()
	questTbl.RawSetString("define", L.NewFunction(func(L *lua.LState) int {
		tbl := L.CheckTable(1)
		def, err := tableToQuestDef(tbl)
		if err != nil {
			L.RaiseError("quest.define: %s", err.Error())
			return 0
		}
		defined = append(defined, def)
		return 0
	}))
	L.SetGlobal("quest", questTbl)

	utilTbl := L.NewTable()
	utilTbl.RawSetString("event", L.NewFunction(func(L *lua.LState) int {
		eventType := L.ToString(1)
		tbl := L.NewTable()
		tbl.RawSetString("type", lua.LString(eventType))
		if L.GetTop() >= 2 {
			if data, ok := L.Get(2).(*lua.LTable); ok {
				tbl.RawSetString("data", data)
			}
		}
		L.Push(tbl)
		return 1
	}))
	L.SetGlobal("util", utilTbl)

	proto, err := L.LoadString(source)
	if err != nil {
		return nil, nil, err
	}
	err = L.CallByParam(lua.P{
		Fn:      proto,
		NRet:    1,
		Protect: true,
	})
	if err != nil {
		return nil, nil, err
	}

	var simEvents []*events.SimEvent
	val := L.Get(-1)
	if tbl, ok := val.(*lua.LTable); ok {
		simEvents = decodeQuestEvents(tbl)
	}

	if len(defined) == 0 {
		return nil, nil, fmt.Errorf("no quest.define() call in %s", name)
	}
	return defined, simEvents, nil
}

var questEventTypeNames = map[string]events.EventType{
	"quest_accepted":  events.EventTypeQuestAccept,
	"quest_completed": events.EventTypeQuestComplete,
	"quest_progress":  events.EventTypeQuestProgress,
	"spawn_creature":  events.EventEntityBorn,
	"world":           events.EventWorld,
	"ambient":         events.EventAmbient,
	"mood":            events.EventMood,
	"divine":          events.EventDivine,
	"combat":          events.EventEntityKilled,
	"item_collected":  events.EventItemCollected,
	"item_delivered":  events.EventItemDelivered,
	"location_entered": events.EventLocationEntered,
	"entity_talked":   events.EventEntityTalked,
	"travel_completed": events.EventTravelCompleted,
	"xp_gained":       events.EventXPGained,
}

func decodeQuestEvents(tbl *lua.LTable) []*events.SimEvent {
	var luaEvents []*events.SimEvent
	tbl.ForEach(func(k, v lua.LValue) {
		eventTbl, ok := v.(*lua.LTable)
		if !ok {
			return
		}
		ev := &events.SimEvent{
			Tick: 0,
		}

		if typeVal := eventTbl.RawGetString("type"); typeVal != lua.LNil {
			typeStr := typeVal.String()
			if named, ok := questEventTypeNames[typeStr]; ok {
				ev.Type = named
			} else if typeInt, err := strconv.Atoi(typeStr); err == nil {
				ev.Type = events.EventType(typeInt)
			}
		}

		if srcVal := eventTbl.RawGetString("source"); srcVal != lua.LNil {
			ev.Source = srcVal.String()
		}

		if dataVal := eventTbl.RawGet(lua.LString("data")); dataVal != lua.LNil {
			result := make(map[string]any)
			table, ok := dataVal.(*lua.LTable)
			if ok {
				table.ForEach(func(key lua.LValue, val lua.LValue) {
					if strKey, ok := key.(lua.LString); ok {
						result[string(strKey)] = goValue(val)
					}
				})
			}
			ev.Data = result
		}

		luaEvents = append(luaEvents, ev)
	})
	return luaEvents
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
		return nil
	}
}

func tableToQuestDef(tbl *lua.LTable) (*QuestDef, error) {
	id := luaString(tbl, "id")
	if id == "" {
		return nil, fmt.Errorf("missing id")
	}
	title := luaString(tbl, "title")
	if title == "" {
		return nil, fmt.Errorf("missing title")
	}

	def := &QuestDef{
		ID:          id,
		Title:       title,
		Type:        QuestType(luaStringDefault(tbl, "type", "side")),
		Level:       luaInt(tbl, "level"),
		Description: luaString(tbl, "description"),
	}

	if src := luaTable(tbl, "source"); src != nil {
		def.Source = &QuestSource{
			Type:       luaString(src, "type"),
			NPCID:      luaStringFirst(src, "npc_id", "npc"),
			LocationID: luaStringFirst(src, "location_id", "location"),
		}
		if dlg := luaTable(src, "dialog"); dlg != nil {
			def.Source.Dialog = &QuestDialog{
				Accept:   luaString(dlg, "accept"),
				Progress: luaString(dlg, "progress"),
				Complete: luaString(dlg, "complete"),
			}
		}
	}

	if prereq := luaTable(tbl, "prerequisites"); prereq != nil {
		def.Prereqs = &Prerequisites{
			QuestsCompleted: luaStringSlice(prereq, "quests_completed"),
			QuestsActive:    luaStringSlice(prereq, "quests_active"),
			LevelMin:        luaInt(prereq, "level_min"),
			LevelMax:        luaInt(prereq, "level_max"),
		}
	}

	stagesTbl := luaTable(tbl, "stages")
	if stagesTbl == nil {
		return nil, fmt.Errorf("missing stages")
	}
	stagesTbl.ForEach(func(_, v lua.LValue) {
		st, ok := v.(*lua.LTable)
		if !ok {
			return
		}
		stage := StageDef{
			ID:           luaString(st, "id"),
			Name:         luaString(st, "name"),
			Description:  luaString(st, "description"),
			Requirements: luaStringSlice(st, "requirements"),
		}
		if objs := luaTable(st, "objectives"); objs != nil {
			objs.ForEach(func(_, ov lua.LValue) {
				ot, ok := ov.(*lua.LTable)
				if !ok {
					return
				}
				obj := ObjectiveDef{
					ID:             luaString(ot, "id"),
					Type:           luaString(ot, "type"),
					Description:    luaString(ot, "description"),
					Optional:       luaBool(ot, "optional"),
					Count:          luaInt(ot, "count"),
					EntityTemplate: luaStringFirst(ot, "entity_template", "entity"),
					LocationID:     luaStringFirst(ot, "location_id", "location"),
					NPCID:          luaStringFirst(ot, "npc_id", "npc"),
					ItemTemplate:   luaStringFirst(ot, "item_template", "item"),
				}
				// Binary objectives (talk/visit/deliver) default to count 1 so they
				// don't auto-complete when Count is omitted.
				if obj.Count == 0 {
					switch obj.Type {
					case "talk_to_npc", "visit_location", "deliver_item":
						obj.Count = 1
					case "kill_entities", "collect_items":
						obj.Count = 1
					}
				}
				stage.Objectives = append(stage.Objectives, obj)
			})
		}
		def.Stages = append(def.Stages, stage)
	})
	if len(def.Stages) == 0 {
		return nil, fmt.Errorf("stages empty")
	}

	if rew := luaTable(tbl, "rewards"); rew != nil {
		def.Rewards = &Rewards{
			Experience: luaInt(rew, "experience"),
			Gold:       luaInt(rew, "gold"),
		}
		if items := luaTable(rew, "items"); items != nil {
			items.ForEach(func(_, v lua.LValue) {
				it, ok := v.(*lua.LTable)
				if !ok {
					return
				}
				count := luaInt(it, "count")
				if count == 0 {
					count = 1
				}
				def.Rewards.Items = append(def.Rewards.Items, RewardItem{
					Template: luaStringFirst(it, "template", "item", "id"),
					Count:    count,
				})
			})
		}
		if unlocks := luaTable(rew, "unlocks"); unlocks != nil {
			def.Rewards.Unlocks = &Unlocks{
				Quests:    luaStringSlice(unlocks, "quests"),
				Locations: luaStringSlice(unlocks, "locations"),
				Recipes:   luaStringSlice(unlocks, "recipes"),
			}
		}
	}

	if fails := luaTable(tbl, "failure_conditions"); fails != nil {
		fails.ForEach(func(_, v lua.LValue) {
			ft, ok := v.(*lua.LTable)
			if !ok {
				return
			}
			def.FailConditions = append(def.FailConditions, FailCondition{
				Type:     luaString(ft, "type"),
				Hours:    luaInt(ft, "hours"),
				EntityID: luaStringFirst(ft, "entity_id", "entity"),
				Flag:     luaString(ft, "flag"),
			})
		})
	}

	return def, nil
}

func luaString(tbl *lua.LTable, key string) string {
	v := tbl.RawGetString(key)
	if s, ok := v.(lua.LString); ok {
		return string(s)
	}
	return ""
}

func luaStringDefault(tbl *lua.LTable, key, def string) string {
	s := luaString(tbl, key)
	if s == "" {
		return def
	}
	return s
}

func luaStringFirst(tbl *lua.LTable, keys ...string) string {
	for _, k := range keys {
		if s := luaString(tbl, k); s != "" {
			return s
		}
	}
	return ""
}

func luaInt(tbl *lua.LTable, key string) int {
	v := tbl.RawGetString(key)
	switch n := v.(type) {
	case lua.LNumber:
		return int(n)
	}
	return 0
}

func luaBool(tbl *lua.LTable, key string) bool {
	v := tbl.RawGetString(key)
	if b, ok := v.(lua.LBool); ok {
		return bool(b)
	}
	return false
}

func luaTable(tbl *lua.LTable, key string) *lua.LTable {
	v := tbl.RawGetString(key)
	if t, ok := v.(*lua.LTable); ok {
		return t
	}
	return nil
}

func luaStringSlice(tbl *lua.LTable, key string) []string {
	t := luaTable(tbl, key)
	if t == nil {
		return nil
	}
	var out []string
	t.ForEach(func(_, v lua.LValue) {
		if s, ok := v.(lua.LString); ok {
			out = append(out, string(s))
		}
	})
	return out
}
