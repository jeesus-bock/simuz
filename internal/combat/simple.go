package combat

import (
	"encoding/json"
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

type FactionRelation int

const (
	Neutral  FactionRelation = 0
	Hostile  FactionRelation = iota
	Friendly
)

var factionRelations map[string]map[string]FactionRelation

func initFactionRelations() {
	if factionRelations != nil {
		return
	}
	factionRelations = make(map[string]map[string]FactionRelation)
	pairs := [][2]string{
		{"orc", "elf"},
		{"civilian", "thief"}, {"civilian", "bandit"}, {"civilian", "vermin"},
		{"civilian", "werewolf"}, {"civilian", "cultist"},
		{"guard", "thief"}, {"guard", "bandit"}, {"guard", "vermin"},
		{"guard", "werewolf"}, {"guard", "cultist"},
		{"merchant", "thief"}, {"merchant", "bandit"}, {"merchant", "werewolf"}, {"merchant", "cultist"},
		{"vermin", "civilian"}, {"vermin", "guard"}, {"vermin", "ranger"},
		{"beast", "ranger"},
		{"kobold", "civilian"}, {"kobold", "guard"}, {"kobold", "beast"}, {"kobold", "fey"},
		{"undead", "civilian"}, {"undead", "guard"}, {"undead", "fey"},
		{"hag", "civilian"}, {"hag", "guard"}, {"hag", "fey"},
		{"werewolf", "ranger"}, {"cultist", "ranger"},
	}
	for _, p := range pairs {
		if _, ok := factionRelations[p[0]]; !ok {
			factionRelations[p[0]] = make(map[string]FactionRelation)
		}
		factionRelations[p[0]][p[1]] = Hostile
	}
	deityHostiles := []string{"orc", "elf", "beast", "thief", "bandit", "vermin", "goblin", "kobold", "undead", "hag", "werewolf", "cultist", "dragon"}
	factionRelations["deity"] = make(map[string]FactionRelation)
	for _, h := range deityHostiles {
		factionRelations["deity"][h] = Hostile
	}
}

func Relation(a, b string) FactionRelation {
	initFactionRelations()
	if a == b {
		return Friendly
	}
	if m, ok := factionRelations[a]; ok {
		if rel, ok := m[b]; ok {
			return rel
		}
	}
	if m, ok := factionRelations[b]; ok {
		if rel, ok := m[a]; ok {
			return rel
		}
	}
	return Neutral
}

func SetRelation(a, b string, rel FactionRelation) {
	initFactionRelations()
	if a == b {
		return
	}
	if _, ok := factionRelations[a]; !ok {
		factionRelations[a] = make(map[string]FactionRelation)
	}
	factionRelations[a][b] = rel
	if _, ok := factionRelations[b]; !ok {
		factionRelations[b] = make(map[string]FactionRelation)
	}
	factionRelations[b][a] = rel
}

func RelationsJSON() string {
	initFactionRelations()
	b, err := json.Marshal(factionRelations)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func LoadRelationsJSON(jsonStr string) {
	factionRelations = make(map[string]map[string]FactionRelation)
	if jsonStr == "" || jsonStr == "{}" {
		initFactionRelations()
		return
	}
	var raw map[string]map[string]float64
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		initFactionRelations()
		return
	}
	factionRelations = make(map[string]map[string]FactionRelation)
	for a, m := range raw {
		factionRelations[a] = make(map[string]FactionRelation)
		for b, v := range m {
			factionRelations[a][b] = FactionRelation(v)
		}
	}
}

func ShiftRelation(a, b string, delta int) {
	initFactionRelations()
	cur := Relation(a, b)
	newRel := Neutral
	switch cur {
	case Friendly:
		if delta < 0 {
			newRel = Hostile
		}
	case Hostile:
		if delta > 0 {
			newRel = Neutral
		}
	case Neutral:
		if delta > 0 {
			newRel = Friendly
		} else if delta < 0 {
			newRel = Hostile
		}
	}
	SetRelation(a, b, newRel)
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

func SimpleAttack(attacker, defender *entity.Entity, rng *rand.Rand) bool {
	atkAttrs := attacker.EffectiveAttrs()
	defAttrs := defender.EffectiveAttrs()
	atkSkill := 10 + attacker.Level + (atkAttrs.DEX-10)/2 + (atkAttrs.STR-10)/4
	defSkill := 8 + defender.Level + (defAttrs.DEX-10)/2

	weapon := attacker.Equipment.Weapon
	wpnName := "fists"
	if weapon != nil && weapon.Def != nil {
		wpnName = weapon.Def.Name
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
