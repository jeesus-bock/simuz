// Package entity defines the simulation entities, their attributes, and related behaviors.
package entity

import "simuz/internal/relation"

// CanMate checks whether two entities are compatible for reproduction.
func CanMate(a, b *Entity) bool {
	if a == nil || b == nil {
		return false
	}
	if a.Species != b.Species {
		return false
	}
	if a.Gender == b.Gender {
		return false
	}
	if !a.Alive || !b.Alive {
		return false
	}
	if a.Reproduction.Pregnant || b.Reproduction.Pregnant {
		return false
	}

	if a.Species == "vampire" {
		return false
	}
	return true
}

// SpawnBaby creates a new offspring entity from two parents.
// The caller is responsible for providing a unique ID for the baby.
func SpawnBaby(parent1, parent2 *Entity, id, babyName string, rng func(int) int) *Entity {
	if !CanMate(parent1, parent2) {
		return nil
	}

	attrs := inheritAttributes(parent1.Attributes, parent2.Attributes, rng)
	baby := NewEntity(
		id,
		babyName,
		parent1.Species,
		attrs,
		1,
		relation.Relation{},
	)

	// Inherit gender randomly from one of the parents
	if rng(2) == 0 {
		baby.Gender = parent1.Gender
	} else {
		baby.Gender = parent2.Gender
	}

	// Mark whichever parent can get pregnant
	if parent1.CanGetPregnant() {
		parent1.Reproduction.Pregnant = true
		parent1.Reproduction.FatherID = parent2.ID
	} else if parent2.CanGetPregnant() {
		parent2.Reproduction.Pregnant = true
		parent2.Reproduction.FatherID = parent1.ID
	}

	return baby
}

func inheritAttributes(a, b Attributes, rng func(int) int) Attributes {
	return Attributes{
		STR: clampAttr((a.STR+b.STR)/2 + rng(3) - 1),
		DEX: clampAttr((a.DEX+b.DEX)/2 + rng(3) - 1),
		CON: clampAttr((a.CON+b.CON)/2 + rng(3) - 1),
		INT: clampAttr((a.INT+b.INT)/2 + rng(3) - 1),
		WIS: clampAttr((a.WIS+b.WIS)/2 + rng(3) - 1),
		CHA: clampAttr((a.CHA+b.CHA)/2 + rng(3) - 1),
	}
}
