// Package quest defines quest definitions, state handling, and quest progression logic.
package quest

import (
	"embed"
	"fmt"
	"log"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

//go:embed scripts/*.lua
var questScriptFS embed.FS

// LoadScripts loads every .lua file under scripts/ and returns registered quest defs.
// Each script must call quest.define({ ... }) once (or more).
func LoadScripts() ([]*QuestDef, error) {
	entries, err := questScriptFS.ReadDir("scripts")
	if err != nil {
		return nil, fmt.Errorf("read quest scripts: %w", err)
	}

	var defs []*QuestDef
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
			return nil, fmt.Errorf("read quest script %s: %w", name, err)
		}
		loaded, err := execQuestScript(name, string(data))
		if err != nil {
			return nil, fmt.Errorf("quest script %s: %w", name, err)
		}
		defs = append(defs, loaded...)
		for _, d := range loaded {
			log.Printf("Loaded quest script: %s (%s)", name, d.ID)
		}
	}
	return defs, nil
}

// MustLoadScripts is LoadScripts that panics on error (for startup).
func MustLoadScripts() []*QuestDef {
	defs, err := LoadScripts()
	if err != nil {
		panic(err)
	}
	return defs
}

// execQuestScript runs a single quest Lua script and returns the quest definitions.
func execQuestScript(name, source string) ([]*QuestDef, error) {
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

	proto, err := L.LoadString(source)
	if err != nil {
		return nil, err
	}
	err = L.CallByParam(lua.P{
		Fn:      proto,
		NRet:    0,
		Protect: true,
	})
	if err != nil {
		return nil, err
	}

	if len(defined) == 0 {
		return nil, fmt.Errorf("no quest.define() call in %s", name)
	}
	return defined, nil
}

func tableToQuestDef(tbl *lua.LTable) (*QuestDef, error) {
	def := &QuestDef{}

	def.ID = luaString(tbl, "id")
	def.Title = luaString(tbl, "title")
	def.Type = QuestType(luaString(tbl, "type"))
	def.Level = luaInt(tbl, "level")
	def.Description = luaString(tbl, "description")

	if src := luaTable(tbl, "source"); src != nil {
		def.Source = &QuestSource{
			Type:       luaString(src, "type"),
			NPCID:      luaString(src, "npc_id"),
			LocationID: luaString(src, "location_id"),
		}
		if dlg := luaTable(src, "dialog"); dlg != nil {
			def.Source.Dialog = &QuestDialog{
				Accept:   luaString(dlg, "accept"),
				Progress: luaString(dlg, "progress"),
				Complete: luaString(dlg, "complete"),
			}
		}
	}

	if prereqs := luaTable(tbl, "prerequisites"); prereqs != nil {
		def.Prereqs = &Prerequisites{
			QuestsCompleted: luaStringSlice(prereqs, "quests_completed"),
			QuestsActive:    luaStringSlice(prereqs, "quests_active"),
			LevelMin:        luaInt(prereqs, "level_min"),
			LevelMax:        luaInt(prereqs, "level_max"),
			FactionRep:      luaStringIntMap(prereqs, "faction_reputation"),
		}
		if flagsTbl := luaTable(prereqs, "flags"); flagsTbl != nil {
			flagsTbl.ForEach(func(_, v lua.LValue) {
				if flagTbl, ok := v.(*lua.LTable); ok {
					fc := FlagCondition{
						Flag: luaString(flagTbl, "flag"),
					}
					if valTbl := luaTable(flagTbl, "value"); valTbl != nil {
						// value can be any lua type; store as string for simplicity
						fc.Value = luaLValueToString(valTbl)
					}
					def.Prereqs.Flags = append(def.Prereqs.Flags, fc)
				}
			})
		}
	}

	if stagesTbl := luaTable(tbl, "stages"); stagesTbl != nil {
		stagesTbl.ForEach(func(_, v lua.LValue) {
			if stageTbl, ok := v.(*lua.LTable); ok {
				stage := StageDef{
					ID:          luaString(stageTbl, "id"),
					Name:        luaString(stageTbl, "name"),
					Description: luaString(stageTbl, "description"),
					Requirements: luaStringSlice(stageTbl, "requirements"),
				}
				if objsTbl := luaTable(stageTbl, "objectives"); objsTbl != nil {
					objsTbl.ForEach(func(_, ov lua.LValue) {
						if objTbl, ok := ov.(*lua.LTable); ok {
							stage.Objectives = append(stage.Objectives, ObjectiveDef{
								ID:             luaString(objTbl, "id"),
								Type:           luaString(objTbl, "type"),
								Description:    luaString(objTbl, "description"),
								Optional:       luaBool(objTbl, "optional"),
								Count:          luaInt(objTbl, "count"),
								EntityTemplate: luaString(objTbl, "entity_template"),
								LocationID:     luaString(objTbl, "location_id"),
								NPCID:          luaString(objTbl, "npc_id"),
								ItemTemplate:   luaString(objTbl, "item_template"),
							})
						}
					})
				}
				def.Stages = append(def.Stages, stage)
			}
		})
	}

	if rewardsTbl := luaTable(tbl, "rewards"); rewardsTbl != nil {
		def.Rewards = &Rewards{
			Experience: luaInt(rewardsTbl, "experience"),
			Gold:       luaInt(rewardsTbl, "gold"),
		}
		if itemsTbl := luaTable(rewardsTbl, "items"); itemsTbl != nil {
			itemsTbl.ForEach(func(_, v lua.LValue) {
				if itemTbl, ok := v.(*lua.LTable); ok {
					def.Rewards.Items = append(def.Rewards.Items, RewardItem{
						Template: luaString(itemTbl, "template"),
						Count:    luaInt(itemTbl, "count"),
					})
				}
			})
		}
		def.Rewards.FactionRep = luaStringIntMap(rewardsTbl, "faction_reputation")
		if unlocksTbl := luaTable(rewardsTbl, "unlocks"); unlocksTbl != nil {
			def.Rewards.Unlocks = &Unlocks{
				Quests:    luaStringSlice(unlocksTbl, "quests"),
				Locations: luaStringSlice(unlocksTbl, "locations"),
				Recipes:   luaStringSlice(unlocksTbl, "recipes"),
			}
		}
	}

	if fcTbl := luaTable(tbl, "failure_conditions"); fcTbl != nil {
		fcTbl.ForEach(func(_, v lua.LValue) {
			if fcItem, ok := v.(*lua.LTable); ok {
				def.FailConditions = append(def.FailConditions, FailCondition{
					Type:     luaString(fcItem, "type"),
					Hours:    luaInt(fcItem, "hours"),
					EntityID: luaString(fcItem, "entity_id"),
					Flag:     luaString(fcItem, "flag"),
				})
			}
		})
	}

	return def, nil
}

func luaStringIntMap(tbl *lua.LTable, key string) map[string]int {
	sub := luaTable(tbl, key)
	if sub == nil {
		return nil
	}
	m := make(map[string]int)
	sub.ForEach(func(k, v lua.LValue) {
		if ks, ok := k.(lua.LString); ok {
			switch n := v.(type) {
			case lua.LNumber:
				m[string(ks)] = int(n)
			}
		}
	})
	return m
}

func luaLValueToString(v lua.LValue) string {
	switch val := v.(type) {
	case lua.LString:
		return string(val)
	case lua.LNumber:
		return val.String()
	case lua.LBool:
		return val.String()
	default:
		return val.String()
	}
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
