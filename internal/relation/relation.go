// Package relation contains all the relatin code
package relation

type HostilityRelation int

type SpeciesRelation map[string]HostilityRelation

type FactionRelation map[string]HostilityRelation

type ProfessionRelation map[string]HostilityRelation

type EntityRelation map[string]HostilityRelation

func (r HostilityRelation) Int() int {
	return int(r)
}
func (r HostilityRelation) String() string {
	if r > 5 {
		return "friendly"
	}
	if r < -5 {
		return "hostile"
	}
	return "neutral"
}

type Relation struct {
	SpeciesRelation    SpeciesRelation    `json:"speciesRelation,omitempty"`
	FactionRelation    FactionRelation    `json:"factionRelation,omitempty"`
	ProfessionRelation ProfessionRelation `json:"professionRelation,omitempty"`
	EntityRelation     EntityRelation     `json:"entityRelation,omitempty"`
}
type EntityLookup interface {
	GetID() string
	GetFaction() string
	GetSpecies() string
	GetProfession() string
}

// Relation calculates the combined hostility relation based on species, faction, profession, and entity relations.
func (f *Relation) Relation(ent EntityLookup) HostilityRelation {
	combined := 0
	combined += f.GetEntityRelation(ent.GetID()).Int()
	combined += f.GetFactionRelation(ent.GetFaction()).Int()
	combined += f.GetSpeciesRelation(ent.GetSpecies()).Int()
	combined += f.GetProfessionRelation(ent.GetProfession()).Int()
	return HostilityRelation(combined)
}

type FactionLookup interface {
	GetID() string
	FactionExists(id string) bool
	GetName(id string) string
}

func (f *Relation) FactinRelation(fac FactionLookup) HostilityRelation {
	combined := 0
	combined += f.GetFactionRelation(fac.GetID()).Int()

	return HostilityRelation(combined)
}

// SetSpeciesRelation sets the hostility relation for a specific species.
func (f *Relation) SetSpeciesRelation(species string, relation HostilityRelation) {
	if f.SpeciesRelation == nil {
		f.SpeciesRelation = make(SpeciesRelation)
	}
	f.SpeciesRelation[species] = relation
}

// GetSpeciesRelation retrieves the hostility relation for a specific species.
func (f *Relation) GetSpeciesRelation(species string) HostilityRelation {
	if f.SpeciesRelation == nil {
		return HostilityRelation(0) // Default to neutral if no relations are defined
	}
	if relation, exists := f.SpeciesRelation[species]; exists {
		return relation
	}
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
}

// ChangeSpeciesRelation adjusts the hostility relation for a specific species by a change amount.
func (f *Relation) ChangeSpeciesRelation(species string, change int) HostilityRelation {
	if f.SpeciesRelation == nil {
		f.SpeciesRelation = make(SpeciesRelation)
		f.SpeciesRelation[species] = HostilityRelation(change)
	}

	if relation, exists := f.SpeciesRelation[species]; exists {
		f.SpeciesRelation[species] = relation + HostilityRelation(change)
	} else {
		f.SpeciesRelation[species] = HostilityRelation(change)
	}
	return f.SpeciesRelation[species]
}

// SetFactionRelation sets the hostility relation for a specific faction.
func (f *Relation) SetFactionRelation(faction string, relation HostilityRelation) {
	if f.FactionRelation == nil {
		f.FactionRelation = make(FactionRelation)
	}
	f.FactionRelation[faction] = relation
}

// GetFactionRelation retrieves the hostility relation for a specific faction.
func (f *Relation) GetFactionRelation(faction string) HostilityRelation {
	if f.FactionRelation == nil {
		return HostilityRelation(0) // Default to neutral if no relations are defined
	}
	if relation, exists := f.FactionRelation[faction]; exists {
		return relation
	}
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
}

// ChangeFactionRelation adjusts the hostility relation for a specific faction by a change amount.
func (f *Relation) ChangeFactionRelation(faction string, change int) HostilityRelation {
	if f.FactionRelation == nil {
		f.FactionRelation = make(FactionRelation)
		f.FactionRelation[faction] = HostilityRelation(change)
	}

	if relation, exists := f.FactionRelation[faction]; exists {
		f.FactionRelation[faction] = relation + HostilityRelation(change)
	} else {
		f.FactionRelation[faction] = HostilityRelation(change)
	}
	return f.FactionRelation[faction]
}

