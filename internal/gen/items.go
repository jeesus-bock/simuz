// Package gen contains world generation helpers and seeded simulation setup utilities.
package gen

import (
	"fmt"
	"log"

	"simuz/internal/entity"
	"simuz/internal/items"
)

func defineItem(id, name string, typ items.ItemType, weight float64, value int, slot string) *items.ItemDef {
	return &items.ItemDef{
		ID:     id,
		Name:   name,
		Type:   typ,
		Weight: weight,
		Value:  value,
		Slot:   slot,
	}
}

func defineCurrency(id, name string, value int) *items.ItemDef {
	return defineItem(id, name, items.TypeCurrency, 0.01, value, "")
}

var ItemDefs = []*items.ItemDef{
	defineCurrency("cp", "Copper Piece", 1),
	defineCurrency("sp", "Silver Piece", 10),
	defineCurrency("gp", "Gold Piece", 100),
	defineCurrency("tp", "Titanium Piece", 1000),
	defineCurrency("mp", "Mithril Piece", 10000),
	defineCurrency("ep", "Electronium Piece", 100000),

	defineItem("simple_robe", "Simple Robe", items.TypeArmor, 1, 5, "body"),
	defineItem("work_tunic", "Work Tunic", items.TypeArmor, 1.5, 3, "body"),
	defineItem("common_clothes", "Common Clothes", items.TypeArmor, 1, 2, "body"),
	defineItem("priest_robe", "Priest's Vestments", items.TypeArmor, 1.5, 15, "body"),
	defineItem("fine_clothes", "Fine Clothes", items.TypeArmor, 1, 20, "body"),
	defineItem("leather_armor", "Leather Armor", items.TypeArmor, 7, 50, "body"),
	defineItem("scale_armor", "Scale Armor", items.TypeArmor, 15, 150, "body"),
	defineItem("chainmail", "Chainmail", items.TypeArmor, 20, 200, "body"),
	defineItem("iron_helmet", "Iron Helmet", items.TypeArmor, 2, 30, "head"),
	defineItem("leather_helmet", "Leather Cap", items.TypeArmor, 0.5, 8, "head"),
	defineItem("iron_boots", "Iron Boots", items.TypeArmor, 3, 25, "feet"),
	defineItem("leather_boots", "Leather Boots", items.TypeArmor, 1, 10, "feet"),
	defineItem("wooden_shield", "Wooden Shield", items.TypeArmor, 3, 15, "offhand"),
	defineItem("iron_shield", "Iron Shield", items.TypeArmor, 7, 50, "offhand"),
	defineItem("leather_gloves", "Leather Gloves", items.TypeArmor, 0.3, 5, "hands"),

	defineItem("iron_sword", "Iron Sword", items.TypeWeapon, 3, 100, "weapon"),
	defineItem("iron_spear", "Iron Spear", items.TypeWeapon, 3.5, 80, "weapon"),
	defineItem("iron_axe", "Iron Axe", items.TypeWeapon, 4, 90, "weapon"),
	defineItem("short_sword", "Short Sword", items.TypeWeapon, 2, 60, "weapon"),
	defineItem("dagger", "Dagger", items.TypeWeapon, 0.5, 25, "weapon"),
	defineItem("cudgel", "Cudgel", items.TypeWeapon, 2, 5, "weapon"),
	defineItem("goblin_shiv", "Goblin Shiv", items.TypeWeapon, 0.4, 8, "weapon"),
	defineItem("orc_cleaver", "Orc Cleaver", items.TypeWeapon, 5, 70, "weapon"),
	defineItem("claws", "Claws", items.TypeWeapon, 0.1, 1, "weapon"),
	defineItem("fangs", "Fangs", items.TypeWeapon, 0.1, 1, "weapon"),
	defineItem("tusks", "Tusks", items.TypeWeapon, 0.2, 1, "weapon"),
	defineItem("holy_symbol", "Holy Symbol", items.TypeMisc, 0.2, 30, "neck"),
	defineItem("tankard", "Tankard", items.TypeMisc, 0.3, 1, ""),
	defineItem("smith_hammer", "Smith's Hammer", items.TypeWeapon, 2, 15, "weapon"),

	defineItem("lightning_bolt", "Lightning Bolt", items.TypeWeapon, 0.5, 15000, "weapon"),
	defineItem("bident", "Bident", items.TypeWeapon, 4, 12000, "weapon"),
	defineItem("trident", "Trident", items.TypeWeapon, 4, 12000, "weapon"),
	defineItem("gungnir", "Gungnir", items.TypeWeapon, 3, 25000, "weapon"),
	defineItem("mjolnir", "Mjölnir", items.TypeWeapon, 8, 30000, "weapon"),
	defineItem("ruyi_bang", "Ruyi Jingu Bang", items.TypeWeapon, 10, 35000, "weapon"),
	defineItem("aegis_shield", "Aegis Shield", items.TypeArmor, 3, 10000, "offhand"),
	defineItem("golden_circlet", "Golden Circlet", items.TypeArmor, 0.1, 8000, "head"),
	defineItem("iron_gauntlets", "Iron Gauntlets", items.TypeArmor, 1, 5000, "hands"),
	defineItem("megingjord", "Megingjörð", items.TypeArmor, 0.5, 18000, "body"),
	defineItem("feather_cloak", "Feather Cloak", items.TypeArmor, 0.5, 12000, "body"),
	defineItem("imperial_robe", "Imperial Robe", items.TypeArmor, 2, 15000, "body"),
	defineItem("jade_scepter", "Jade Scepter", items.TypeWeapon, 1, 20000, "weapon"),
	defineItem("willow_branch", "Willow Branch", items.TypeMisc, 0.2, 9000, ""),
	defineItem("kusanagi", "Kusanagi-no-Tsurugi", items.TypeWeapon, 3, 28000, "weapon"),
	defineItem("raijin_drums", "Raijin Drums", items.TypeMisc, 5, 15000, ""),
	defineItem("helm_of_darkness", "Helm of Darkness", items.TypeArmor, 1, 14000, "head"),
	defineItem("brisingamen", "Brísingamen", items.TypeArmor, 0.3, 16000, "neck"),
	defineItem("toga", "Toga", items.TypeArmor, 0.8, 5000, "body"),
	defineItem("dark_robe", "Dark Robe", items.TypeArmor, 1, 8000, "body"),
	defineItem("dragon_crown", "Dragon Crown", items.TypeArmor, 2, 22000, "head"),
	defineSubstance("fish_trout", "Trout", 3, 0.3,
		&items.SubstanceEffect{Name: "Sustenance", HealHP: 5}),
	defineSubstance("fish_salmon", "Salmon", 8, 0.5,
		&items.SubstanceEffect{Name: "Sustenance", HealHP: 10}),
	defineSubstance("fish_catfish", "Catfish", 12, 0.8,
		&items.SubstanceEffect{Name: "Sustenance", HealHP: 10, HealFP: 5}),

	defineItem("rat_fang", "Rat Fang", items.TypeMaterial, 0.1, 2, ""),
	defineItem("rat_king_crown", "Rat King Crown", items.TypeArmor, 0.5, 500, "head"),
	defineItem("lute", "Lute", items.TypeMisc, 2, 25, ""),
	defineItem("flute", "Flute", items.TypeMisc, 0.5, 15, ""),

	defineItem("fishing_rod", "Fishing Rod", items.TypeMisc, 1.5, 25, ""),
	defineItem("bait", "Bait", items.TypeConsumable, 0.1, 5, ""),

	defineItem("eye_of_vecna", "Eye of Vecna", items.TypeMisc, 0.1, 50000, "neck"),
	defineItem("hand_of_vecna", "Hand of Vecna", items.TypeMisc, 0.5, 50000, "hands"),

	defineSubstance("beer", "Beer", 3, 5,
		&items.SubstanceEffect{Name: "Drunk (Beer)", Duration: 20, CrashDuration: 10,
			STR: 1, CON: 1, INT: -2, WIS: -2, CHA: -2,
			CrashSTR: -1, CrashCON: -1}),
	defineSubstance("wine", "Wine", 5, 15,
		&items.SubstanceEffect{Name: "Drunk (Wine)", Duration: 25, CrashDuration: 12,
			STR: 2, CON: 2, INT: -3, WIS: -3, CHA: -3,
			CrashSTR: -1, CrashCON: -1, CrashINT: -1, CrashWIS: -1}),
	defineSubstance("liquor", "Liquor", 5, 25,
		&items.SubstanceEffect{Name: "Drunk (Liquor)", Duration: 30, CrashDuration: 15,
			STR: 3, CON: 3, INT: -4, WIS: -4, CHA: -4,
			CrashSTR: -2, CrashCON: -2, CrashINT: -1, CrashWIS: -1}),
	defineSubstance("mead", "Mead", 4, 8,
		&items.SubstanceEffect{Name: "Drunk (Mead)", Duration: 22, CrashDuration: 11,
			STR: 2, CON: 1, INT: -1, WIS: -2, CHA: -1,
			CrashSTR: -1, CrashCON: -1}),
	defineSubstance("brandy", "Brandy", 8, 18,
		&items.SubstanceEffect{Name: "Drunk (Brandy)", Duration: 28, CrashDuration: 14,
			STR: 1, CON: 1, INT: -3, WIS: -3, CHA: 2,
			CrashSTR: -1, CrashCON: -1, CrashCHA: -1}),
	defineSubstance("ale", "Ale", 2, 4,
		&items.SubstanceEffect{Name: "Drunk (Ale)", Duration: 18, CrashDuration: 9,
			STR: 1, CON: 1, INT: -1, WIS: -1, CHA: -1,
			CrashSTR: -1}),
	defineSubstance("moonshine", "Moonshine", 10, 30,
		&items.SubstanceEffect{Name: "Drunk (Moonshine)", Duration: 35, CrashDuration: 18,
			STR: 4, CON: 3, INT: -5, WIS: -5, CHA: -5,
			CrashSTR: -2, CrashCON: -2, CrashINT: -2, CrashWIS: -2}),
	defineSubstance("mushroom_puff", "Mushroom Puff", 12, 60,
		&items.SubstanceEffect{Name: "Hallucinating", Duration: 35, CrashDuration: 17,
			WIS: 3, INT: 3, CON: -1,
			CrashINT: -2, CrashWIS: -2}),
	defineSubstance("firewater", "Firewater", 7, 35,
		&items.SubstanceEffect{Name: "Fired Up", Duration: 25, CrashDuration: 12,
			DEX: 3, STR: 2, CON: -1,
			CrashDEX: -2, CrashCON: -1}),
	defineSubstance("cactus_juice", "Cactus Juice", 6, 25,
		&items.SubstanceEffect{Name: "Prickly", Duration: 28, CrashDuration: 14,
			DEX: 4, CON: -1,
			CrashINT: -2, CrashWIS: -2}),
	defineSubstance("shadowroot", "Shadowroot", 9, 45,
		&items.SubstanceEffect{Name: "Shadow-Touched", Duration: 32, CrashDuration: 16,
			CHA: 3, WIS: 2, CON: -1,
			CrashCHA: -2}),
	defineSubstance("night_bloom", "Night Bloom", 10, 50,
		&items.SubstanceEffect{Name: "Stimulated", Duration: 30, CrashDuration: 15,
			DEX: 4, CON: -1,
			CrashDEX: -3}),
	defineSubstance("sage_leaf", "Sage Leaf", 8, 40,
		&items.SubstanceEffect{Name: "Focused", Duration: 30, CrashDuration: 15,
			INT: 4, WIS: 4,
			CrashINT: -3, CrashWIS: -3}),
	defineSubstance("dreamers_cap", "Dreamer's Cap", 15, 150,
		&items.SubstanceEffect{Name: "Dreaming", Duration: 40, CrashDuration: 20,
			STR: 2, DEX: 2, CON: 2, INT: 6, WIS: 6,
			CrashSTR: -3, CrashDEX: -3, CrashCON: -3, CrashINT: -4, CrashWIS: -4}),

	defineSubstance("raw_chicken", "Raw Chicken", 5, 1.0,
		&items.SubstanceEffect{Name: "Sustenance", HealHP: 8}),
	defineSubstance("raw_pork", "Raw Pork", 8, 1.5,
		&items.SubstanceEffect{Name: "Sustenance", HealHP: 12}),
	defineSubstance("raw_beef", "Raw Beef", 15, 2.0,
		&items.SubstanceEffect{Name: "Sustenance", HealHP: 20}),
	defineSubstance("raw_mutton", "Raw Mutton", 10, 1.5,
		&items.SubstanceEffect{Name: "Sustenance", HealHP: 14}),
	defineSubstance("raw_goat", "Raw Goat Meat", 10, 1.5,
		&items.SubstanceEffect{Name: "Sustenance", HealHP: 14}),
	defineSubstance("egg", "Egg", 1, 0.1,
		&items.SubstanceEffect{Name: "Sustenance", HealHP: 2}),
	defineSubstance("milk", "Milk", 2, 0.5,
		&items.SubstanceEffect{Name: "Sustenance", HealHP: 4}),
	defineItem("wool", "Wool", items.TypeMaterial, 0.5, 5, ""),
	defineItem("feather", "Feather", items.TypeMaterial, 0.05, 1, ""),
	defineItem("leather", "Leather", items.TypeMaterial, 1.0, 8, ""),
	defineItem("grain", "Grain", items.TypeConsumable, 0.5, 1, ""),
	defineItem("trap_kit", "Trap Kit", items.TypeConsumable, 2, 30, ""),
	defineItem("sleep_dust", "Sleep Dust", items.TypeConsumable, 0.1, 40, ""),
	defineItem("memory_shard", "Memory Shard", items.TypeMisc, 0.2, 60, ""),
	defineItem("vampire_fang", "Vampire Fang", items.TypeWeapon, 0.3, 200, "weapon"),
	defineItem("pickaxe", "Pickaxe", items.TypeMisc, 3, 15, ""),
	defineItem("herb_pouch", "Herb Pouch", items.TypeMisc, 0.5, 5, ""),
	defineItem("sickle", "Sickle", items.TypeWeapon, 2, 25, "weapon"),
	defineItem("cultist_dagger", "Cultist Dagger", items.TypeWeapon, 0.5, 20, "weapon"),
	defineItem("necromancer_staff", "Necromancer Staff", items.TypeWeapon, 2, 80, "weapon"),
	defineItem("dragon_scale", "Dragon Scale", items.TypeMaterial, 0.5, 250, ""),
	defineItem("iron_ore", "Iron Ore", items.TypeMaterial, 2.0, 3, ""),
	defineItem("coal", "Coal", items.TypeMaterial, 1.0, 2, ""),
	defineItem("cloth", "Cloth", items.TypeMaterial, 0.3, 4, ""),
	defineItem("leather_strips", "Leather Strips", items.TypeMaterial, 0.2, 3, ""),
	defineItem("herb", "Herb", items.TypeMaterial, 0.1, 6, ""),
	defineItem("iron_ingot", "Iron Ingot", items.TypeMaterial, 1.0, 10, ""),
	defineItem("copper_ingot", "Copper Ingot", items.TypeMaterial, 1.0, 8, ""),
	defineSubstance("bandage", "Bandage", 8, 0.2,
		&items.SubstanceEffect{Name: "Regeneration", Duration: 10, HealPerTick: 2}),
	defineSubstance("herbal_poultice", "Herbal Poultice", 15, 0.3,
		&items.SubstanceEffect{Name: "Regeneration", Duration: 15, HealPerTick: 3, HealHP: 2}),
	defineSubstance("healing_salve", "Healing Salve", 30, 0.4,
		&items.SubstanceEffect{Name: "Regeneration", Duration: 10, HealPerTick: 5, HealHP: 5}),
}

