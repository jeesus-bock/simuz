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
var RealmDefs = []RealmRoomDef{
	{
		ID:          "gilded_high_hall",
		Name:        "The Gilded High Hall",
		Description: "A flashy, gold-plated chamber that smells intensely of sulfur and stale ozone.",
		Exits:       []string{"chancery_of_clouds", "the_cluttered_attic"},
		Atmosphere:  "smoky",
		DeityIDs:    []string{"seus_crackbolt", "oriz_the_bloodshot"}, // Pre-assigning residents
	},
	{
		ID:          "chancery_of_clouds",
		Name:        "The Chancery of Clouds",
		Description: "A silent, drafty archive room filled with brittle, yellowed paper stacks.",
		Exits:       []string{"gilded_high_hall", "drafty_marble_gallery"},
		Atmosphere:  "drafty",
		DeityIDs:    []string{"ooh_huang", "groan_yin"},
	},
	{
		ID:          "drafty_longhouse",
		Name:        "The Drafty Longhouse",
		Description: "A cold, damp hall where massive old tables are warped by ambient moisture.",
		Exits:       []string{"gilded_high_hall", "the_cluttered_attic"},
		Atmosphere:  "cold",
		DeityIDs:    []string{"odd_in", "thurn_the_thumper"},
	},
	{
		ID:          "moldering_root_vault",
		Name:        "The Moldering Root Vault",
		Description: "A cramped basement space packed with spiderwebs and bags of useless zinc coins.",
		Exits:       []string{"the_mildew_corner", "the_sump_tank"},
		Atmosphere:  "damp",
		DeityIDs:    []string{"haydes_the_hoarder", "tie_o_mat"},
	},
	{
		ID:          "the_cluttered_attic",
		Name:        "The Cluttered Attic",
		Description: "A chaotic, drafty storage attic where outcasts huddle together and pull fleas.",
		Exits:       []string{"gilded_high_hall", "drafty_longhouse", "the_mildew_corner"},
		Atmosphere:  "dusty",
		DeityIDs:    []string{"low_key", "wukong_the_mangy", "cuckoo_kan"},
	},
	{
		ID:          "the_sump_tank",
		Name:        "The Sump Tank",
		Description: "A stagnant, leaky drainage pit collecting the runoff water of the upper plane.",
		Exits:       []string{"moldering_root_vault"},
		Atmosphere:  "stagnant",
		DeityIDs:    []string{"posse_eidon", "snoozanoo", "raijin_the_rattler", "choke_the_drenched"},
	},
	{
		ID:          "the_velvet_cesspool",
		Name:        "The Velvet Cesspool",
		Description: "A luxurious, dim alcove filled with rotting silk cushions and poisonous mushrooms.",
		Exits:       []string{"the_mildew_corner"},
		Atmosphere:  "toxic",
		DeityIDs:    []string{"froyda_the_thistle"},
	},
	{
		ID:          "the_mildew_corner",
		Name:        "The Mildew Corner",
		Description: "An overgrown, water-damaged gallery corner where moldy scrolls rot away.",
		Exits:       []string{"the_cluttered_attic", "moldering_root_vault", "the_velvet_cesspool"},
		Atmosphere:  "musty",
		DeityIDs:    []string{"amater_ashes", "vaicna_the_unwashed"},
	},
	{
		ID:          "drafty_marble_gallery",
		Name:        "The Drafty Marble Gallery",
		Description: "A grand, freezing hallway lined with cracked white pillars and heavy stone tablets.",
		Exits:       []string{"chancery_of_clouds"},
		Atmosphere:  "freezing",
		DeityIDs:    []string{"othena_the_pedantic", "baa_hamut"},
	},
}
