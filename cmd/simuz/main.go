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
	"simuz/internal/world"
)

func main() {
	log.Println("Initializing Simuz Simulation Engine...")

	// 1. Setup global runtime foundations
	seed := "estonian_gloom_2026"
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 2. Instantiate your concrete managers natively
	w := world.NewWorld()           // Use your actual constructor names here
	em := entity.NewEntityManager() // Use your actual constructor names here

	// 3. Instantiate the builder orchestrator
	builder := engine.NewWorldBuilder(seed, rng)

	// 4. Build the simulation context by feeding it the managers
	sim, err := builder.BootstrapWorld(w, em)
	if err != nil {
		log.Fatalf("Critical Engine Boot Failure: %v", err)
	}

	log.Println("Bootstrapping Complete. Engine running smoothly.")

	// 5. Start the tick loop in a goroutine
	go sim.Start()

	// 6. Wait for interrupt signal to gracefully shut down
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down simulation...")
	sim.Stop()
}
