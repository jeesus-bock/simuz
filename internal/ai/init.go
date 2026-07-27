package ai

import (
	"embed"
	"log"
)

//go:embed scripts/*.lua
var scriptFS embed.FS

func InitScripts() {
	entries, err := scriptFS.ReadDir("scripts")
	if err != nil {
		log.Printf("No embedded scripts directory: %v", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 4 || name[len(name)-4:] != ".lua" {
			continue
		}
		data, err := scriptFS.ReadFile("scripts/" + name)
		if err != nil {
			log.Printf("Failed to read script %s: %v", name, err)
			continue
		}
		scriptName := name[:len(name)-4]
		if err := LoadScript(scriptName, string(data)); err != nil {
			log.Printf("Failed to load script %s: %v", name, err)
		}
	}
}
