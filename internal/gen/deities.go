// Package gen contains world generation helpers and seeded simulation setup utilities.
package gen

import (
	"log"

	"simuz/internal/entity"
)

type deityDef struct {
	ID          string
	Name        string
	Pantheon    string
	Domain      string
	Attributes  entity.Attributes
	Active      bool
	RealmRoomID string
	Needs       []string
	Shape       string
	WorshippedBy []string // species that worship this deity
}

var DeityDefs = []deityDef{
	// Room 1: gilded_high_hall — greek + norse + japanese
	{
		ID:          "seus_crackbolt",
		Name:        "Seus Crackbolt",
		Pantheon:    "greek_offbrand",
		Domain:      "leaky_thunder_petty_grudges",
		RealmRoomID: "gilded_high_hall",
		Needs:       []string{"cheap_wine", "lard", "melted_copper"},
		Shape:       "A mountain of a man bound to an unpolished iron chair, whose beard constantly drips grey kitchen ash. His eyes flicker like dying candle stubs.",
		Attributes:  entity.Attributes{STR: 31, DEX: 22, CON: 38, INT: 24, WIS: 25, CHA: 35},
		WorshippedBy: []string{"human", "hobbit", "orc", "goblin", "ogre", "lizardfolk", "hobgoblin", "gnoll"},
	},
	{
		ID:          "low_key",
		Name:        "Low-Key the Fumbler",
		Pantheon:    "norse_offbrand",
		Domain:      "bad_advice_spilled_grease",
		RealmRoomID: "gilded_high_hall",
		Needs:       []string{"greased_strings", "rotten_eggs", "attention"},
		Shape:       "A gaunt, twitchy trickster whose shape-shifting is broken. He occasionally gets stuck halfway between an elven merchant and a wet sewer fox.",
		Attributes:  entity.Attributes{STR: 19, DEX: 35, CON: 22, INT: 35, WIS: 20, CHA: 35},
		WorshippedBy: []string{"human", "hobbit"},
	},
	{
		ID:          "amater_ashes",
		Name:        "Amater-Ashes the Dim",
		Pantheon:    "japanese_offbrand",
		Domain:      "flickering_lanterns_bad_sunburns",
		RealmRoomID: "gilded_high_hall",
		Needs:       []string{"tallow_candles", "blankets", "clear_night"},
		Shape:       "A shivering goddess wrapped in coarse woolen quilts, clutching a rusted oil lantern that refuses to stay lit for more than ten ticks.",
		Attributes:  entity.Attributes{STR: 22, DEX: 25, CON: 29, INT: 31, WIS: 36, CHA: 39},
		WorshippedBy: []string{"elf"},
	},

	// Room 2: chancery_of_clouds — greek + chinese + norse
	{
		ID:          "othena_the_pedantic",
		Name:        "Othena the Pedantic",
		Pantheon:    "greek_offbrand",
		Domain:      "grammatical_heresy_unwinnable_debates",
		RealmRoomID: "chancery_of_clouds",
		Needs:       []string{"ink_vials", "herbal_tea", "wax_candles"},
		Shape:       "A stern, unblinking woman carrying an excessively heavy stone tablet, constantly sighing and muttering corrections at mortal grammar errors.",
		Attributes:  entity.Attributes{STR: 24, DEX: 27, CON: 25, INT: 39, WIS: 35, CHA: 28},
		WorshippedBy: []string{"elf"},
	},
	{
		ID:          "groan_yin",
		Name:        "Groan-Yin the Sighing",
		Pantheon:    "chinese_offbrand",
		Domain:      "passive_aggressive_pity",
		RealmRoomID: "chancery_of_clouds",
		Needs:       []string{"bandages", "soothing_balm", "quiet_crying"},
		Shape:       "A weeping entity sitting in a lotus position, who listens to mortal prayers only to sigh heavily and tell them their problems are their own fault.",
		Attributes:  entity.Attributes{STR: 15, DEX: 19, CON: 27, INT: 35, WIS: 40, CHA: 39},
		WorshippedBy: []string{"elf"},
	},
	{
		ID:          "froyda_the_thistle",
		Name:        "Froyda the Thistle-Queen",
		Pantheon:    "norse_offbrand",
		Domain:      "obsessive_infatuation_stinging_betrayals",
		RealmRoomID: "chancery_of_clouds",
		Needs:       []string{"violet_spores", "honey_pots", "gossip"},
		Shape:       "A majestic figure whose porcelain skin cracks like dry clay upon close inspection, revealing writhing green stinging nettles beneath her silk dress.",
		Attributes:  entity.Attributes{STR: 22, DEX: 27, CON: 25, INT: 27, WIS: 28, CHA: 39},
		WorshippedBy: []string{"elf"},
	},

	// Room 3: drafty_longhouse — norse + greek + japanese
	{
		ID:          "thurn_the_thumper",
		Name:        "Thurn the Thumper",
		Pantheon:    "norse_offbrand",
		Domain:      "loud_noises_shattered_handles",
		RealmRoomID: "drafty_longhouse",
		Needs:       []string{"roasted_marrow", "ale_barrels", "anvil_to_smash"},
		Shape:       "A massive brawler who cannot speak below a deafening shout. His legendary war-hammer has a loose head that flies off whenever he swings it.",
		Attributes:  entity.Attributes{STR: 45, DEX: 25, CON: 40, INT: 12, WIS: 15, CHA: 22},
		WorshippedBy: []string{"dwarf"},
	},
	{
		ID:          "oriz_the_bloodshot",
		Name:        "Oriz the Bloodshot",
		Pantheon:    "greek_offbrand",
		Domain:      "unwarranted_brawls_bruised_shins",
		RealmRoomID: "drafty_longhouse",
		Needs:       []string{"raw_meat", "bandages", "loud_shouting"},
		Shape:       "A hyperactive fighter whose rusty plate armor is missing several rivets. He swings an oversized iron sword but frequently trips over his own boots.",
		Attributes:  entity.Attributes{STR: 39, DEX: 29, CON: 32, INT: 14, WIS: 10, CHA: 16},
		WorshippedBy: []string{"human", "hobbit", "orc", "goblin", "ogre", "lizardfolk", "hobgoblin", "gnoll"},
	},
	{
		ID:          "snoozanoo",
		Name:        "Snoozano-o the Lethargic",
		Pantheon:    "japanese_offbrand",
		Domain:      "damp_gales_clogged_drains",
		RealmRoomID: "drafty_longhouse",
		Needs:       []string{"brackish_water", "dry_pillows", "salted_fish"},
		Shape:       "A bloated, sleeping sea-spirit resting face down in a puddle of muddy drainage water, snoring with the sound of a small, distant thunderstorm.",
		Attributes:  entity.Attributes{STR: 35, DEX: 27, CON: 32, INT: 19, WIS: 15, CHA: 25},
		WorshippedBy: []string{"orc", "goblin", "ogre", "lizardfolk", "hobgoblin", "gnoll"},
	},

	// Room 4: moldering_root_vault — greek + dnd + japanese
	{
		ID:          "haydes_the_hoarder",
		Name:        "Haydes the Hoarder",
		Pantheon:    "greek_offbrand",
		Domain:      "cluttered_graves_unpaid_debts",
		RealmRoomID: "moldering_root_vault",
		Needs:       []string{"zinc_coins", "rusty_nails", "cold_gruel"},
		Shape:       "A hunched figure wrapped in canvas potato sacks, sitting on piles of worthless iron coins and sorting through dead people's loose shoes.",
		Attributes:  entity.Attributes{STR: 27, DEX: 18, CON: 32, INT: 31, WIS: 35, CHA: 25},
		WorshippedBy: []string{"dwarf"},
	},
	{
		ID:          "tie_o_mat",
		Name:        "Tie-O-Mat the Flatulent",
		Pantheon:    "dnd_offbrand",
		Domain:      "hoarded_copper_dungeon_odors",
		RealmRoomID: "moldering_root_vault",
		Needs:       []string{"copper_scraps", "rotten_meat", "charcoal"},
		Shape:       "A bloated, lizard-like entity with five small reptile heads that spend all their time biting each other over single scraps of copper wire.",
		Attributes:  entity.Attributes{STR: 41, DEX: 22, CON: 41, INT: 26, WIS: 25, CHA: 31},
		WorshippedBy: []string{"dwarf"},
	},
	{
		ID:          "raijin_the_rattler",
		Name:        "Raijin the Rattler",
		Pantheon:    "japanese_offbrand",
		Domain:      "cracked_drums_tinnitus",
		RealmRoomID: "moldering_root_vault",
		Needs:       []string{"raw_hide", "wooden_mallets", "earplugs"},
		Shape:       "An old oni whose thunder-drums have loose, dry skins. Every time he hits them, they make a sad, hollow rattling sound instead of thunder.",
		Attributes:  entity.Attributes{STR: 29, DEX: 25, CON: 27, INT: 15, WIS: 19, CHA: 22},
		WorshippedBy: []string{"dwarf"},
	},

	// Room 5: the_cluttered_attic — norse + chinese + dnd
	{
		ID:          "odd_in",
		Name:        "Odd-In the Near-Sighted",
		Pantheon:    "norse_offbrand",
		Domain:      "conspiracy_theories_damp_scrolls",
		RealmRoomID: "the_cluttered_attic",
		Needs:       []string{"secret_journals", "old_cheese", "earplugs"},
		Shape:       "A paranoid patriarch squinting through a cracked leather eye-patch. He has two mangy ravens that keep biting his ears instead of giving secrets.",
		Attributes:  entity.Attributes{STR: 29, DEX: 22, CON: 32, INT: 39, WIS: 35, CHA: 28},
		WorshippedBy: []string{"dwarf"},
	},
	{
		ID:          "wukong_the_mangy",
		Name:        "Wukong the Mangy",
		Pantheon:    "chinese_offbrand",
		Domain:      "unwarranted_confidence_bruised_shins",
		RealmRoomID: "the_cluttered_attic",
		Needs:       []string{"stolen_peaches", "flea_combs", "stale_bread"},
		Shape:       "An old baboon wearing an oversized bronze helmet that continuously slips down over his eyes, causing him to hit walls with his quarterstaff.",
		Attributes:  entity.Attributes{STR: 35, DEX: 41, CON: 31, INT: 26, WIS: 16, CHA: 31},
		WorshippedBy: []string{"orc", "goblin", "ogre", "lizardfolk", "hobgoblin", "gnoll"},
	},
	{
		ID:          "vaicna_the_unwashed",
		Name:        "Vaicna the Unwashed",
		Pantheon:    "dnd_offbrand",
		Domain:      "petty_secrets_mildew_scrolls",
		RealmRoomID: "the_cluttered_attic",
		Needs:       []string{"stale_bread", "secret_journals", "moldy_cheese"},
		Shape:       "A gaunt, skeletal lich missing its left earlobe and hand, wearing a moth-eaten woolen sweater and hiding moldy scrolls inside its ribs.",
		Attributes:  entity.Attributes{STR: 19, DEX: 18, CON: 25, INT: 41, WIS: 36, CHA: 26},
		WorshippedBy: []string{"orc", "goblin", "ogre", "lizardfolk", "hobgoblin", "gnoll"},
	},

	// Room 6: the_sump_tank — greek + chinese + dnd
	{
		ID:          "posse_eidon",
		Name:        "Posse-Eidon the Silt-King",
		Pantheon:    "greek_offbrand",
		Domain:      "stagnant_puddles_well_collapses",
		RealmRoomID: "the_sump_tank",
		Needs:       []string{"brackish_water", "dead_frogs", "salt_pork"},
		Shape:       "A shivering, wet humanoid covered in pond scum and smelling of stagnant bogs, holding a bent iron fishing trident he uses as a walking cane.",
		Attributes:  entity.Attributes{STR: 35, DEX: 20, CON: 35, INT: 21, WIS: 20, CHA: 24},
		WorshippedBy: []string{"human", "hobbit"},
	},
	{
		ID:          "ooh_huang",
		Name:        "Ooh-Huang the Clerk",
		Pantheon:    "chinese_offbrand",
		Domain:      "divine_red_tape_stagnation",
		RealmRoomID: "the_sump_tank",
		Needs:       []string{"stamped_deeds", "dried_ink", "absolute_silence"},
		Shape:       "An ancient, unblinking celestial bureaucrat whose fingers are made entirely of brittle, yellowed audit parchment rolls.",
		Attributes:  entity.Attributes{STR: 31, DEX: 22, CON: 36, INT: 41, WIS: 41, CHA: 36},
		WorshippedBy: []string{"human", "hobbit"},
	},
	{
		ID:          "baa_hamut",
		Name:        "Baa-Hamut the Blunderer",
		Pantheon:    "dnd_offbrand",
		Domain:      "misplaced_justice_broken_scales",
		RealmRoomID: "the_sump_tank",
		Needs:       []string{"polishing_wax", "parchment_deeds", "herbal_tea"},
		Shape:       "A massive platinum-scaled creature whose eyesight is so poor he continuously knocks down ancient marble columns while looking for his glasses.",
		Attributes:  entity.Attributes{STR: 38, DEX: 22, CON: 38, INT: 31, WIS: 34, CHA: 34},
		WorshippedBy: []string{"elf"},
	},
}

