// Package entity defines core simulation types, including entities, attributes, and hostility relations.
package entity

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

type Hostilities struct {
	SpeciesRelation    SpeciesRelation    `json:"speciesRelation,omitempty"`
	FactionRelation    FactionRelation    `json:"factionRelation,omitempty"`
	ProfessionRelation ProfessionRelation `json:"professionRelation,omitempty"`
	EntityRelation     EntityRelation     `json:"entityRelation,omitempty"`
}

// Relation calculates the combined hostility relation based on species, faction, profession, and entity relations.
func (f *Hostilities) Relation(ent Entity) HostilityRelation {
	combined := 0
	combined += f.GetEntityRelation(ent.ID).Int()
	combined += f.GetFactionRelation(ent.Faction).Int()
	combined += f.GetSpeciesRelation(ent.Species).Int()
	combined += f.GetProfessionRelation(ent.Profession).Int()
	return HostilityRelation(combined)
}
func (f *Hostilities) FactinRelation(fac Faction) HostilityRelation {
	combined := 0
	combined += f.GetFactionRelation(fac.ID).Int()

	return HostilityRelation(combined)
}

// SetSpeciesRelation sets the hostility relation for a specific species.
func (f *Hostilities) SetSpeciesRelation(species string, relation HostilityRelation) {
	if f.SpeciesRelation == nil {
		f.SpeciesRelation = make(SpeciesRelation)
	}
	f.SpeciesRelation[species] = relation
}

// GetSpeciesRelation retrieves the hostility relation for a specific species.
func (f *Hostilities) GetSpeciesRelation(species string) HostilityRelation {
	if f.SpeciesRelation == nil {
		return HostilityRelation(0) // Default to neutral if no relations are defined
	}
	if relation, exists := f.SpeciesRelation[species]; exists {
		return relation
	}
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
}

