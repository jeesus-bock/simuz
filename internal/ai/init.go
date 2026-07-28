// Package ai contains the AI runtime, script loading, and Lua-facing helpers for entities.
package ai

import (
	"embed"
	"log"
	"path"
	"strings"
)

//go:embed scripts/**/*.lua
var scriptFS embed.FS

func InitScripts() {
	if err := loadScriptsDir("scripts"); err != nil {
		log.Printf("No embedded scripts directory: %v", err)
	}
}

func loadScriptsDir(dir string) error {
	entries, err := scriptFS.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			subDir := path.Join(dir, entry.Name())
			if err := loadScriptsDir(subDir); err != nil {
				log.Printf("Failed to read script dir %s: %v", subDir, err)
			}
			continue
		}
		name := entry.Name()
		if len(name) < 4 || !strings.HasSuffix(name, ".lua") {
			continue
		}
		filePath := path.Join(dir, name)
		data, err := scriptFS.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read script %s: %v", filePath, err)
			continue
		}
		relPath := strings.TrimPrefix(filePath, "scripts/")
		scriptName := relPath[:len(relPath)-4]
		if err := LoadScript(scriptName, string(data)); err != nil {
			log.Printf("Failed to load script %s: %v", filePath, err)
			continue
		}
	}
	return nil
}
