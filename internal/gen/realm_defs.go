package gen

// RealmRoomDef defines the static layout and connectivity metadata for a divine room space.
type RealmRoomDef struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Exits       []string `json:"exits"`
	Atmosphere  string   `json:"atmosphere"`

	// --- FIX: Add the missing field expected by the compiler ---
	DeityIDs []string `json:"deity_ids,omitempty"` // String IDs of gods who reside here
}

// RealmDefs is the exported collection with matching resident assignments.
// Deities from different pantheons are mixed together in each room.
var RealmDefs = []RealmRoomDef{
	{
		ID:          "gilded_high_hall",
		Name:        "The Gilded High Hall",
		Description: "A flashy, gold-plated chamber that smells intensely of sulfur and stale ozone.",
		Exits:       []string{"chancery_of_clouds", "the_cluttered_attic"},
		Atmosphere:  "smoky",
		DeityIDs:    []string{"seus_crackbolt", "low_key", "amater_ashes"},
	},
	{
		ID:          "chancery_of_clouds",
		Name:        "The Chancery of Clouds",
		Description: "A silent, drafty archive room filled with brittle, yellowed paper stacks.",
		Exits:       []string{"gilded_high_hall", "drafty_marble_gallery"},
		Atmosphere:  "drafty",
		DeityIDs:    []string{"othena_the_pedantic", "groan_yin", "froyda_the_thistle"},
	},
	{
		ID:          "drafty_longhouse",
		Name:        "The Drafty Longhouse",
		Description: "A cold, damp hall where massive old tables are warped by ambient moisture.",
		Exits:       []string{"gilded_high_hall", "the_cluttered_attic"},
		Atmosphere:  "cold",
		DeityIDs:    []string{"thurn_the_thumper", "oriz_the_bloodshot", "snoozanoo"},
	},
	{
		ID:          "moldering_root_vault",
		Name:        "The Moldering Root Vault",
		Description: "A cramped basement space packed with spiderwebs and bags of useless zinc coins.",
		Exits:       []string{"the_cluttered_attic", "the_sump_tank"},
		Atmosphere:  "damp",
		DeityIDs:    []string{"haydes_the_hoarder", "tie_o_mat", "raijin_the_rattler"},
	},
	{
		ID:          "the_cluttered_attic",
		Name:        "The Cluttered Attic",
		Description: "A chaotic, drafty storage attic where outcasts huddle together and pull fleas.",
		Exits:       []string{"gilded_high_hall", "drafty_longhouse", "moldering_root_vault"},
		Atmosphere:  "dusty",
		DeityIDs:    []string{"odd_in", "wukong_the_mangy", "vaicna_the_unwashed"},
	},
	{
		ID:          "the_sump_tank",
		Name:        "The Sump Tank",
		Description: "A stagnant, leaky drainage pit collecting the runoff water of the upper plane.",
		Exits:       []string{"moldering_root_vault"},
		Atmosphere:  "stagnant",
		DeityIDs:    []string{"posse_eidon", "ooh_huang", "baa_hamut"},
	},
}
