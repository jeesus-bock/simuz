// Package gen contains world generation helpers and seeded simulation setup utilities.
package gen

import "simuz/internal/quest"

// SeedQuests loads all quest definitions from embedded Lua scripts
// under internal/quest/scripts/*.lua. Prefer quest.LoadScripts() directly.
func SeedQuests() []*quest.QuestDef {
	return quest.MustLoadScripts()
}
