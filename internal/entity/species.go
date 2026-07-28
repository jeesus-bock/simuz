package entity

// Species defines the base data for a creature species in the simulation.
// It is the single source of truth for all species-related information.
type Species struct {
	ID                  string
	Name                string
	MaxAge              int
	AdultAge            int
	CanLevelUp          bool
	CanReproduce        bool
	IsCaveman           bool
	IsImmortal          bool
	GestationTicks      int
	DefaultScripts      []string
	DefaultSleepCycle   string // "diurnal", "nocturnal", "none"
	AutoFeed            bool
	StarvationThreshold int // ticks before starvation damage begins; 0 means immune
	MaleNames           []string
	FemaleNames         []string
	BaseAttrs           Attributes
}

// SpeciesRegistry is the one source of truth for all species data in simuz.
// Every species used in the simulation must have an entry here.
var SpeciesRegistry = map[string]Species{
	"human": {
		ID:                  "human",
		Name:                "Human",
		MaxAge:              100, // years
		AdultAge:            18,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      280,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 259200, // 3 days
		MaleNames:           []string{"Aldric", "Brenna", "Cedric", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jenna", "Kol", "Lyssa", "Maren", "Nolan", "Opal", "Petra", "Quinn", "Rhea", "Soren", "Tessa"},
		FemaleNames:         []string{"Aldrica", "Brenna", "Cedrica", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jenna", "Kol", "Lyssa", "Maren", "Nolan", "Opal", "Petra", "Quinn", "Rhea", "Soren", "Tessa"},
		BaseAttrs:           Attributes{STR: 10, DEX: 10, CON: 10, INT: 10, WIS: 10, CHA: 10},
	},
	"elf": {
		ID:                  "elf",
		Name:                "Elf",
		MaxAge:              500, // years
		AdultAge:            100,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      300,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 345600, // 4 days
		MaleNames:           []string{"Aldric", "Brenna", "Cedric", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jenna", "Kol", "Lyssa", "Maren", "Nolan", "Opal", "Petra", "Quinn", "Rhea", "Soren", "Tessa"},
		FemaleNames:         []string{"Aldrica", "Brenna", "Cedrica", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jenna", "Kol", "Lyssa", "Maren", "Nolan", "Opal", "Petra", "Quinn", "Rhea", "Soren", "Tessa"},
		BaseAttrs:           Attributes{STR: 10, DEX: 12, CON: 10, INT: 12, WIS: 12, CHA: 10},
	},
	"orc": {
		ID:                  "orc",
		Name:                "Orc",
		MaxAge:              40, // years
		AdultAge:            12,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           true,
		IsImmortal:          false,
		GestationTicks:      200,
		DefaultScripts:      []string{"aggressive", "raiding"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 172800, // 2 days
		MaleNames:           []string{"Mog", "Zog", "Thrak", "Gruul", "Drok", "Krag", "Snag", "Ruk"},
		FemaleNames:         []string{"Mog", "Zog", "Thrak", "Gruul", "Drok", "Krag", "Snag", "Ruk"},
		BaseAttrs:           Attributes{STR: 14, DEX: 10, CON: 13, INT: 6, WIS: 6, CHA: 5},
	},
	"goblin": {
		ID:                  "goblin",
		Name:                "Goblin",
		MaxAge:              30, // years
		AdultAge:            10,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      100,
		DefaultScripts:      []string{"gathering"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 129600, // 1.5 days
		MaleNames:           []string{"Göz", "Snag", "Grib", "Nog", "Blink", "Mug"},
		FemaleNames:         []string{"Snag", "Grib", "Nog", "Blink", "Mug"},
		BaseAttrs:           Attributes{STR: 8, DEX: 12, CON: 10, INT: 8, WIS: 6, CHA: 6},
	},
	"kobold": {
		ID:                  "kobold",
		Name:                "Kobold",
		MaxAge:              25, // years
		AdultAge:            5,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      80,
		DefaultScripts:      []string{"kobold"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 129600, // 1.5 days
		MaleNames:           []string{"Andres", "Margus", "Toomas", "Kristjan", "Martin", "Priit", "Sander", "Indrek", "Jaan", "Markus", "Rasmus", "Tanel", "Kaurits", "Kalle"},
		FemaleNames:         []string{"Mari", "Katriin", "Triin", "Pille", "Kadri", "Tiina", "Laura", "Eveli", "Sirje", "Kristel", "Anu", "Katrin"},
		BaseAttrs:           Attributes{STR: 8, DEX: 14, CON: 9, INT: 8, WIS: 7, CHA: 6},
	},
	"fey": {
		ID:                  "fey",
		Name:                "Fey",
		MaxAge:              200, // years
		AdultAge:            50,
		CanLevelUp:          true,
		CanReproduce:        true,
		IsCaveman:           false,
		IsImmortal:          false,
		GestationTicks:      250,
		DefaultScripts:      []string{"aggressive"},
		DefaultSleepCycle:   "diurnal",
		AutoFeed:            true,
		StarvationThreshold: 86400, // 1 day
		MaleNames:           []string{"Aldric", "Brenna", "Cedric", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jenna", "Kol", "Lyssa", "Maren", "Nolan", "Opal", "Petra", "Quinn", "Rhea", "Soren", "Tessa"},
		FemaleNames:         []string{"Aldrica", "Brenna", "Cedrica", "Delara", "Eamon", "Fiona", "Gareth", "Hilda", "Ivan", "Jen