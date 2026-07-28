// Package combat contains the core combat rules, relation handling, and attack resolution helpers.
package combat

import (
	"fmt"
	"log"
	"math/rand"

	"simuz/internal/entity"
	"simuz/internal/items"
)

type Event struct {
	Tick         uint64 `json:"tick"`
	LocationID   string `json:"location_id"`
	AttackerID   string `json:"attacker_id"`
	AttackerName string `json:"attacker_name"`
	DefenderID   string `json:"defender_id"`
	DefenderName string `json:"defender_name"`
	Action       string `json:"action"` // "miss","crit_miss","evade","hit","death","loot"
	Damage       int    `json:"damage"`
	HPLeft       int    `json:"hp_left"`
	Weapon       string `json:"weapon"`
}

var (
	CombatLog      []string
	locationEvents = make(map[string][]Event)
	globalTick     uint64
	weatherVisMod  = 1.0
)

func SetTick(t uint64) { globalTick = t }

// SetWeatherVisibility sets outdoor visibility modifier for hit rolls (1.0 = clear).
func SetWeatherVisibility(mod float64) {
	if mod <= 0 {
		mod = 0.1
	}
	if mod > 1.5 {
		mod = 1.5
	}
	weatherVisMod = mod
}

func ResetWeatherVisibility() { weatherVisMod = 1.0 }

func ClearLog() {
	CombatLog = nil
	locationEvents = make(map[string][]Event)
}

func RecentLog(n int) []string {
	if len(CombatLog) <= n {
		out := make([]string, len(CombatLog))
		copy(out, CombatLog)
		return out
	}
	return CombatLog[len(CombatLog)-n:]
}

func LocationEvents(locID string, n int) []Event {
	events := locationEvents[locID]
	if len(events) <= n {
		out := make([]Event, len(events))
		copy(out, events)
		return out
	}
	out := make([]Event, n)
	copy(out, events[len(events)-n:])
	return out
}

func addLog(msg string) {
	CombatLog = append(CombatLog, msg)
	if len(CombatLog) > 500 {
		CombatLog = CombatLog[len(CombatLog)-250:]
	}
	log.Printf("[combat] %s", msg)
}

func recordEvent(locID string, e Event) {
	locationEvents[locID] = append(locationEvents[locID], e)
	if len(locationEvents[locID]) > 500 {
		locationEvents[locID] = locationEvents[locID][len(locationEvents[locID])-250:]
	}
}

func LootCorpse(winner, loser *entity.Entity) {
	for i := len(loser.Inventory) - 1; i >= 0; i-- {
		item := loser.Inventory[i]
		loser.Inventory = append(loser.Inventory[:i], loser.Inventory[i+1:]...)
		winner.Inventory = append(winner.Inventory, item)
	}
	slots := []struct {
		slot string
		item **items.ItemInstance
	}{
		{"head", &loser.Equipment.Head},
		{"body", &loser.Equipment.Body},
		{"weapon", &loser.Equipment.Weapon},
		{"offhand", &loser.Equipment.Offhand},
		{"feet", &loser.Equipment.Feet},
		{"hands", &loser.Equipment.Hands},
		{"neck", &loser.Equipment.Neck},
		{"finger1", &loser.Equipment.Finger1},
		{"finger2", &loser.Equipment.Finger2},
	}
	for _, s := range slots {
		item := *s.item
		if item == nil {
			continue
		}
		item.Equipped = false
		*s.item = nil
		winner.Inventory = append(winner.Inventory, *item)
	}

	msg := fmt.Sprintf("%s looted %s's corpse", winner.Name, loser.Name)
	addLog(msg)
	recordEvent(loser.LocationID, Event{
		Tick:         globalTick,
		LocationID:   loser.LocationID,
		AttackerID:   winner.ID,
		AttackerName: winner.Name,
		DefenderID:   loser.ID,
		DefenderName: loser.Name,
		Action:       "loot",
	})
}