// SetProfessionRelation sets the hostility relation for a specific profession.
func (f *Relation) SetProfessionRelation(profession string, relation HostilityRelation) {
	if f.ProfessionRelation == nil {
		f.ProfessionRelation = make(ProfessionRelation)
	}
	f.ProfessionRelation[profession] = relation
}

// GetProfessionRelation retrieves the hostility relation for a specific profession.
func (f *Relation) GetProfessionRelation(profession string) HostilityRelation {
	if f.ProfessionRelation == nil {
		return HostilityRelation(0) // Default to neutral if no relations are defined
	}
	if relation, exists := f.ProfessionRelation[profession]; exists {
		return relation
	}
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
}

// ChangeProfessionRelation adjusts the hostility relation for a specific profession by a change amount.
func (f *Relation) ChangeProfessionRelation(profession string, change int) HostilityRelation {
	if f.ProfessionRelation == nil {
		f.ProfessionRelation = make(ProfessionRelation)
		f.ProfessionRelation[profession] = HostilityRelation(change)
	}

	if relation, exists := f.ProfessionRelation[profession]; exists {
		f.ProfessionRelation[profession] = relation + HostilityRelation(change)
	} else {
		f.ProfessionRelation[profession] = HostilityRelation(change)
	}
	return f.ProfessionRelation[profession]
}

// SetEntityRelation sets the hostility relation for a specific entity.
func (f *Relation) SetEntityRelation(entityID string, relation HostilityRelation) {
	if f.EntityRelation == nil {
		f.EntityRelation = make(EntityRelation)
	}
	f.EntityRelation[entityID] = relation
}

// GetEntityRelation retrieves the hostility relation for a specific entity.
func (f *Relation) GetEntityRelation(entityID string) HostilityRelation {
	if f.EntityRelation == nil {
		return HostilityRelation(0) // Default to neutral if no relations are defined
	}
	if relation, exists := f.EntityRelation[entityID]; exists {
		return relation
	}
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
}

// ChangeEntityRelation adjusts the hostility relation for a specific entity by a change amount.
func (f *Relation) ChangeEntityRelation(entityID string, change int) HostilityRelation {
	if f.EntityRelation == nil {
		f.EntityRelation = make(EntityRelation)
		f.EntityRelation[entityID] = HostilityRelation(change)
	}

	if relation, exists := f.EntityRelation[entityID]; exists {
		f.EntityRelation[entityID] = relation + HostilityRelation(change)
	} else {
		f.EntityRelation[entityID] = HostilityRelation(change)
	}
	return f.EntityRelation[entityID]
}

// CombineRelation merges two Relation structs, summing up relations for overlapping keys.
func CombineRelation(h1, h2 Relation) Relation {
	combined := Relation{
		SpeciesRelation:    make(SpeciesRelation),
		FactionRelation:    make(FactionRelation),
		ProfessionRelation: make(ProfessionRelation),
		EntityRelation:     make(EntityRelation),
	}

	for species, relation := range h1.SpeciesRelation {
		combined.SpeciesRelation[species] = relation
	}
	for species, relation := range h2.SpeciesRelation {
		if existing, exists := combined.SpeciesRelation[species]; exists {
			combined.SpeciesRelation[species] = existing + relation
		} else {
			combined.SpeciesRelation[species] = relation
		}
	}

	for faction, relation := range h1.FactionRelation {
		combined.FactionRelation[faction] = relation
	}
	for faction, relation := range h2.FactionRelation {
		if existing, exists := combined.FactionRelation[faction]; exists {
			combined.FactionRelation[faction] = existing + relation
		} else {
			combined.FactionRelation[faction] = relation
		}
	}

	for profession, relation := range h1.ProfessionRelation {
		combined.ProfessionRelation[profession] = relation
	}
	for profession, relation := range h2.ProfessionRelation {
		if existing, exists := combined.ProfessionRelation[profession]; exists {
			combined.ProfessionRelation[profession] = existing + relation
		} else {
			combined.ProfessionRelation[profession] = relation
		}
	}

	for entityID, relation := range h1.EntityRelation {
		combined.EntityRelation[entityID] = relation
	}
	for entityID, relation := range h2.EntityRelation {
		if existing, exists := combined.EntityRelation[entityID]; exists {
			combined.EntityRelation[entityID] = existing + relation
		} else {
			combined.EntityRelation[entityID] = relation
		}
	}

	return combined
}