func findRealmForDeity(deityID string) string {
	for _, rd := range RealmDefs {
		for _, did := range rd.DeityIDs {
			if did == deityID {
				return rd.ID
			}
		}
	}
	log.Printf("[gen] findRealmForDeity: no realm found for deity %s", deityID)
	return ""
}

var deityIDs []string

// speciesWorshipPool is built at init() from DeityDefs.WorshippedBy tags.
// Maps species → list of deity IDs that species can worship.
var speciesWorshipPool = map[string][]string{}

// speciesWorshipRange defines min/max deity count per species at seeding time.
var speciesWorshipRange = map[string][2]int{
	"elf":      {1, 2},
	"dwarf":    {1, 2},
	"human":    {1, 3},
	"hobbit":   {1, 3},
	"orc":      {0, 2},
	"goblin":   {0, 2},
	"ogre":     {0, 2},
	"lizardfolk": {0, 2},
	"hobgoblin": {0, 2},
	"gnoll":    {0, 2},
}

func init() {
	for _, d := range DeityDefs {
		deityIDs = append(deityIDs, d.ID)
		for _, sp := range d.WorshippedBy {
			speciesWorshipPool[sp] = append(speciesWorshipPool[sp], d.ID)
		}
	}
}

// AssignWorship sets the Worship and Cause fields on an entity.
// The pool is built from DeityDefs.WorshippedBy tags at init time,
// so adding a new deity only requires tagging it in the definition.
func AssignWorship(e *entity.Entity, rng interface{ Intn(int) int }) {
	if e.Species == "divine" || e.Species == "deity" {
		return
	}

	pool := speciesWorshipPool[e.Species]
	if len(pool) == 0 {
		pool = deityIDs
	}

	rngRange := speciesWorshipRange[e.Species]
	minCount, maxCount := rngRange[0], rngRange[1]
	if maxCount == 0 {
		minCount, maxCount = 0, 1
	}
	count := minCount
	if maxCount > minCount {
		count += rng.Intn(maxCount - minCount)
	}

	if count == 0 || len(pool) == 0 {
		return
	}

	picked := make(map[string]bool)
	for len(picked) < count {
		d := pool[rng.Intn(len(pool))]
		if !picked[d] {
			picked[d] = true
			e.Worship = append(e.Worship, d)
		}
	}

	if len(e.Worship) > 0 && e.Cause == "" {
		for _, dd := range DeityDefs {
			if dd.ID == e.Worship[0] {
				e.Cause = dd.Domain
				break
			}
		}
	}
}

