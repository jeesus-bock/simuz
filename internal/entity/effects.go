package entity

type ActiveEffect struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ItemDefID      string `json:"item_def_id"`
	StartTick      uint64 `json:"start_tick"`
	BoostRemaining int    `json:"boost_remaining"`
	CrashRemaining int    `json:"crash_remaining"`
	BoostMod       Attributes `json:"boost_mod"`
	CrashMod       Attributes `json:"crash_mod"`
	HealPerTick    int    `json:"heal_per_tick"`
	FPPerTick      int    `json:"fp_per_tick"`
}

func (e *Entity) ApplySubstance(name, itemDefID string, duration, crashDuration int, boostMod, crashMod Attributes, healHP, healFP, healPerTick, fpPerTick int, tick uint64) {
	if healHP > 0 {
		e.Heal(healHP)
	}
	if healFP > 0 {
		e.RestFP(healFP)
	}

	if duration <= 0 && crashDuration <= 0 && healPerTick <= 0 && fpPerTick <= 0 {
		return
	}

	effectID := itemDefID + "_" + e.ID + "_" + string(rune(tick))
	eff := ActiveEffect{
		ID:             effectID,
		Name:           name,
		ItemDefID:      itemDefID,
		StartTick:      tick,
		BoostRemaining: duration,
		CrashRemaining: 0,
		BoostMod:       boostMod,
		CrashMod:       crashMod,
		HealPerTick:    healPerTick,
		FPPerTick:      fpPerTick,
	}
	e.Effects = append(e.Effects, eff)
	if e.Effects == nil {
		e.Effects = []ActiveEffect{eff}
	}
}

func (e *Entity) HasEffect(name string) bool {
	for _, eff := range e.Effects {
		if eff.Name == name && (eff.BoostRemaining > 0 || eff.CrashRemaining > 0) {
			return true
		}
	}
	return false
}

func (e *Entity) TickEffects() {
	var active []ActiveEffect
	for _, eff := range e.Effects {
		if eff.HealPerTick > 0 {
			e.Heal(eff.HealPerTick)
		}
		if eff.FPPerTick > 0 {
			e.RestFP(eff.FPPerTick)
		}
		if eff.CrashRemaining > 0 {
			eff.CrashRemaining--
			if eff.CrashRemaining > 0 {
				active = append(active, eff)
			}
		} else if eff.BoostRemaining > 0 {
			eff.BoostRemaining--
			if eff.BoostRemaining > 0 {
				active = append(active, eff)
			} else if eff.CrashMod != (Attributes{}) {
				eff.CrashRemaining = 10
				active = append(active, eff)
			}
		}
	}
	e.Effects = active
}

func (e *Entity) EffectiveAttrs() Attributes {
	base := e.Attributes
	moodMods := e.MoodStatMods()
	base.STR += moodMods.STR
	base.DEX += moodMods.DEX
	base.CON += moodMods.CON
	base.INT += moodMods.INT
	base.WIS += moodMods.WIS
	base.CHA += moodMods.CHA
	for _, eff := range e.Effects {
		if eff.CrashRemaining > 0 {
			base.STR += eff.CrashMod.STR
			base.DEX += eff.CrashMod.DEX
			base.CON += eff.CrashMod.CON
			base.INT += eff.CrashMod.INT
			base.WIS += eff.CrashMod.WIS
			base.CHA += eff.CrashMod.CHA
		} else if eff.BoostRemaining > 0 {
			base.STR += eff.BoostMod.STR
			base.DEX += eff.BoostMod.DEX
			base.CON += eff.BoostMod.CON
			base.INT += eff.BoostMod.INT
			base.WIS += eff.BoostMod.WIS
			base.CHA += eff.BoostMod.CHA
		}
	}
	return base
}
