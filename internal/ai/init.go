// Package ai contains the AI runtime, script loading, and Lua-facing helpers for entities.
package ai

import (
	// embed allows embedding files from the filesystem into the Go binary at compile time.

	// fmt provides formatted I/O utilities for error message construction.
	"fmt"
	// log provides standard logging functionality for runtime diagnostics.
	"log"
	// os provides operating system functions such as file stat and reading.
	"os"
	// filepath provides utilities for manipulating file path strings.
	"path/filepath"
	// strings provides utilities for manipulating string values.
	"strings"

	// lua is the Gopher-Lua interpreter used to load and execute embedded Lua scripts.
	lua "github.com/yuin/gopher-lua"
)

// InitScripts loads all embedded AI Lua scripts from the scripts directory
// into the global script registry, making them available for runtime execution.
// It logs each loaded script's name and type for diagnostic purposes.
func InitScripts() {
	// Load embedded scripts from the internal/ai/scripts directory
	err := loadScriptsDir("internal/ai/scripts")
	log.Printf("Loaded embedded AI scripts")
	if err != nil {
		log.Fatalf("Failed to load embedded scripts: %v", err)
	}
	// Log each loaded script's name and its associated type
	for name := range globalScripts.scripts {
		log.Printf("%s - %s\n", name, globalScripts.scriptTypes[name])
	}
}

// loadScriptsDir walks the given root directory recursively, finds all .lua
// files, and loads each one into the global script registry. It returns an
// error if the directory cannot be accessed or a script fails to load.
func loadScriptsDir(root string) error {
	log.Printf("Loading AI scripts from directory: %s", root)
	// Ensure the directory exists before attempting to walk it
	_, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("stat directory %s: %w", root, err)
	}
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		log.Printf("Visiting path: %s", path)
		if err != nil {
			return fmt.Errorf("walk path %s: %w", path, err)
		}
		// Skip directories; only process files
		if info.IsDir() {
			return nil
		}
		// Skip non-Lua files
		if filepath.Ext(path) != ".lua" {
			return nil
		}

		// Extract the base filename without the .lua extension
		name := filepath.Base(path)
		log.Printf("Loading AI script: %s", name)
		name = strings.TrimSuffix(name, ".lua")
		log.Printf("Loading AI script: %s", name)
		// Determine the script type from the top-level directory in the path
		scriptType := filepath.SplitList(path)[0]
		// Read the raw Lua source bytes from disk
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read script %s: %w", name, err)
		}
		// Parse and store the script in the global registry
		return loadScript(name, scriptType, string(data))
	})
}

// loadScript compiles the given Lua source string, stores the resulting
// bytecode prototype in the global scripts registry under the provided name,
// and associates it with the given script type. It returns an error if the
// script fails to parse or load.
func loadScript(name, scriptType, source string) error {
	// Create a new Lua state for compiling the script
	L := lua.NewState()
	defer L.Close()

	// Load the Lua source string into a Lua function
	fn, err := L.LoadString(source)
	if err != nil {
		return fmt.Errorf("load script %s: %w", name, err)
	}

	// Extract the bytecode prototype from the loaded function for storage
	proto := fn.Proto

	// Store the compiled script and its type in the global registry
	globalScripts.scripts[name] = proto
	globalScripts.scriptTypes[name] = scriptType
	log.Printf("Loaded AI script: %s", name)
	return nil
}