// KillChance returns the percent chance an attacker finishes a downed foe
// instead of leaving them knocked out. Merciless factions almost always kill.
func KillChance(attacker *entity.Entity) int {
	if attacker == nil {
		return 50
	}
	switch attacker.Species {
	case "orc":
		return 95 // orcs almost never spare anyone
	case "wolf", "bear", "boar", "spider", "bat", "badger", "rat", "rat_king", "dog":
		return 90 // beasts rarely show mercy
	case "dragon":
		return 98
	case "vampire", "hag":
		return 88
	case "kobold", "goblin":
		return 85
	case "werewolf":
		return 92
	}

	switch attacker.Faction {
	case "orc":
		return 95
	case "beast", "vermin":
		return 90
	case "dragon":
		return 98
	case "undead", "hag", "kobold", "goblin", "cultist", "werewolf":
		return 85
	case "bandit", "thief":
		return 72
	case "guard":
		return 40 // often take prisoners
	case "civilian", "merchant", "elf", "fey", "ranger":
		return 28
	case "deity":
		return 55
	default:
		return 55
	}
}

// resolveLethality decides whether a downed defender dies based on the attacker's nature.
// alreadyDown is true when the defender was unconscious before this blow (finishing move).
func resolveLethality(attacker, defender *entity.Entity, alreadyDown bool, damage int, rng *rand.Rand) {
	if defender == nil || !defender.Alive || defender.Conscious || defender.Immortal {
		return
	}
	chance := KillChance(attacker)
	if alreadyDown {
		chance += 15 // finishing a helpless foe is easier / more deliberate
	}
	if defender.MaxHP > 0 && damage >= defender.MaxHP/2 {
		chance += 10 // brutal overkill
	}
	if chance > 100 {
		chance = 100
	}
	if chance < 5 {
		chance = 5
	}
	if rng.Intn(100) < chance {
		defender.Kill()
	}
}

func weaponDisplayName(attacker *entity.Entity) string {
	if attacker == nil || attacker.Equipment.Weapon == nil || attacker.Equipment.Weapon.Def == nil {
		return "fists"
	}
	return attacker.Equipment.Weapon.Def.Name
}

func combatPowerScore(e *entity.Entity) int {
	if e == nil {
		return 0
	}
	attrs := e.EffectiveAttrs()
	score := e.Level*4 + attrs.STR*2 + attrs.DEX + attrs.CON
	if e.MaxHP > 0 {
		score += e.HP * 10 / e.MaxHP
	}
	if e.AI.Brave {
		score += 6
	}
	return score
}

func applyBraveryBoost(ent, opponent *entity.Entity) {
	if ent == nil || opponent == nil || !ent.AI.Brave {
		return
	}
	if ent.HasMoodModifierSource("combat_bravery") {
		return
	}
	if combatPowerScore(opponent) >= combatPowerScore(ent)+8 {
		ent.AddMoodModifier("combat_bravery", "brave", 6)
	}
}

