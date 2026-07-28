package entity

type HostilityRelation int

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

type Hostilities struct {
	speciesRelation    SpeciesRelation
	factionRelation    FactionRelation
	professionRelation ProfessionRelation
	entityRelation     EntityRelation
}

type FactionRelation map[string]HostilityRelation
type SpeciesRelation map[string]HostilityRelation
type ProfessionRelation map[string]HostilityRelation
type EntityRelation map[string]HostilityRelation

func (f *Hostilities) Relation(ent Entity) HostilityRelation {
	combined := 0
	if f.entityRelation == nil {
		f.entityRelation = make(EntityRelation)
	}
	if relation, exists := f.entityRelation[ent.ID]; exists {
		combined += relation.Int()
	}
	return HostilityRelation(combined)
}

func (f *Hostilities) SetSpeciesRelation(species string, relation HostilityRelation) {
	if f.speciesRelation == nil {
		f.speciesRelation = make(SpeciesRelation)
	}
	f.speciesRelation[species] = relation
}

func (f *Hostilities) GetSpeciesRelation(species string) HostilityRelation {
	if f.speciesRelation == nil {
		return HostilityRelation(0) // Default to neutral if no relations are defined
	}
	if relation, exists := f.speciesRelation[species]; exists {
		return relation
	}
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
}

func (f *Hostilities) ChangeSpeciesRelation(species string, change int) HostilityRelation {
	if f.speciesRelation == nil {
		f.speciesRelation = make(SpeciesRelation)
		f.speciesRelation[species] = HostilityRelation(change)
	}

	if relation, exists := f.speciesRelation[species]; exists {
		f.speciesRelation[species] = relation + HostilityRelation(change)
	} else {
		f.speciesRelation[species] = HostilityRelation(change)
	}
	return f.speciesRelation[species]
}

// SetFactionRelation sets the hostility relation for a specific faction.
func (f *Hostilities) SetFactionRelation(faction string, relation HostilityRelation) {
	if f.factionRelation == nil {
		f.factionRelation = make(FactionRelation)
	}
	f.factionRelation[faction] = relation
}

// GetFactionRelation retrieves the hostility relation for a specific faction.
func (f *Hostilities) GetFactionRelation(faction string) HostilityRelation {
	if f.factionRelation == nil {
		return HostilityRelation(0) // Default to neutral if no relations are defined
	}
	if relation, exists := f.factionRelation[faction]; exists {
		return relation
	}
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
}

// ChangeFactionRelation adjusts the hostility relation for a specific faction by a change amount.
func (f *Hostilities) ChangeFactionRelation(faction string, change int) HostilityRelation {
	if f.factionRelation == nil {
		f.factionRelation = make(FactionRelation)
		f.factionRelation[faction] = HostilityRelation(change)
	}

	if relation, exists := f.factionRelation[faction]; exists {
		f.factionRelation[faction] = relation + HostilityRelation(change)
	} else {
		f.factionRelation[faction] = HostilityRelation(change)
	}
	return f.factionRelation[faction]
}

// SetProfessionRelation sets the hostility relation for a specific profession.
func (f *Hostilities) SetProfessionRelation(profession string, relation HostilityRelation) {
	if f.professionRelation == nil {
		f.professionRelation = make(ProfessionRelation)
	}
	f.professionRelation[profession] = relation
}

// GetProfessionRelation retrieves the hostility relation for a specific profession.
func (f *Hostilities) GetProfessionRelation(profession string) HostilityRelation {
	if f.professionRelation == nil {
		return HostilityRelation(0) // Default to neutral if no relations are defined
	}
	if relation, exists := f.professionRelation[profession]; exists {
		return relation
	}
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
}

// ChangeProfessionRelation adjusts the hostility relation for a specific profession by a change amount.
func (f *Hostilities) ChangeProfessionRelation(profession string, change int) HostilityRelation {
	if f.professionRelation == nil {
		f.professionRelation = make(ProfessionRelation)
		f.professionRelation[profession] = HostilityRelation(change)
	}

	if relation, exists := f.professionRelation[profession]; exists {
		f.professionRelation[profession] = relation + HostilityRelation(change)
	} else {
		f.professionRelation[profession] = HostilityRelation(change)
	}
	return f.professionRelation[profession]
}

// SetEntityRelation sets the hostility relation for a specific entity.
func (f *Hostilities) SetEntityRelation(entityID string, relation HostilityRelation) {
	if f.entityRelation == nil {
		f.entityRelation = make(EntityRelation)
	}
	f.entityRelation[entityID] = relation
}

// GetEntityRelation retrieves the hostility relation for a specific entity.
func (f *Hostilities) GetEntityRelation(entityID string) HostilityRelation {
	if f.entityRelation == nil {
		return HostilityRelation(0) // Default to neutral if no relations are defined
	}
	if relation, exists := f.entityRelation[entityID]; exists {
		return relation
	}
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
}

// ChangeEntityRelation adjusts the hostility relation for a specific entity by a change amount.
func (f *Hostilities) ChangeEntityRelation(entityID string, change int) HostilityRelation {
	if f.entityRelation == nil {
		f.entityRelation = make(EntityRelation)
		f.entityRelation[entityID] = HostilityRelation(change)
	}

	if relation, exists := f.entityRelation[entityID]; exists {
		f.entityRelation[entityID] = relation + HostilityRelation(change)
	} else {
		f.entityRelation[entityID] = HostilityRelation(change)
	}
	return f.entityRelation[entityID]
}
