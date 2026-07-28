// Package combat contains the core combat rules, relation handling, and attack resolution helpers.
package combat

import (
	"math"
	"math/rand"

	"simuz/internal/entity"
	"simuz/internal/items"
)

type AttackType int

const (
	AttackMelee AttackType = iota
	AttackRanged
	AttackNatural
)

type CombatAction int

const (
	ActionAttack CombatAction = iota
	ActionAllOutAttack
	ActionFeint
	ActionMoveAndAttack
	ActionReady
	ActionWait
)

type Attack struct {
	ID          string              `json:"id"`
	AttackerID  string              `json:"attacker_id"`
	DefenderID  string              `json:"defender_id"`
	Action      CombatAction        `json:"action"`
	Type        AttackType          `json:"type"`
	Weapon      *items.ItemInstance `json:"weapon,omitempty"`
	HitLocation items.HitLocation   `json:"hit_location"`
	Roll        int                 `json:"roll"`
	SkillLevel  int                 `json:"skill_level"`
	Margin      int                 `json:"margin"`
	Hit         bool                `json:"hit"`
}

type AttackResult struct {
	Attack          Attack            `json:"attack"`
	Damage          int               `json:"damage"`
	DamageType      items.DamageType  `json:"damage_type"`
	RawDamage       int               `json:"raw_damage"`
	DR              float64           `json:"dr"`
	NetDamage       int               `json:"net_damage"`
	WoundMultiplier float64           `json:"wound_multiplier"`
	WoundDamage     int               `json:"wound_damage"`
	Location        items.HitLocation `json:"location"`
	Defended        bool              `json:"defended"`
	DefenseRoll     int               `json:"defense_roll,omitempty"`
	DefenseType     string            `json:"defense_type,omitempty"`
	Critical        bool              `json:"critical"`
}

type DefenseType int

const (
	DefenseDodge DefenseType = iota
	DefenseParry
	DefenseBlock
)

func ResolveMeleeAttack(attacker, defender *entity.Entity, action CombatAction, hitLoc items.HitLocation, weapon *items.ItemInstance, rng *rand.Rand) AttackResult {
	result := AttackResult{
		Location: hitLoc,
	}

	skillName := "brawling"
	damage := items.DamageRoll{Dice: 1, Sides: 2, Flat: 0}
	dmgType := items.DamageCrush

	if weapon != nil && weapon.Def != nil {
		skillName = "sword"
	}

	skill := attacker.SkillLevel(skillName)
	roll := rng.Intn(16) + 3

	atk := Attack{
		AttackerID:  attacker.ID,
		DefenderID:  defender.ID,
		Action:      action,
		Type:        AttackMelee,
		Weapon:      weapon,
		HitLocation: hitLoc,
		Roll:        roll,
		SkillLevel:  skill,
		Margin:      skill - roll,
		Hit:         roll <= skill,
	}
	result.Attack = atk

	if !atk.Hit {
		return result
	}

	result.Critical = roll <= 4

	defSkill := defender.SkillLevel("dodge")
	defRoll := rng.Intn(16) + 3
	result.DefenseRoll = defRoll
	result.DefenseType = "dodge"

	if defRoll <= defSkill {
		result.Damage = 0
		result.Defended = true
		return result
	}

	diceResult := 0
	for i := 0; i < damage.Dice; i++ {
		diceResult += rng.Intn(damage.Sides) + 1
	}
	result.RawDamage = diceResult + damage.Flat + attacker.EffectiveAttrs().STRMod()
	result.DamageType = dmgType
	if result.RawDamage < 0 {
		result.RawDamage = 0
	}

	result.NetDamage = result.RawDamage
	result.WoundMultiplier = hitLoc.WoundMultiplier()
	woundDmg := int(math.Round(float64(result.NetDamage) * result.WoundMultiplier))
	result.WoundDamage = woundDmg
	result.Damage = woundDmg

	return result
}

func RollDamage(damage items.DamageRoll, rng *rand.Rand) int {
	result := 0
	for i := 0; i < damage.Dice; i++ {
		result += rng.Intn(damage.Sides) + 1
	}
	return result + damage.Flat
}