func SimpleAttack(attacker, defender *entity.Entity, rng *rand.Rand) bool {
	if attacker == nil || defender == nil || !attacker.Alive || !defender.Alive {
		return false
	}
	if !attacker.Conscious {
		return false
	}

	applyBraveryBoost(attacker, defender)
	applyBraveryBoost(defender, attacker)

	wpnName := weaponDisplayName(attacker)
	weapon := attacker.Equipment.Weapon
	alreadyDown := !defender.Conscious

	// Coup de grace: already unconscious targets cannot evade.
	if alreadyDown {
		finalDmg := 1 + attacker.EffectiveAttrs().STRMod()
		if finalDmg < 1 {
			finalDmg = 1
		}
		if weapon != nil && weapon.Def != nil {
			for _, mwd := range meleeWeapons {
				if mwd.DefID == weapon.DefID {
					finalDmg = mwd.Damage.Flat + mwd.Damage.Dice + attacker.EffectiveAttrs().STRMod()
					if finalDmg < 2 {
						finalDmg = 2
					}
					break
				}
			}
		}
		defender.TakeDamage(finalDmg)
		resolveLethality(attacker, defender, true, finalDmg, rng)

		action := "knockout"
		if !defender.Alive {
			action = "death"
		}
		if !defender.Alive {
			addLog(fmt.Sprintf("%s finishes %s with %s — %s is slain!", attacker.Name, defender.Name, wpnName, defender.Name))
		} else {
			addLog(fmt.Sprintf("%s stands over %s but spares their life", attacker.Name, defender.Name))
			action = "spare"
		}
		recordEvent(defender.LocationID, Event{
			Tick:         globalTick,
			LocationID:   defender.LocationID,
			AttackerID:   attacker.ID,
			AttackerName: attacker.Name,
			DefenderID:   defender.ID,
			DefenderName: defender.Name,
			Action:       action,
			Damage:       finalDmg,
			HPLeft:       defender.HP,
			Weapon:       wpnName,
		})
		wpnDefID := "fists"
		if weapon != nil && weapon.Def != nil {
			wpnDefID = weapon.DefID
		}
		attacker.AddSkillXP(entity.WeaponSkillName(wpnDefID), 5+defender.Level)
		return true
	}

	atkAttrs := attacker.EffectiveAttrs()
	defAttrs := defender.EffectiveAttrs()
	atkSkill := 10 + attacker.Level + (atkAttrs.DEX-10)/2 + (atkAttrs.STR-10)/4
	defSkill := 8 + defender.Level + (defAttrs.DEX-10)/2

	if weapon != nil && weapon.Def != nil {
		atkSkill += 2
	}

	if attacker.IsBlessed(globalTick) {
		atkSkill += 2
	}

	effectiveAtk := atkSkill
	if weatherVisMod < 1.0 {
		effectiveAtk = int(float64(atkSkill) * weatherVisMod)
		if effectiveAtk < 1 {
			effectiveAtk = 1
		}
	}

	roll := rng.Intn(20) + 1
	if roll == 1 || roll > effectiveAtk {
		action := "miss"
		if roll == 1 {
			action = "crit_miss"
		}
		msg := fmt.Sprintf("%s attacked %s with %s but missed!", attacker.Name, defender.Name, wpnName)
		if roll == 1 {
			msg = fmt.Sprintf("%s attacked %s with %s but critically missed!", attacker.Name, defender.Name, wpnName)
		}
		addLog(msg)
		recordEvent(defender.LocationID, Event{
			Tick:         globalTick,
			LocationID:   defender.LocationID,
			AttackerID:   attacker.ID,
			AttackerName: attacker.Name,
			DefenderID:   defender.ID,
			DefenderName: defender.Name,
			Action:       action,
			Weapon:       wpnName,
		})
		return false
	}

	defRoll := rng.Intn(20) + 1
	if defRoll <= defSkill {
		addLog(fmt.Sprintf("%s attacked %s with %s but was evaded!", attacker.Name, defender.Name, wpnName))
		recordEvent(defender.LocationID, Event{
			Tick:         globalTick,
			LocationID:   defender.LocationID,
			AttackerID:   attacker.ID,
			AttackerName: attacker.Name,
			DefenderID:   defender.ID,
			DefenderName: defender.Name,
			Action:       "evade",
			Weapon:       wpnName,
		})
		return false
	}

	baseDmg := (1 + atkAttrs.STRMod()) * 2
	if baseDmg < 1 {
		baseDmg = 1
	}

	if weapon != nil && weapon.Def != nil {
		for _, mwd := range meleeWeapons {
			if mwd.DefID == weapon.DefID {
				diceResult := 0
				for d := 0; d < mwd.Damage.Dice; d++ {
					diceResult += rng.Intn(mwd.Damage.Sides) + 1
				}
				baseDmg = diceResult + mwd.Damage.Flat + atkAttrs.STRMod()
				if baseDmg < 1 {
					baseDmg = 1
				}
				break
			}
		}
	}

	defenderDR := 0
	for _, piece := range []*items.ItemInstance{defender.Equipment.Body, defender.Equipment.Head, defender.Equipment.Offhand} {
		if piece != nil && piece.Def != nil {
			for _, ad := range armorDefs {
				if ad.DefID == piece.DefID {
					defenderDR += int(ad.DR)
					break
				}
			}
		}
	}

	if defender.IsBlessed(globalTick) {
		defenderDR++
	}

	finalDmg := baseDmg - defenderDR
	if finalDmg < 1 {
		finalDmg = 1
	}
	if roll == 20 {
		finalDmg *= 2
	}

	defender.TakeDamage(finalDmg)
	if defender.Alive && !defender.Conscious {
		resolveLethality(attacker, defender, false, finalDmg, rng)
	}

	action := "hit"
	hpLeft := defender.HP
	if !defender.Alive {
		action = "death"
	} else if !defender.Conscious {
		action = "knockout"
	}
	event := Event{
		Tick:         globalTick,
		LocationID:   defender.LocationID,
		AttackerID:   attacker.ID,
		AttackerName: attacker.Name,
		DefenderID:   defender.ID,
		DefenderName: defender.Name,
		Action:       action,
		Damage:       finalDmg,
		HPLeft:       hpLeft,
		Weapon:       wpnName,
	}

	if !defender.Alive {
		addLog(fmt.Sprintf("%s strikes %s with %s for %d damage — %s is slain!", attacker.Name, defender.Name, wpnName, finalDmg, defender.Name))
	} else if !defender.Conscious {
		addLog(fmt.Sprintf("%s hits %s with %s for %d damage — %s is knocked out!", attacker.Name, defender.Name, wpnName, finalDmg, defender.Name))
	} else {
		addLog(fmt.Sprintf("%s hits %s with %s for %d damage (%d HP left)", attacker.Name, defender.Name, wpnName, finalDmg, defender.HP))
	}
	recordEvent(defender.LocationID, event)

	wpnDefID := "fists"
	if weapon != nil && weapon.Def != nil {
		wpnDefID = weapon.DefID
	}
	skillName := entity.WeaponSkillName(wpnDefID)
	attacker.AddSkillXP(skillName, 5+defender.Level)
	return true
}

