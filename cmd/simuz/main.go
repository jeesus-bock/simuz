package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"simuz/internal/ai"
	"simuz/internal/api"
	"simuz/internal/engine"
	"simuz/internal/gen"
	"simuz/internal/storage"
	"simuz/internal/web"

	"github.com/gin-gonic/gin"
)

func main() {
	seed := flag.String("seed", "default", "World generation seed")
	dbPath := flag.String("db", "", "Path to SQLite database (empty = no persistence)")
	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

	ai.InitScripts()

	g := gen.New(*seed)
	w, entities := g.Generate()

	deities, _ := gen.GenerateDeities(w)

	sim := engine.NewSimulation(w)
	for _, e := range entities {
		sim.Entities.Add(e)
	}
	for _, d := range deities {
		sim.Entities.Add(d)
	}
	// Quests are defined as Lua scripts in internal/quest/scripts/*.lua
	for _, q := range gen.SeedQuests() {
		sim.Quests.Register(q)
	}

	if *dbPath != "" {
		store := storage.NewSQLiteStore(*dbPath)
		if err := store.Open(); err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer store.Close()
		sim.Storage = store
		log.Printf("Persistence enabled: %s", *dbPath)
	} else {
		log.Println("Running without persistence")
	}

	go sim.Start()

	router := gin.Default()

	v1 := router.Group("/api/v1")
	api.RegisterRoutes(v1, sim)
	web.SetupRoutes(router, sim)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Simuz listening on :%s", *port)
		if err := router.Run(":" + *port); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down...")
	sim.Stop()
}
