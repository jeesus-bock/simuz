package entity

type Faction struct {
	ID          string
	Name        string
	Hostilities Hostilities
}

func GetFactionByID(id string) (*Faction, bool) {
	faction, exists := FactionRegistry[id]
	return faction, exists
}

// FactionRegistry is the one source of truth for all faction data in simuz.
// Every faction used in the simulation must have an entry here.
var FactionRegistry = map[string]*Faction{
	"civilian": {
		ID:          "civilian",
		Name:        "Civilian",
		Hostilities: CivilianHostilities,
	},
	"merchant": {
		ID:          "merchant",
		Name:        "Merchant",
		Hostilities: MerchantHostilities,
	},
	"beast": {
		ID:          "garkhorn",
		Name:        "Garkhorn",
		Hostilities: BeastHostilities,
	},
	"undead": {
		ID:          "undead",
		Name:        "Undead",
		Hostilities: UndeadHostilities,
	},
	"deity": {
		ID:          "deity",
		Name:        "Deity",
		Hostilities: EmptyHostilities,
	},
}
