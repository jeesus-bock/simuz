// Package storage defines the persistence interfaces and SQLite-backed implementation used by the simulation.
package storage

import "simuz/internal/engine"

type Storage interface {
	Save(sim *engine.Simulation) error
	Load() (*engine.Simulation, error)
	Enabled() bool
}
