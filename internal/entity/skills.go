// Package entity defines the simulation entities, their attributes, and related behaviors.
package entity

const (
	SkillXPPerLevel = 100
)

var SkillNames = []string{
	"swords", "axes", "spears", "daggers", "unarmed",
	"smithing", "alchemy", "cooking", "tailoring",
	"speech", "barter",
}

func (e *Entity) AddSkillXP(skillName string, amount int) {
	if amount <= 0 {
		return
	}
	cur := e.Skills[skillName]
	totalXP := cur*SkillXPPerLevel + amount
	e.Skills[skillName] = totalXP / SkillXPPerLevel
}

func (e *Entity) SkillLevel(skillName string) int {
	if pts, ok := e.Skills[skillName]; ok {
		return pts
	}
	return 0
}

func (e *Entity) SkillXP(skillName string) int {
	return e.SkillLevel(skillName) * SkillXPPerLevel
}

func WeaponSkillName(weaponDefID string) string {
	switch weaponDefID {
	case "iron_sword", "short_sword", "kusanagi":
		return "swords"
	case "iron_axe", "iron_gauntlets", "orc_cleaver":
		return "axes"
	case "iron_spear", "gungnir", "bident", "trident":
		return "spears"
	case "dagger", "vampire_fang", "goblin_shiv", "cultist_dagger":
		return "daggers"
	case "claws", "fangs", "tusks":
		return "unarmed"
	default:
		return "unarmed"
	}
}
