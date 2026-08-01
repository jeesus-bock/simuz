package gen

import (
	"simuz/internal/entity"
)

// GeneratePregnancies marks up to count female entities as pregnant for testing.
func GeneratePregnancies(entities []*entity.Entity, count int, tick uint64) {
	created := 0
	for _, e := range entities {
		if created >= count {
			break
		}
		if e.Gender != "female" {
			continue
		}
		e.Reproduction.Pregnant = true
		e.Reproduction.PregnantSinceTick = tick
		created++
	}
}

// GenerateRelationships creates spouse relationships between consecutive entity pairs for testing.
func GenerateRelationships(entities []*entity.Entity, count int, tick uint64) {
	created := 0
	for i := 0; i < len(entities)-1 && created < count; i += 2 {
		e := entities[i]
		other := entities[i+1]
		e.Relationships = append(e.Relationships, entity.EntityRelationship{
			OtherID:   other.ID,
			Type:      entity.RelationshipType("spouse"),
			SinceTick: tick,
		})
		other.Relationships = append(other.Relationships, entity.EntityRelationship{
			OtherID:   e.ID,
			Type:      entity.RelationshipType("spouse"),
			SinceTick: tick,
		})
		created++
	}
}