// ChangeSpeciesRelation adjusts the hostility relation for a specific species by a change amount.
func (f *Hostilities) ChangeSpeciesRelation(species string, change int) HostilityRelation {
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
func (f *Hostilities) SetFactionRelation(faction string, relation HostilityRelation) {
	if f.FactionRelation == nil {
		f.FactionRelation = make(FactionRelation)
	}
	f.FactionRelation[faction] = relation
}

// GetFactionRelation retrieves the hostility relation for a specific faction.
func (f *Hostilities) GetFactionRelation(faction string) HostilityRelation {
	if f.FactionRelation == nil {
		return HostilityRelation(0) // Default to neutral if no relations are defined
	}
	if relation, exists := f.FactionRelation[faction]; exists {
		return relation
	}
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
}

// ChangeFactionRelation adjusts the hostility relation for a specific faction by a change amount.
func (f *Hostilities) ChangeFactionRelation(faction string, change int) HostilityRelation {
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
func (f *Hostilities) SetProfessionRelation(profession string, relation HostilityRelation) {
	if f.ProfessionRelation == nil {
		f.ProfessionRelation = make(ProfessionRelation)
	}
	f.ProfessionRelation[profession] = relation
}

// GetProfessionRelation retrieves the hostility relation for a specific profession.
func (f *Hostilities) GetProfessionRelation(profession string) HostilityRelation {
	if f.ProfessionRelation == nil {
		return HostilityRelation(0) // Default to neutral if no relations are defined
	}
	if relation, exists := f.ProfessionRelation[profession]; exists {
		return relation
	}
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
}

// ChangeProfessionRelation adjusts the hostility relation for a specific profession by a change amount.
func (f *Hostilities) ChangeProfessionRelation(profession string, change int) HostilityRelation {
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
func (f *Hostilities) SetEntityRelation(entityID string, relation HostilityRelation) {
	if f.EntityRelation == nil {
		f.EntityRelation = make(EntityRelation)
	}
	f.EntityRelation[entityID] = relation
}

// GetEntityRelation retrieves the hostility relation for a specific entity.
func (f *Hostilities) GetEntityRelation(entityID string) HostilityRelation {
	if f.EntityRelation == nil {
		return HostilityRelation(0) // Default to neutral if no relations are defined
	}
	if relation, exists := f.EntityRelation[entityID]; exists {
		return relation
	}
	return HostilityRelation(0) // Default to neutral if no specific relation is defined
}

// ChangeEntityRelation adjusts the hostility relation for a specific entity by a change amount.
func (f *Hostilities) ChangeEntityRelation(entityID string, change int) HostilityRelation {
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

// CombineHostilities merges two Hostilities structs, summing up relations for overlapping keys.
func CombineHostilities(h1, h2 Hostilities) Hostilities {
	combined := Hostilities{
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

// EmptyHostilities is a zero-value Hostilities struct with all relation maps initialized but empty.
var EmptyHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// OrcHostilities defines the default hostility relations for Orc species.
var OrcHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{"elf": -100, "orc": 50},                         // SpeciesRelation (Populated)
	FactionRelation:    FactionRelation{"orcslayer": -1000},                             // FactionRelation (Empty but non-nil)
	ProfessionRelation: ProfessionRelation{"wizard": -20, "rangers": -50, "guard": -50}, // ProfessionRelation (Populated)
	EntityRelation:     EntityRelation{},                                                // EntityRelation (Empty but non-nil)
}

// ChildHostilities defines the default hostility relations for Child species.
var ChildHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}
var FarmAnimalHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{"wolf": -50},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// DragonHostilities defines the default hostility relations for Dragon species.
var DragonHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// HagHostilities defines the default hostility relations for Hag species.
var HagHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{"humans": -50, "elves": -20},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// KoboldHostilities defines the default hostility relations for Kobold species.
var KoboldHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// RatKingHostilities defines the default hostility relations for Rat King species.
var RatKingHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// HumanHostilities defines the default hostility relations for Human species.
var HumanHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// ElfHostilities defines the default hostility relations for Elf species.
var ElfHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// GoblinHostilities defines the default hostility relations for Goblin species.
var GoblinHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// FeyHostilities defines the default hostility relations for Fey species.
var FeyHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// ThiefHostilities defines the default hostility relations for Thief profession.
var ThiefHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{"thieves_guild": 150, "military": -50, "guard": -50},
	ProfessionRelation: ProfessionRelation{"guard": -50, "thief": 50},
	EntityRelation:     EntityRelation{},
}

// BanditHostilities defines the default hostility relations for Bandit profession.
var BanditHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{"bandit": 100, "military": -50, "guard": -50, "thieves_guild": 20},
	ProfessionRelation: ProfessionRelation{"guard": -50, "bandit": 50},
	EntityRelation:     EntityRelation{},
}

// BeastHostilities defines the default hostility relations for Beast species.
var BeastHostilities = Hostilities{
	SpeciesRelation: SpeciesRelation{
		"human":  -50,
		"elf":    -30,
		"orc":    -80,
		"goblin": -60,
		"kobold": -70,
		"fey":    -40,
		"dragon": -100,
		"hag":    -90,
		"rat":    -100,
	},
	FactionRelation:    FactionRelation{"beast_slayers": -100, "military": -50, "guard": -50},
	ProfessionRelation: ProfessionRelation{"hunter": -50, "guard": -50},
	EntityRelation:     EntityRelation{},
}
var UndeadHostilities = Hostilities{
	SpeciesRelation: SpeciesRelation{
		"human":  -50,
		"elf":    -30,
		"orc":    -80,
		"goblin": -60,
		"kobold": -70,
		"fey":    -40,
		"dragon": -100,
		"hag":    -90,
		"rat":    -100,
	},
	FactionRelation:    FactionRelation{"undead_slayers": -100, "military": -50, "guard": -50},
	ProfessionRelation: ProfessionRelation{"vampire_hunter": -50, "guard": -50},
	EntityRelation:     EntityRelation{},
}
var MerchantHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{"merchant_guild": 100, "guard": 20},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}
var VerminHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{"human": -50, "elf": -30, "orc": -80, "goblin": -60, "kobold": -70, "fey": -40, "dragon": -100, "hag": -90},
	FactionRelation:    FactionRelation{"vermin_slayers": -100, "military": -50, "guard": -50},
	ProfessionRelation: ProfessionRelation{"hunter": -50, "guard": -50},
	EntityRelation:     EntityRelation{},
}
var CivilianHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{"civilian": 100, "guard": 20},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}
var VampireHostilities = Hostilities{
	SpeciesRelation:    SpeciesRelation{"human": -50, "elf": -30, "orc": -80, "goblin": -60, "kobold": -70, "fey": -40, "dragon": -100, "hag": -90},
	FactionRelation:    FactionRelation{"vampire_hunters": -100, "military": -50, "guard": -50},
	ProfessionRelation: ProfessionRelation{"hunter": -50, "guard": -50, "vampire_slayer": -100},
	EntityRelation:     EntityRelation{},
}
