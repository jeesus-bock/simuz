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

func main() {
	// seed := flag.String("seed", "default", "World generation seed")
	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

	ai.InitScripts()

	g := gen.NewGenerator("puisto")
	w, entities := g.Generate()

	deities, _ := gen.GenerateDeities(w)

	engine.Sim = engine.NewSimulation(w)
	for _, e := range entities {
		engine.Sim.Entities.Add(e)
	}
	for _, d := range deities {
		engine.Sim.Entities.Add(d)
	}
	// Quests are defined as Lua scripts in internal/quest/scripts/*.lua
	for _, q := range gen.SeedQuests() {
		engine.Sim.Quests.Register(q)
	}

	go engine.Sim.Start()

	router := gin.Default()

	v1 := router.Group("/api/v1")
	api.RegisterRoutes(v1, engine.Sim)
	web.SetupRoutes(router, engine.Sim)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("engine.Simuz listening on :%s", *port)
		if err := router.Run(":" + *port); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down...")
	engine.Sim.Stop()
}
