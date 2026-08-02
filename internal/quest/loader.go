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
