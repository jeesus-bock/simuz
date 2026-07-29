package engine

type RealmRoomDef struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Exits       []string `json:"exits"` // Connected room IDs inside the divine plane
	Atmosphere  string   `json:"atmosphere"`
}

var realmDefs = []RealmRoomDef{
	{
		ID:          "gilded_high_hall",
		Name:        "The Gilded High Hall",
		Description: "A flashy, gold-plated chamber that smells intensely of sulfur and stale ozone.",
		Exits:       []string{"chancery_of_clouds", "the_cluttered_attic"},
		Atmosphere:  "smoky",
	},
	{
		ID:          "chancery_of_clouds",
		Name:        "The Chancery of Clouds",
		Description: "A silent, drafty archive room filled with brittle, yellowed paper stacks.",
		Exits:       []string{"gilded_high_hall", "drafty_marble_gallery"},
		Atmosphere:  "drafty",
	},
	{
		ID:          "drafty_longhouse",
		Name:        "The Drafty Longhouse",
		Description: "A cold, damp hall where massive old tables are warped by ambient moisture.",
		Exits:       []string{"gilded_high_hall", "the_cluttered_attic"},
		Atmosphere:  "cold",
	},
	{
		ID:          "moldering_root_vault",
		Name:        "The Moldering Root Vault",
		Description: "A cramped basement space packed with spiderwebs and bags of useless zinc coins.",
		Exits:       []string{"the_mildew_corner", "the_sump_tank"},
		Atmosphere:  "damp",
	},
	{
		ID:          "the_cluttered_attic",
		Name:        "The Cluttered Attic",
		Description: "A chaotic, drafty storage attic where outcasts huddle together and pull fleas.",
		Exits:       []string{"gilded_high_hall", "drafty_longhouse", "the_mildew_corner"},
		Atmosphere:  "dusty",
	},
	{
		ID:          "the_sump_tank",
		Name:        "The Sump Tank",
		Description: "A stagnant, leaky drainage pit collecting the runoff water of the upper plane.",
		Exits:       []string{"moldering_root_vault"},
		Atmosphere:  "stagnant",
	},
	{
		ID:          "the_velvet_cesspool",
		Name:        "The Velvet Cesspool",
		Description: "A luxurious, dim alcove filled with rotting silk cushions and poisonous mushrooms.",
		Exits:       []string{"the_mildew_corner"},
		Atmosphere:  "toxic",
	},
	{
		ID:          "the_mildew_corner",
		Name:        "The Mildew Corner",
		Description: "An overgrown, water-damaged gallery corner where moldy scrolls rot away.",
		Exits:       []string{"the_cluttered_attic", "moldering_root_vault", "the_velvet_cesspool"},
		Atmosphere:  "musty",
	},
	{
		ID:          "drafty_marble_gallery",
		Name:        "The Drafty Marble Gallery",
		Description: "A grand, freezing hallway lined with cracked white pillars and heavy stone tablets.",
		Exits:       []string{"chancery_of_clouds"},
		Atmosphere:  "freezing",
	},
}
