package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"simuz/internal/engine"
	"simuz/internal/entity"
	"simuz/internal/gen"
	"simuz/internal/web"
	"simuz/internal/world"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Initializing Simuz Simulation Engine...")

	// 1. Setup global runtime foundations
	seed := "estonian_gloom_2026"
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 2. Instantiate your concrete managers natively
	w := world.NewWorld()
	em := entity.NewEntityManager()

	// 3. Instantiate the builder orchestrator
	builder := engine.NewWorldBuilder(seed, rng)

	// 4. Build the simulation context by feeding it the managers
	sim, err := builder.BootstrapWorld(w, em)
	if err != nil {
		log.Fatalf("Critical Engine Boot Failure: %v", err)
	}

	log.Println("Bootstrapping Complete. Engine running smoothly.")

	// 5. Generate test data for pregnancies and relationships
	gen.GeneratePregnancies(sim.Entities.All(), 5, sim.Tick)
	gen.GenerateRelationships(sim.Entities.All(), 3, sim.Tick)

	// 6. Start the tick loop in a goroutine
	go sim.Start()

	// 7. Start the web UI
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	web.SetupRoutes(router, sim)

	go func() {
		log.Printf("Web UI listening on http://localhost:8080")
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("Web server failed: %v", err)
		}
	}()

	// 8. Wait for interrupt signal to gracefully shut down
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down simulation...")
	sim.Stop()
}
