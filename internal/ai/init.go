// Package ai contains the AI runtime, script loading, and Lua-facing helpers for entities.
package ai

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

//go:embed scripts/**/*.lua
var scriptFS embed.FS

func InitScripts() {
	// Load embedded scripts
	err := loadScriptsDir("internal/ai/scripts")
	log.Printf("Loaded embedded AI scripts")
	if err != nil {
		log.Fatalf("Failed to load embedded scripts: %v", err)
	}
	for name := range globalScripts.scripts {
		log.Printf("%s - %s\n", name, globalScripts.scriptTypes[name])
	}
}
func loadScriptsDir(root string) error {
	log.Printf("Loading AI scripts from directory: %s", root)
	_, err := os.Stat(root) // Ensure the directory exists
	if err != nil {
		return fmt.Errorf("stat directory %s: %w", root, err)
	}
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		log.Printf("Visiting path: %s", path)
		if err != nil {
			return fmt.Errorf("walk path %s: %w", path, err)
		}
		if info.IsDir() {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".lua" {
			return nil
		}

		name := filepath.Base(path)
		log.Printf("Loading AI script: %s", name)
		name = strings.TrimSuffix(name, ".lua")
		log.Printf("Loading AI script: %s", name)
		scriptType := filepath.SplitList(path)[0]
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read script %s: %w", name, err)
		}
		return loadScript(name, scriptType, string(data))
	})
}
func loadScript(name, scriptType, source string) error {
	L := lua.NewState()
	defer L.Close()

	fn, err := L.LoadString(source)
	if err != nil {
		return fmt.Errorf("load script %s: %w", name, err)
	}

	proto := fn.Proto

	globalScripts.scripts[name] = proto
	globalScripts.scriptTypes[name] = scriptType
	log.Printf("Loaded AI script: %s", name)
	return nil
}
