package entity

var RaceGenderNames = map[string]map[string][]string{
	"elves": {
		"male": {
			"Leoleht",    // Poetic, sounds like an elegant leaf whisperer.
			"Urmasvalg",  // Urmas, but glowing with celestial light.
			"Sinitäht",   // "Blue star" - excessively majestic.
			"Kuuvarjund", // "Moon shade" - deep, brooding, dramatic.
			"Tuulepesa",  // "Wind nest" - messy hair but somehow elegant.
			"Kuldvibu",   // "Golden bow" - spent all his gold on a bow he can't aim.
			"Hõbejuus",   // "Silver hair" - standard classic elf cliché.
			"Aasalaul",   // "Meadow song" - exclusively talks about grass.
			"Vihmatants", // "Rain dance" - refuses to fight if it drizzles.
			"Valguse",    // "Of Light" - glows so bright no one can sleep.
			"Kastepiisk", // "Dewdrop" - overly delicate, breaks things.
			"Smaragdi",   // Named after emeralds, hoards green things.
			"Tähesära",   // "Starlight" - uses it as an excuse for bad eyesight.
			"Pilvepiir",  // "Cloud horizon" - completely detached from reality.
			"Ojahein",    // "Brook grass" - smells slightly like wet moss.
		},
		"female": {
			"Ilmatar",    // Traditional air spirit, acts incredibly snobbish.
			"Kastepärl",  // "Dew-pearl" - rescues the party, complains about mud.
			"Päikesekiir",// "Sunbeam" - blinds enemies with aggressive optimism.
			"Uduloor",    // "Fog veil" - consistently disappears during chores.
			"Ööbiku",     // "Nightingale" - sings instead of talking, very dramatic.
			"Sireli",     // "Lilac" - smells great, uses it to mask dungeon odors.
			"Lilleleht",  // "Flower leaf" - gracefully dodges axes.
			"Helendus",   // "Glow" - reads maps at night without a torch.
			"Kuldjuus",   // "Golden hair" - takes three hours to brush it.
			"Hingelind",  // "Soul bird" - provides profound, useless advice.
			"Metsapiiga", // "Forest maiden" - speaks fluent squirrel, ignores humans.
			"Allikavesi", // "Spring water" - completely pure, judges everyone's lifestyle.
			"Pilvetee",   // "Cloud path" - pathfinding skills are too abstract.
			"Õiesilm",    // "Blossom eye" - spots treasure from a mile away.
			"Härmatis",   // "Hoarfrost" - looks stunning, has an icy attitude.
		},
	},
	"orcs": {
		"male": {
			"Mürsk",      // "Artillery shell" - just charges into stone walls.
			"Kolge",      // Sounds like a heavy thud, probably lacks front teeth.
			"Raudrusikas",// "Iron fist" - forgets how to open doors gently.
			"Põmm",       // Literally "Bang" - his only strategy.
			"Kivirind",   // "Stone chest" - uses no armor, gets hurt constantly.
			"Porinägu",   // "Mud face" - trips over during battles.
			"Räme",       // "Gross/Harsh" - clear indicator of hygiene standards.
			"Kolv",       // "Piston" - moves fast, stops abruptly.
			"Mudaauk",    // "Mud hole" - prefers sleeping in wet ditches.
			"Kirvespea",  // "Axe head" - very stubborn, skull shaped like a weapon.
			"Murru",      // "Breaker" - accidentally breaks camp furniture.
			"Rusk",       // Sounds like a broken bone sound effect.
			"Känd",       // "Tree stump" - slow, wide, and refuses to move.
			"Tont",       // "Bogeyman" - trying his best to look scary.
			"Päts",       // "Loaf" - thick-headed and soft in the middle.
		},
		"female": {
			"Mürina",     // "Thunderous" - yells the battle plans, alerts enemies.
			"Raudmeel",   // "Iron mind" - handles the budget, terrifies the male orcs.
			"Kiviraev",   // "Stone fury" - literally throws boulders when annoyed.
			"Rabakoll",   // "Swamp hag" - highly territorial, cooks suspicious stews.
			"Poritallaja",// "Mud trampler" - doesn't stop for anyone, walks straight through.
			"Raskesaabas",// "Heavy boot" - you can hear her coming from a mile away.
			"Kandiline",  // "Square" - perfect, sturdy defensive build.
			"Tormi",      // "Storm" - a whirlwind of flying axes and bad attitude.
			"Teravnuga",  // "Sharp knife" - handles the skinning, never blunts her blade.
			"Karuema",    // "Mother bear" - protects the party, will slap you if you slack.
			"Kandamil",   // "The Burden" - carries all the loot without breaking a sweat.
			"Sõjatrumm",  // "War drum" - loud voice, keeps everyone in a rhythm.
			"Tahmasilm",  // "Soot eye" - looks ferocious, excellent night vision.
			"Raudnõges",  // "Iron nettle" - prickly to touch, stings her enemies.
			"Põlendik",   // "Burnt clearing" - survived a dragon, left with a cool scar.
		},
	},
	"fays": {
		"male": {
			"Sirts",      // Small, loud, and highly annoying.
			"Päkapikk",   // "Gnome" - identity crisis locked in a tiny body.
			"Kihulane",   // "Midge/Gnat" - bites your ankles and vanishes.
			"Säde",       // "Spark" - sets camp gear on fire by accident.
			"Putukas",    // Just "Bug" - treats human-scale things with suspicion.
			"Okas",       // "Thorn" - small, sharp, hides in people's boots.
			"Tirtsu",     // "Grasshopper" - jumps into danger headfirst.
			"Kärbseseen", // "Fly agaric" - colorful but highly toxic.
			"Virvendus",  // "Glimmer" - blurry, hard to look at directly.
			"Pilveke",    // "Little cloud" - hovers slightly above your head.
			"Tuulispask", // "Whirlwind" - zooms around, knocks over dynamic maps.
			"Leheke",     // "Little leaf" - floats away if it gets too windy.
			"Puruvana",   // "Dusty old guy" - tiny, grumpy, collects lint.
			"Kollake",    // "Little yellow" - bright, easily spotted by predators.
			"Välek",      // "Flashy" - blinks continuously, causes headaches.
		},
		"female": {
			"Marjake",    // "Little berry" - looks sweet, secretly paralyzes enemies.
			"Liblikas",   // "Butterfly" - attention span of exactly three seconds.
			"Mesilane",   // "Bee" - constantly busy, stings if provoked.
			"Lepatriinu", // "Ladybug" - brings good luck, mostly to herself.
			"Kastepiiga", // "Dew maiden" - wakes up early to prank the dwarves.
			"Õietolm",    // "Pollen" - causes the enemies to sneeze uncontrollably.
			"Nõgese",     // "Nettle" - fierce attitude, leaves a rash.
			"Kullake",    // "Darling" - steals your gold coins while smiling.
			"Pärlake",    // "Little pearl" - tiny, shiny, impossible to find if dropped.
			"Haldjatants",// "Fairy dance" - distracts the guards with flashing lights.
			"Käopoeg",    // "Cuckoo chick" - sneaks her problems into your inventory.
			"Lilleke",    // "Little flower" - cute exterior, lethal magic.
			"Sinitihane", // "Blue tit (bird)" - chirps warnings, never stops talking.
			"Kirgas",     // "Bright/Vivid" - radiates light, ruins stealth missions.
			"Tuuleiil",   // "Gust of wind" - steals hats and blows away loose papers.
		},
	},
	"dwarves": {
		"male": {
			"Mägra",      // "Badger" - lives underground, angry about daylight.
			"Kullakang",  // "Gold bar" - keeps his savings hidden in his beard.
			"Sepapoiss",  // "Blacksmith boy" - 150 years old, still an apprentice.
			"Kirka",      // "Pickaxe" - essentially married to his mining tool.
			"Suurkivi",   // "Big stone" - refuses to step aside for anyone.
			"Õllevaat",   // "Beer barrel" - build shape matches his favorite beverage.
			"Habeme",     // "Beard" - built his entire identity around facial hair.
			"Kaevur",     // "Miner" - hasn't seen the sun since the First Age.
			"Raudsaabas", // "Iron boot" - steps on everyone's toes on purpose.
			"Söetolm",    // "Coal dust" - permanently covered in black soot.
			"Graniit",    // "Granite" - literal blockhead, doesn't understand jokes.
			"Tõrvik",     // "Torch" - dropped it down the mine shaft once.
			"Kamm",       // "Comb" - ironic, beard is a complete rat's nest.
			"Vaskne",     // "Coppery" - smells strongly of old loose change.
			"Paas",       // "Limestone" - crumbling gray hair, smells like dust.
		},
		"female": {
			"Malm",       // "Cast Iron" - unyielding, indestructible, logical leader.
			"Kalliskivi", // "Gemstone" - rare, sharp, and cuts through any argument.
			"Ahjusoe",    // "Oven-warm" - bakes the emergency bread, keeps spirits high.
			"Raudvara",   // "Iron reserve" - manages the logistics, never runs out of ammo.
			"Kullasoon",  // "Gold vein" - has an absolute nose for hidden treasure.
			"Keldri",     // "Of the cellar" - knows exactly where the best ale is hidden.
			"Pliit",      // "Stove" - warm personality, but burns you if you are careless.
			"Kõvakivi",   // "Hard stone" - completely immune to psychological warfare.
			"Sepapiiga",  // "Smith maiden" - swings a sledgehammer better than her brother.
			"Alasi",      // "Anvil" - takes the heavy hits so the party doesn't have to.
			"Pätsike",    // "Little loaf" - looks round and soft, secretly dense as a brick.
			"Hõbevalge",  // "Silver-white" - prestigious keeper of ancestral clan lore.
			"Mõõdulint",  // "Measuring tape" - perfectionist, hates misaligned masonry.
			"Kirkasilm",  // "Bright/Pickaxe eye" - notices structural weak points instantly.
			"Väävel",     // "Sulfur" - explosive temper when the males do something stupid.
		},
	},
}

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