type meleeWeaponDef struct {
	DefID  string
	Damage items.DamageRoll
}

var meleeWeapons = []meleeWeaponDef{
	{"iron_sword", items.DamageRoll{Dice: 2, Sides: 6, Flat: 0}},
	{"short_sword", items.DamageRoll{Dice: 1, Sides: 6, Flat: 2}},
	{"dagger", items.DamageRoll{Dice: 1, Sides: 4, Flat: 1}},
	{"cudgel", items.DamageRoll{Dice: 1, Sides: 6, Flat: 0}},
	{"iron_axe", items.DamageRoll{Dice: 2, Sides: 6, Flat: 2}},
	{"iron_spear", items.DamageRoll{Dice: 2, Sides: 6, Flat: 1}},
	{"lightning_bolt", items.DamageRoll{Dice: 4, Sides: 6, Flat: 10}},
	{"bident", items.DamageRoll{Dice: 3, Sides: 6, Flat: 5}},
	{"trident", items.DamageRoll{Dice: 3, Sides: 6, Flat: 5}},
	{"gungnir", items.DamageRoll{Dice: 5, Sides: 6, Flat: 10}},
	{"mjolnir", items.DamageRoll{Dice: 6, Sides: 6, Flat: 15}},
	{"ruyi_bang", items.DamageRoll{Dice: 4, Sides: 6, Flat: 10}},
	{"kusanagi", items.DamageRoll{Dice: 4, Sides: 6, Flat: 8}},
	{"jade_scepter", items.DamageRoll{Dice: 2, Sides: 6, Flat: 5}},
	{"smith_hammer", items.DamageRoll{Dice: 1, Sides: 6, Flat: 1}},
	{"sickle", items.DamageRoll{Dice: 1, Sides: 6, Flat: 2}},
	{"cultist_dagger", items.DamageRoll{Dice: 1, Sides: 4, Flat: 2}},
	{"necromancer_staff", items.DamageRoll{Dice: 1, Sides: 8, Flat: 3}},
	{"vampire_fang", items.DamageRoll{Dice: 2, Sides: 4, Flat: 2}},
	{"claws", items.DamageRoll{Dice: 1, Sides: 6, Flat: 1}},
	{"fangs", items.DamageRoll{Dice: 1, Sides: 4, Flat: 2}},
	{"tusks", items.DamageRoll{Dice: 1, Sides: 8, Flat: 1}},
	{"goblin_shiv", items.DamageRoll{Dice: 1, Sides: 4, Flat: 1}},
	{"orc_cleaver", items.DamageRoll{Dice: 2, Sides: 6, Flat: 3}},
}

type armorDef struct {
	DefID string
	DR    float64
}

var armorDefs = []armorDef{
	{"chainmail", 4},
	{"scale_armor", 3},
	{"leather_armor", 2},
	{"iron_helmet", 2},
	{"leather_helmet", 1},
	{"iron_boots", 1},
	{"leather_boots", 0},
	{"iron_shield", 3},
	{"wooden_shield", 1},
	{"leather_gloves", 0},
	{"fine_clothes", 0},
	{"common_clothes", 0},
	{"work_tunic", 0},
	{"simple_robe", 0},
	{"priest_robe", 0},
	{"toga", 0},
	{"dark_robe", 0},
	{"imperial_robe", 1},
	{"feather_cloak", 0},
	{"megingjord", 5},
	{"aegis_shield", 6},
	{"helm_of_darkness", 3},
	{"dragon_crown", 4},
	{"golden_circlet", 1},
	{"iron_gauntlets", 1},
}