func equipDeity(e *entity.Entity, id string) {
	switch id {
	case "zeus":
		log.Printf("[gen] equipDeity: %s -> toga, lightning_bolt, aegis_shield", id)
		equipItem(e, lookup("toga"))
		equipItem(e, lookup("lightning_bolt"))
		equipItem(e, lookup("aegis_shield"))
	case "hades":
		log.Printf("[gen] equipDeity: %s -> dark_robe, bident, helm_of_darkness", id)
		equipItem(e, lookup("dark_robe"))
		equipItem(e, lookup("bident"))
		equipItem(e, lookup("helm_of_darkness"))
	case "poseidon":
		log.Printf("[gen] equipDeity: %s -> toga, trident", id)
		equipItem(e, lookup("toga"))
		equipItem(e, lookup("trident"))
	case "athena":
		log.Printf("[gen] equipDeity: %s -> scale_armor, iron_helmet, aegis_shield, iron_spear", id)
		equipItem(e, lookup("scale_armor"))
		equipItem(e, lookup("iron_helmet"))
		equipItem(e, lookup("aegis_shield"))
		equipItem(e, lookup("iron_spear"))
	case "ares":
		log.Printf("[gen] equipDeity: %s -> chainmail, iron_helmet, iron_sword, iron_shield, iron_boots", id)
		equipItem(e, lookup("chainmail"))
		equipItem(e, lookup("iron_helmet"))
		equipItem(e, lookup("iron_sword"))
		equipItem(e, lookup("iron_shield"))
		equipItem(e, lookup("iron_boots"))
	case "odin":
		log.Printf("[gen] equipDeity: %s -> fine_clothes, gungnir", id)
		equipItem(e, lookup("fine_clothes"))
		equipItem(e, lookup("gungnir"))
	case "thor":
		log.Printf("[gen] equipDeity: %s -> megingjord, iron_gauntlets, mjolnir", id)
		equipItem(e, lookup("megingjord"))
		equipItem(e, lookup("iron_gauntlets"))
		equipItem(e, lookup("mjolnir"))
	case "loki":
		log.Printf("[gen] equipDeity: %s -> fine_clothes", id)
		equipItem(e, lookup("fine_clothes"))
	case "freya":
		log.Printf("[gen] equipDeity: %s -> feather_cloak, brisingamen", id)
		equipItem(e, lookup("feather_cloak"))
		equipItem(e, lookup("brisingamen"))
	case "kukulkan":
		log.Printf("[gen] equipDeity: %s -> feather_cloak, golden_circlet", id)
		equipItem(e, lookup("feather_cloak"))
		equipItem(e, lookup("golden_circlet"))
	case "chaac":
		log.Printf("[gen] equipDeity: %s -> golden_circlet, iron_axe", id)
		equipItem(e, lookup("golden_circlet"))
		equipItem(e, lookup("iron_axe"))
	case "yu_huang":
		log.Printf("[gen] equipDeity: %s -> imperial_robe, jade_scepter", id)
		equipItem(e, lookup("imperial_robe"))
		equipItem(e, lookup("jade_scepter"))
	case "guan_yin":
		log.Printf("[gen] equipDeity: %s -> imperial_robe, willow_branch", id)
		equipItem(e, lookup("imperial_robe"))
		equipItem(e, lookup("willow_branch"))
	case "sun_wukong":
		log.Printf("[gen] equipDeity: %s -> golden_circlet, ruyi_bang, leather_boots", id)
		equipItem(e, lookup("golden_circlet"))
		equipItem(e, lookup("ruyi_bang"))
		equipItem(e, lookup("leather_boots"))
	case "amaterasu":
		log.Printf("[gen] equipDeity: %s -> imperial_robe, jade_scepter", id)
		equipItem(e, lookup("imperial_robe"))
		equipItem(e,lookup("jade_scepter"))
	case "susanoo":
		log.Printf("[gen] equipDeity: %s -> common_clothes, kusanagi", id)
		equipItem(e, lookup("common_clothes"))
		equipItem(e, lookup("kusanagi"))
	case "raijin":
		log.Printf("[gen] equipDeity: %s -> common_clothes, raijin_drums", id)
		equipItem(e, lookup("common_clothes"))
		equipItem(e, lookup("raijin_drums"))
	case "tiamat":
		log.Printf("[gen] equipDeity: %s -> dragon_crown", id)
		equipItem(e, lookup("dragon_crown"))
	case "bahamut":
		log.Printf("[gen] equipDeity: %s -> dragon_crown", id)
		equipItem(e, lookup("dragon_crown"))
	case "vecna":
		log.Printf("[gen] equipDeity: %s -> dark_robe, eye_of_vecna, hand_of_vecna", id)
		equipItem(e, lookup("dark_robe"))
		equipItem(e, lookup("eye_of_vecna"))
		equipItem(e, lookup("hand_of_vecna"))
	default:
		log.Printf("[gen] equipDeity: unknown deity id %s, no equipment assigned", id)
	}
}