func NewItemInstance(def *items.ItemDef, count int) items.ItemInstance {
	return items.NewItemInstance(def.ID+"_"+fmt.Sprint(count), def.ID, def, count)
}

func giveCurrency(e *entity.Entity, copper, silver, gold int) {
	if copper > 0 {
		e.AddItem(NewItemInstance(lookup("cp"), copper))
	}
	if silver > 0 {
		e.AddItem(NewItemInstance(lookup("sp"), silver))
	}
	if gold > 0 {
		e.AddItem(NewItemInstance(lookup("gp"), gold))
	}
}

func equipItem(e *entity.Entity, def *items.ItemDef) {
	if def == nil {
		return
	}
	inst := NewItemInstance(def, 1)
	e.AddItem(inst)
	if def.Slot != "" {
		e.Equip(&e.Inventory[len(e.Inventory)-1])
	}
}

func addItem(e *entity.Entity, def *items.ItemDef) {
	if def == nil {
		return
	}
	inst := NewItemInstance(def, 1)
	e.AddItem(inst)
}

func lookup(id string) *items.ItemDef {
	for _, d := range ItemDefs {
		if d.ID == id {
			return d
		}
	}
	return nil
}

func defineSubstance(id, name string, value int, weight float64, se *items.SubstanceEffect) *items.ItemDef {
	return &items.ItemDef{
		ID:        id,
		Name:      name,
		Type:      items.TypeConsumable,
		Weight:    weight,
		Value:     value,
		Slot:      "",
		Substance: se,
	}
}

func init() {
	log.Printf("[gen] items: registering %d item definitions", len(ItemDefs))
	for _, d := range ItemDefs {
		items.RegisterDef(d)
	}
}
