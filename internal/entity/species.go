// Package entity defines the simulation entities, their attributes, and related behaviors.
package entity

// Species defines the base data for a creature species in the simulation.
// It is the single source of truth for all species-related information.
type Species struct {
	ID                  string
	Name                string
	MaxAge              int
	CanLevelUp          bool
	CanReproduce        bool
	IsCaveman           bool
	IsImmortal          bool
	GestationTicks      int
	DefaultScripts      []string
	DefaultSleepCycle   string // "diurnal", "nocturnal", "none"
	AutoFeed            bool
	StarvationThreshold int // ticks before starvation damage begins; 0 means immune
	Names               []string
	BaseAttrs           Attributes
}

// SpeciesRegistry is the one source of truth for all species data in simuz.
// Every species used in the simulation must have an entry here.
var SpeciesRegistry = map[string]Species{
	"human": {
		ID:                  "human",
		Name:                "Human",
		MaxAge:              30000,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      280,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 259200, // 3 days
		Names:               []string{"Aldric", "Brenna", "Cedric", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jenna", "Kol", "Lyssa", "Maren", "Nolan", "Opal", "Petra", "Quinn", "Rhea", "Soren", "Tessa"},
		BaseAttrs:           Attributes{STR: 10, DEX: 10, CON: 10, INT: 10, WIS: 10, CHA: 10},
	},
	"elf": {
		ID:                  "elf",
		Name:                "Elf",
		MaxAge:              60000,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      300,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 345600, // 4 days
		Names:               []string{"Aldric", "Brenna", "Cedric", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jenna", "Kol", "Lyssa", "Maren", "Nolan", "Opal", "Petra", "Quinn", "Rhea", "Soren", "Tessa"},
		BaseAttrs:           Attributes{STR: 10, DEX: 12, CON: 10, INT: 12, WIS: 12, CHA: 10},
	},
	"orc": {
		ID:                  "orc",
		Name:                "Orc",
		MaxAge:              22000,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           true,
		IsImmortal:          false,
		GestationTicks:      200,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 172800, // 2 days
		Names:               []string{"Mog", "Zog", "Thrak", "Gruul", "Drok", "Krag", "Snag", "Ruk"},
		BaseAttrs:           Attributes{STR: 14, DEX: 10, CON: 13, INT: 6, WIS: 6, CHA: 5},
	},
	"goblin": {
		ID:                  "goblin",
		Name:                "Goblin",
		MaxAge:              15000,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      100,
		DefaultScripts:      []string{"gathering"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 129600, // 1.5 days
		Names:               []string{"Snag", "Grib", "Nog", "Blink", "Mug"},
		BaseAttrs:           Attributes{STR: 8, DEX: 12, CON: 10, INT: 8, WIS: 6, CHA: 6},
	},
	"kobold": {
		ID:                  "kobold",
		Name:                "Kobold",
		MaxAge:              12000,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      80,
		DefaultScripts:      []string{"kobold"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 129600, // 1.5 days
		Names:               []string{"Skrit", "Yip", "Klik", "Drak", "Snik"},
		BaseAttrs:           Attributes{STR: 8, DEX: 14, CON: 9, INT: 8, WIS: 7, CHA: 6},
	},
	"fey": {
		ID:                  "fey",
		Name:                "Fey",
		MaxAge:              50000,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      250,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 86400, // 1 day
		Names:               []string{"Aldric", "Brenna", "Cedric", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jenna", "Kol", "Lyssa", "Maren", "Nolan", "Opal", "Petra", "Quinn", "Rhea", "Soren", "Tessa"},
		BaseAttrs:           Attributes{STR: 10, DEX: 11, CON: 10, INT: 11, WIS: 11, CHA: 12},
	},
	"rat_king": {
		ID:                  "rat_king",
		Name:                "Rat King",
		MaxAge:              10000,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      30,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "nocturnal",
		AutoFeed:            true,
		StarvationThreshold: 172800, // 2 days
		Names:               []string{"Squeak", "Nibble", "Skitter", "Dart", "Pip"},
		BaseAttrs:           Attributes{STR: 6, DEX: 12, CON: 8, INT: 2, WIS: 5, CHA: 2},
	},
	"vampire": {
		ID:                  "vampire",
		Name:                "Vampire",
		MaxAge:              0,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          true,
		GestationTicks:      280,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "nocturnal",
		AutoFeed:            true,
		StarvationThreshold: 432000, // 5 days
		Names:               []string{"Aldric", "Brenna", "Cedric", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jenna", "Kol", "Lyssa", "Maren", "Nolan", "Opal", "Petra", "Quinn", "Rhea", "Soren", "Tessa"},
		BaseAttrs:           Attributes{STR: 12, DEX: 12, CON: 12, INT: 10, WIS: 10, CHA: 10},
	},
	"hag": {
		ID:                  "hag",
		Name:                "Hag",
		MaxAge:              20000,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      200,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 172800, // 2 days
		Names:               []string{"Aldric", "Brenna", "Cedric", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jenna", "Kol", "Lyssa", "Maren", "Nolan", "Opal", "Petra", "Quinn", "Rhea", "Soren", "Tessa"},
		BaseAttrs:           Attributes{STR: 10, DEX: 10, CON: 12, INT: 8, WIS: 10, CHA: 8},
	},
	"deity": {
		ID:                  "deity",
		Name:                "Deity",
		MaxAge:              0,
		CanLevelUp:          true,
		CanReproduce:        false,
		IsCaveman:           false,
		IsImmortal:          true,
		GestationTicks:      0,
		DefaultScripts:      []string{},
		DefaultSleepCycle:   "none",
		AutoFeed:            true,
		StarvationThreshold: 0, // immune
		Names:               []string{"Zeus", "Odin", "Ra", "Vishnu", "Athena", "Loki", "Thor", "Isis", "Osiris", "Kali"},
		BaseAttrs:           Attributes{STR: 20, DEX: 20, CON: 20, INT: 20, WIS: 20, CHA: 20},
	},
	"ogre": {
		ID:                  "ogre",
		Name:                "Ogre",
		MaxAge:              18000,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           true,
		IsImmortal:          false,
		GestationTicks:      220,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 259200, // 3 days
		Names:               []string{"Mog", "Zog", "Thrak", "Gruul", "Drok", "Krag", "Snag", "Ruk"},
		BaseAttrs:           Attributes{STR: 18, DEX: 8, CON: 16, INT: 3, WIS: 5, CHA: 3},
	},
	"giant": {
		ID:                  "giant",
		Name:                "Giant",
		MaxAge:              25000,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           true,
		IsImmortal:          false,
		GestationTicks:      260,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 259200, // 3 days
		Names:               []string{"Mog", "Zog", "Thrak", "Gruul", "Drok", "Krag", "Snag", "Ruk"},
		BaseAttrs:           Attributes{STR: 20, DEX: 6, CON: 18, INT: 2, WIS: 6, CHA: 2},
	},
	"troll": {
		ID:                  "troll",
		Name:                "Troll",
		MaxAge:              15000,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           true,
		IsImmortal:          false,
		GestationTicks:      180,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 172800, // 2 days
		Names:               []string{"Mog", "Zog", "Thrak", "Gruul", "Drok", "Krag", "Snag", "Ruk"},
		BaseAttrs:           Attributes{STR: 16, DEX: 6, CON: 14, INT: 2, WIS: 4, CHA: 2},
	},
	"cyclops": {
		ID:                  "cyclops",
		Name:                "Cyclops",
		MaxAge:              12000,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           true,
		IsImmortal:          false,
		GestationTicks:      160,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 172800, // 2 days
		Names:               []string{"Mog", "Zog", "Thrak", "Gruul", "Drok", "Krag", "Snag", "Ruk"},
		BaseAttrs:           Attributes{STR: 18, DEX: 5, CON: 16, INT: 2, WIS: 5, CHA: 1},
	},
	"wolf": {
		ID:                  "wolf",
		Name:                "Wolf",
		MaxAge:              0,
		CanLevelUp:          false,
		CanReproduce:        false,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      60,
		DefaultScripts:      []string{"hunting"},
		DefaultSleepCycle:   "nocturnal",
		AutoFeed:            true,
		StarvationThreshold: 172800, // 2 days
		Names:               []string{"Howl", "Rip", "Claw", "Fang", "Snap", "Growl"},
		BaseAttrs:           Attributes{STR: 12, DEX: 14, CON: 11, INT: 3, WIS: 7, CHA: 3},
	},
	"bear": {
		ID:                  "bear",
		Name:                "Bear",
		MaxAge:              0,
		CanLevelUp:          false,
		CanReproduce:        false,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      90,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 172800, // 2 days
		Names:               []string{"Claw", "Grunt", "Maw", "Huff"},
		BaseAttrs:           Attributes{STR: 18, DEX: 9, CON: 16, INT: 2, WIS: 6, CHA: 2},
	},
	"boar": {
		ID:                  "boar",
		Name:                "Boar",
		MaxAge:              0,
		CanLevelUp:          false,
		CanReproduce:        false,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      70,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 172800, // 2 days
		Names:               []string{"Snout", "Tusk", "Grunta", "Bristle"},
		BaseAttrs:           Attributes{STR: 14, DEX: 11, CON: 13, INT: 2, WIS: 5, CHA: 2},
	},
	"rat": {
		ID:                  "rat",
		Name:                "Rat",
		MaxAge:              0,
		CanLevelUp:          false,
		CanReproduce:        false,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      30,
		DefaultScripts:      []string{"defensive"},
		DefaultSleepCycle:   "nocturnal",
		AutoFeed:            true,
		StarvationThreshold: 86400, // 1 day
		Names:               []string{"Squeak", "Nibble", "Skitter", "Dart", "Pip"},
		BaseAttrs:           Attributes{STR: 6, DEX: 12, CON: 8, INT: 2, WIS: 5, CHA: 2},
	},
	"spider": {
		ID:                  "spider",
		Name:                "Spider",
		MaxAge:              0,
		CanLevelUp:          false,
		CanReproduce:        false,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      40,
		DefaultScripts:      []string{"scouting"},
		DefaultSleepCycle:   "nocturnal",
		AutoFeed:            true,
		StarvationThreshold: 129600, // 1.5 days
		Names:               []string{"Legs", "Weaver", "Sting", "Crawl"},
		BaseAttrs:           Attributes{STR: 10, DEX: 15, CON: 8, INT: 2, WIS: 7, CHA: 2},
	},
}

// GetSpecies returns the Species definition for the given species ID.
// Returns an empty Species if the species is not registered.
func GetSpecies(id string) Species {
	return SpeciesRegistry[id]
}

// StarvationDamageInterval returns the tick interval at which starvation damage is applied.
func StarvationDamageInterval() int {
	return 10
}

func StarvationDamageMin() int {
	return 1
}

func StarvationDamageMax() int {
	return 5
}
