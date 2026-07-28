package entity

type HostilityRelation int

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

func (f *Hostilities) GetEntityRelation(ent Entity) HostilityRelation {
	combined := 0
	if f.entityRelation == nil {
		f.entityRelation[ent.ID]
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
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
