// Package gen contains world generation helpers and seeded simulation setup utilities.
package gen

import (
	"simuz/internal/entity"
	"simuz/internal/world"
)

type deityDef struct {
	ID         string
	Name       string
	Pantheon   string
	Domain     string
	Attributes entity.Attributes
	Active     bool
}

var deityDefs = []deityDef{
	{ID: "zeus", Name: "Zeus", Pantheon: "greek", Domain: "sky_thunder", Attributes: entity.Attributes{STR: 35, DEX: 25, CON: 40, INT: 30, WIS: 35, CHA: 40}},
	{ID: "hades", Name: "Hades", Pantheon: "greek", Domain: "underworld_death", Attributes: entity.Attributes{STR: 30, DEX: 20, CON: 35, INT: 35, WIS: 38, CHA: 30}},
	{ID: "poseidon", Name: "Poseidon", Pantheon: "greek", Domain: "sea_earthquakes", Attributes: entity.Attributes{STR: 38, DEX: 22, CON: 38, INT: 25, WIS: 28, CHA: 32}},
	{ID: "athena", Name: "Athena", Pantheon: "greek", Domain: "wisdom_war", Attributes: entity.Attributes{STR: 28, DEX: 30, CON: 28, INT: 40, WIS: 38, CHA: 35}},
	{ID: "ares", Name: "Ares", Pantheon: "greek", Domain: "war_bloodlust", Attributes: entity.Attributes{STR: 40, DEX: 32, CON: 35, INT: 18, WIS: 15, CHA: 22}},

	{ID: "odin", Name: "Odin", Pantheon: "norse", Domain: "wisdom_war_death", Attributes: entity.Attributes{STR: 32, DEX: 25, CON: 35, INT: 42, WIS: 40, CHA: 35}},
	{ID: "thor", Name: "Thor", Pantheon: "norse", Domain: "thunder_strength", Attributes: entity.Attributes{STR: 45, DEX: 28, CON: 42, INT: 16, WIS: 20, CHA: 28}},
	{ID: "loki", Name: "Loki", Pantheon: "norse", Domain: "trickery_chaos", Attributes: entity.Attributes{STR: 22, DEX: 38, CON: 25, INT: 38, WIS: 25, CHA: 40}},
	{ID: "freya", Name: "Freya", Pantheon: "norse", Domain: "love_beauty_war", Attributes: entity.Attributes{STR: 25, DEX: 30, CON: 28, INT: 30, WIS: 32, CHA: 45}},

	{ID: "kukulkan", Name: "Kukulkan", Pantheon: "mayan", Domain: "wind_knowledge", Attributes: entity.Attributes{STR: 30, DEX: 35, CON: 32, INT: 36, WIS: 38, CHA: 35}},
	{ID: "chaac", Name: "Chaac", Pantheon: "mayan", Domain: "rain_lightning", Attributes: entity.Attributes{STR: 36, DEX: 22, CON: 38, INT: 20, WIS: 28, CHA: 25}},

	{ID: "yu_huang", Name: "Yu Huang", Pantheon: "chinese", Domain: "heaven_order", Attributes: entity.Attributes{STR: 35, DEX: 25, CON: 40, INT: 45, WIS: 45, CHA: 42}},
	{ID: "guan_yin", Name: "Guan Yin", Pantheon: "chinese", Domain: "mercy_compassion", Attributes: entity.Attributes{STR: 18, DEX: 22, CON: 30, INT: 38, WIS: 45, CHA: 45}},
	{ID: "sun_wukong", Name: "Sun Wukong", Pantheon: "chinese", Domain: "trickster_strength", Attributes: entity.Attributes{STR: 38, DEX: 45, CON: 35, INT: 30, WIS: 20, CHA: 35}},

	{ID: "amaterasu", Name: "Amaterasu", Pantheon: "japanese", Domain: "sun_light", Attributes: entity.Attributes{STR: 25, DEX: 28, CON: 32, INT: 35, WIS: 40, CHA: 45}},
	{ID: "susanoo", Name: "Susano-o", Pantheon: "japanese", Domain: "storms_sea", Attributes: entity.Attributes{STR: 38, DEX: 30, CON: 35, INT: 22, WIS: 18, CHA: 28}},
	{ID: "raijin", Name: "Raijin", Pantheon: "japanese", Domain: "thunder_drums", Attributes: entity.Attributes{STR: 32, DEX: 28, CON: 30, INT: 18, WIS: 22, CHA: 26}},

	{ID: "tiamat", Name: "Tiamat", Pantheon: "dnd", Domain: "greed_dragons", Attributes: entity.Attributes{STR: 45, DEX: 25, CON: 45, INT: 30, WIS: 28, CHA: 35}},
	{ID: "bahamut", Name: "Bahamut", Pantheon: "dnd", Domain: "justice_dragons", Attributes: entity.Attributes{STR: 42, DEX: 25, CON: 42, INT: 35, WIS: 38, CHA: 38}},
	{ID: "vecna", Name: "Vecna", Pantheon: "dnd", Domain: "secrets_undeath", Attributes: entity.Attributes{STR: 22, DEX: 20, CON: 28, INT: 45, WIS: 40, CHA: 30}},
}

type realmDef struct {
	ID       string
	Name     string
	DeityIDs []string
}

