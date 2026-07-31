package web

import (
	"testing"

	"simuz/internal/engine"
	"simuz/internal/entity"
	"simuz/internal/events"
	"simuz/internal/relation"
)

func TestBuildRecentBirthsUsesSimulationEventHistory(t *testing.T) {
	sim := engine.NewSimulation(nil, entity.NewEntityManager())

	parent := entity.NewEntity("parent", "Parent", "human", entity.Attributes{}, 3, relation.CivilianRelation)
	parent.Gender = entity.GenderFemale
	parent.Faction = "civilian"
	sim.Entities.Add(parent)

	offspring := entity.NewEntity("child", "Child", "human", entity.Attributes{}, 1, relation.CivilianRelation)
	offspring.Gender = entity.GenderFemale
	offspring.Faction = "civilian"
	sim.Entities.Add(offspring)

	sim.Emit(events.SimEvent{
		Type:   events.EventEntityBorn,
		Tick:   42,
		Source: offspring.ID,
		Data: map[string]any{
			"mother":   parent.ID,
			"father":   "father-1",
			"species":  "human",
			"location": "town",
		},
	})

	births := buildRecentBirths(sim)
	if len(births) != 1 {
		t.Fatalf("expected 1 birth, got %d", len(births))
	}
	if births[0].OffspringName != "Child" {
		t.Fatalf("expected offspring name Child, got %q", births[0].OffspringName)
	}
	if births[0].ParentName != "Parent" {
		t.Fatalf("expected parent name Parent, got %q", births[0].ParentName)
	}
	if births[0].Tick != 42 {
		t.Fatalf("expected tick 42, got %d", births[0].Tick)
	}
}
