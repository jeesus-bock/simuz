// Package entity defines the simulation entities, their attributes, and related behaviors.
package entity

import "strings"

const (
	SkillXPPerLevel = 100
)

// SkillNames lists all assorted fantasy, utility, and survival skills in the simulation.
var SkillNames = []string{
	// Combat (Martial)
	"swords", "axes", "spears", "daggers", "unarmed", "archery",

	// Magic & Astrological Academia
	"astrology", "necromancy", "destruction_magic", "illusion_magic",

	// Crafting & Production (The Industrialists & Peasants)
	"smithing", "alchemy", "cooking", "tailoring", "brewing",

	// Social, Criminal & Underworld (The Gangs & Cults)
	"speech", "barter", "thievery", "stealth", "extortion",

	// Survival & Exploration
	"scavenging", "herbalism", "navigation",
}

// AddSkillXP safely accumulates experience and cleanly processes level promotion.
func (e *Entity) AddSkillXP(skillName string, amount int) {
	if amount <= 0 {
		return
	}

	// 1. Check if the map is initialized to avoid dangerous nil map write panics
	if e.Skills == nil {
		e.Skills = make(map[string]int)
	}
	if e.SkillProgressXP == nil {
		e.SkillProgressXP = make(map[string]int)
	}

	// 2. Accumulate tracking progress inside a separate leftover XP tracking cache map
	currentProgress := e.SkillProgressXP[skillName] + amount
	levelsGained := currentProgress / SkillXPPerLevel
	leftoverXP := currentProgress % SkillXPPerLevel

	// 3. Promote the core level and update leftover remainder states safely
	e.Skills[skillName] = e.SkillLevel(skillName) + levelsGained
	e.SkillProgressXP[skillName] = leftoverXP
}

// SkillLevel safely queries the map layer and provides a fallback baseline level.
func (e *Entity) SkillLevel(skillName string) int {
	if e.Skills == nil {
		return 0
	}
	if pts, ok := e.Skills[skillName]; ok {
		return pts
	}
	return 0
}

// SkillXP calculates the explicit floor milestone point of the current level block.
func (e *Entity) SkillXP(skillName string) int {
	return e.SkillLevel(skillName) * SkillXPPerLevel
}

// WeaponSkillName maps weapon definition items cleanly to combat or magical skill brackets.
func WeaponSkillName(weaponDefID string) string {
	// Standardize to prevent casing validation issues
	id := strings.ToLower(weaponDefID)

	switch id {
	// Swords
	case "iron_sword", "short_sword", "kusanagi":
		return "swords"

	// Axes & Cleavers (Moved gauntlets to unarmed where they dynamically belong)
	case "iron_axe", "orc_cleaver", "battle_axe":
		return "axes"

	// Spears & Polearms
	case "iron_spear", "gungnir", "bident", "trident":
		return "spears"

	// Daggers & Shivs
	case "dagger", "vampire_fang", "goblin_shiv", "cultist_dagger":
		return "daggers"

	// Ranged Missile Options
	case "hunting_bow", "longbow", "cross_bow":
		return "archery"

	// Magic Elements (Astrological Assembly & Covens)
	case "copper_staff", "star_chart_wand", "fire_scroll":
		return "destruction_magic"
	case "scrying_mirror", "illusion_orb":
		return "illusion_magic"
	case "ritual_skull", "bone_staff":
		return "necromancy"

	// Claws, unarmed tools, and heavy punching gear
	case "claws", "fangs", "tusks", "iron_gauntlets", "bare_hands":
		return "unarmed"

	default:
		return "unarmed"
	}
}