var realmDefs = []realmDef{
	{
		ID: "olympus", Name: "Mount Olympus",
		DeityIDs: []string{"zeus", "hades", "poseidon", "athena", "ares"},
	},
	{
		ID: "asgard", Name: "Asgard",
		DeityIDs: []string{"odin", "thor", "loki", "freya"},
	},
	{
		ID: "tamoanchan", Name: "Tamoanchan",
		DeityIDs: []string{"kukulkan", "chaac"},
	},
	{
		ID: "celestial_court", Name: "Celestial Court",
		DeityIDs: []string{"yu_huang", "guan_yin", "sun_wukong", "amaterasu", "susanoo", "raijin"},
	},
	{
		ID: "outer_plane", Name: "The Outer Plane",
		DeityIDs: []string{"tiamat", "bahamut", "vecna"},
	},
}

func GenerateDeities(w *world.World) ([]*entity.Entity, []*world.Location) {
	realmMap := make(map[string]*world.Location)
	var locations []*world.Location

	for _, rd := range realmDefs {
		loc := world.NewLocation(rd.ID, rd.Name, world.LocRealm, "aetheria", world.Position{})
		loc.Tags = []string{"divine_realm"}
		w.AddLocation(loc)
		realmMap[rd.ID] = loc
		locations = append(locations, loc)
	}

	activeDeities := map[string]bool{
		"zeus":     true,
		"odin":     true,
		"yu_huang": true,
	}

	var entities []*entity.Entity
	for _, dd := range deityDefs {
		realmID := findRealmForDeity(dd.ID)
		attrs := dd.Attributes
		level := 50
		ent := entity.NewEntity(dd.ID, dd.Name, "deity", attrs, level, entity.Relation{})
		ent.Immortal = true
		ent.LocationID = realmID
		ent.Faction = "deity"
		ent.Flags["pantheon"] = dd.Pantheon
		ent.Flags["domain"] = dd.Domain

		if activeDeities[dd.ID] {
			ent.AI = entity.EntityAI{
				Type:         "scripted",
				ScriptIDs:    []string{"deity"},
				FactionID:    "deity",
				HomeLocation: realmID,
				SleepCycle:   "none",
			}
		} else {
			ent.AI = entity.EntityAI{
				Type:         "dormant",
				FactionID:    "deity",
				HomeLocation: realmID,
				SleepCycle:   "none",
			}
		}

		equipDeity(ent, dd.ID)

		entities = append(entities, ent)
	}

	return entities, locations
}

func findRealmForDeity(deityID string) string {
	for _, rd := range realmDefs {
		for _, did := range rd.DeityIDs {
			if did == deityID {
				return rd.ID
			}
		}
	}
	return ""
}

func equipDeity(e *entity.Entity, id string) {
	switch id {
	case "zeus":
		equipItem(e, lookup("toga"))
		equipItem(e, lookup("lightning_bolt"))
		equipItem(e, lookup("aegis_shield"))
	case "hades":
		equipItem(e, lookup("dark_robe"))
		equipItem(e, lookup("bident"))
		equipItem(e, lookup("helm_of_darkness"))
	case "poseidon":
		equipItem(e, lookup("toga"))
		equipItem(e, lookup("trident"))
	case "athena":
		equipItem(e, lookup("scale_armor"))
		equipItem(e, lookup("iron_helmet"))
		equipItem(e, lookup("aegis_shield"))
		equipItem(e, lookup("iron_spear"))
	case "ares":
		equipItem(e, lookup("chainmail"))
		equipItem(e, lookup("iron_helmet"))
		equipItem(e, lookup("iron_sword"))
		equipItem(e, lookup("iron_shield"))
		equipItem(e, lookup("iron_boots"))
	case "odin":
		equipItem(e, lookup("fine_clothes"))
		equipItem(e, lookup("gungnir"))
	case "thor":
		equipItem(e, lookup("megingjord"))
		equipItem(e, lookup("iron_gauntlets"))
		equipItem(e, lookup("mjolnir"))
	case "loki":
		equipItem(e, lookup("fine_clothes"))
	case "freya":
		equipItem(e, lookup("feather_cloak"))
		equipItem(e, lookup("brisingamen"))
	case "kukulkan":
		equipItem(e, lookup("feather_cloak"))
		equipItem(e, lookup("golden_circlet"))
	case "chaac":
		equipItem(e, lookup("golden_circlet"))
		equipItem(e, lookup("iron_axe"))
	case "yu_huang":
		equipItem(e, lookup("imperial_robe"))
		equipItem(e, lookup("jade_scepter"))
	case "guan_yin":
		equipItem(e, lookup("imperial_robe"))
		equipItem(e, lookup("willow_branch"))
	case "sun_wukong":
		equipItem(e, lookup("golden_circlet"))
		equipItem(e, lookup("ruyi_bang"))
		equipItem(e, lookup("leather_boots"))
	case "amaterasu":
		equipItem(e, lookup("imperial_robe"))
		equipItem(e, lookup("jade_scepter"))
	case "susanoo":
		equipItem(e, lookup("common_clothes"))
		equipItem(e, lookup("kusanagi"))
	case "raijin":
		equipItem(e, lookup("common_clothes"))
		equipItem(e, lookup("raijin_drums"))
	case "tiamat":
		equipItem(e, lookup("dragon_crown"))
	case "bahamut":
		equipItem(e, lookup("dragon_crown"))
	case "vecna":
		equipItem(e, lookup("dark_robe"))
		equipItem(e, lookup("eye_of_vecna"))
		equipItem(e, lookup("hand_of_vecna"))
	}
}
