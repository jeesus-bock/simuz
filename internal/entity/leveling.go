package entity

import "log"

func (e *Entity) LevelUpThreshold() int {
	return e.Level * 100
}

func (e *Entity) CheckLevelUp() bool {
	if !CanLevelUp(e.Species) || e.XP < e.LevelUpThreshold() {
		return false
	}

	e.XP -= e.LevelUpThreshold()
	e.Level++

	e.Attributes.STR++
	e.Attributes.DEX++
	e.Attributes.CON++

	e.MaxHP = e.Attributes.CON*2 + e.Level*2
	e.MaxFP = e.Attributes.CON + e.Attributes.STR/2
	if e.MaxHP < 1 {
		e.MaxHP = 1
	}
	if e.MaxFP < 1 {
		e.MaxFP = 1
	}
	e.HP = e.MaxHP
	e.FP = e.MaxFP

	log.Printf("[levelup] %s reached level %d!", e.Name, e.Level)

	if e.XP >= e.LevelUpThreshold() {
		e.CheckLevelUp()
	}

	return true
}
