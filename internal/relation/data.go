package relation

// EmptyRelation is a zero-value Relation struct with all relation maps initialized but empty.
var EmptyRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// OrcRelation defines the default hostility relations for Orc species.
var OrcRelation = Relation{
	SpeciesRelation:    SpeciesRelation{"elf": -100, "orc": 50},                         // SpeciesRelation (Populated)
	FactionRelation:    FactionRelation{"orcslayer": -1000},                             // FactionRelation (Empty but non-nil)
	ProfessionRelation: ProfessionRelation{"wizard": -20, "rangers": -50, "guard": -50}, // ProfessionRelation (Populated)
	EntityRelation:     EntityRelation{},                                                // EntityRelation (Empty but non-nil)
}

// ChildRelation defines the default hostility relations for Child species.
var ChildRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}
var FarmAnimalRelation = Relation{
	SpeciesRelation:    SpeciesRelation{"wolf": -50},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// DragonRelation defines the default hostility relations for Dragon species.
var DragonRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// HagRelation defines the default hostility relations for Hag species.
var HagRelation = Relation{
	SpeciesRelation:    SpeciesRelation{"humans": -50, "elves": -20},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// KoboldRelation defines the default hostility relations for Kobold species.
var KoboldRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// RatKingRelation defines the default hostility relations for Rat King species.
var RatKingRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// HumanRelation defines the default hostility relations for Human species.
var HumanRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// ElfRelation defines the default hostility relations for Elf species.
var ElfRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// GoblinRelation defines the default hostility relations for Goblin species.
var GoblinRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// FeyRelation defines the default hostility relations for Fey species.
var FeyRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}

// ThiefRelation defines the default hostility relations for Thief profession.
var ThiefRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{"thieves_guild": 150, "military": -50, "guard": -50},
	ProfessionRelation: ProfessionRelation{"guard": -50, "thief": 50},
	EntityRelation:     EntityRelation{},
}

// BanditRelation defines the default hostility relations for Bandit profession.
var BanditRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{"bandit": 100, "military": -50, "guard": -50, "thieves_guild": 20},
	ProfessionRelation: ProfessionRelation{"guard": -50, "bandit": 50},
	EntityRelation:     EntityRelation{},
}

// BeastRelation defines the default hostility relations for Beast species.
var BeastRelation = Relation{
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
var UndeadRelation = Relation{
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
var MerchantRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{"merchant_guild": 100, "guard": 20},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}
var VerminRelation = Relation{
	SpeciesRelation:    SpeciesRelation{"human": -50, "elf": -30, "orc": -80, "goblin": -60, "kobold": -70, "fey": -40, "dragon": -100, "hag": -90},
	FactionRelation:    FactionRelation{"vermin_slayers": -100, "military": -50, "guard": -50},
	ProfessionRelation: ProfessionRelation{"hunter": -50, "guard": -50},
	EntityRelation:     EntityRelation{},
}
var CivilianRelation = Relation{
	SpeciesRelation:    SpeciesRelation{},
	FactionRelation:    FactionRelation{"civilian": 100, "guard": 20},
	ProfessionRelation: ProfessionRelation{},
	EntityRelation:     EntityRelation{},
}
var VampireRelation = Relation{
	SpeciesRelation:    SpeciesRelation{"human": -50, "elf": -30, "orc": -80, "goblin": -60, "kobold": -70, "fey": -40, "dragon": -100, "hag": -90},
	FactionRelation:    FactionRelation{"vampire_hunters": -100, "military": -50, "guard": -50},
	ProfessionRelation: ProfessionRelation{"hunter": -50, "guard": -50, "vampire_slayer": -100},
	EntityRelation:     EntityRelation{},
}
