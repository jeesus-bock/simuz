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
	"simuz/internal/web"

	"github.com/gin-gonic/gin"
)

var Sim *engine.Simulation

func main() {
	seed := flag.String("seed", "default", "World generation seed")
	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

	ai.InitScripts()

	g := gen.New(*seed)
	w, entities := g.Generate()

	deities, _ := gen.GenerateDeities(w)

	Sim := engine.NewSimulation(w)
	for _, e := range entities {
		Sim.Entities.Add(e)
	}
	for _, d := range deities {
		Sim.Entities.Add(d)
	}
	// Quests are defined as Lua scripts in internal/quest/scripts/*.lua
	for _, q := range gen.SeedQuests() {
		Sim.Quests.Register(q)
	}

	go Sim.Start()

	router := gin.Default()

	v1 := router.Group("/api/v1")
	api.RegisterRoutes(v1, Sim)
	web.SetupRoutes(router, Sim)

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
	Sim.Stop()
}
