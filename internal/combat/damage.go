// Package combat contains the core combat rules, relation handling, and attack resolution helpers.
package combat

import (
	"simuz/internal/entity"
	"simuz/internal/items"
)

type DamageApplication struct {
	TargetID        string            `json:"target_id"`
	RawDamage       int               `json:"raw_damage"`
	DR              float64           `json:"dr"`
	NetDamage       int               `json:"net_damage"`
	DamageType      items.DamageType  `json:"damage_type"`
	HitLocation     items.HitLocation `json:"hit_location"`
	WoundMultiplier float64           `json:"wound_multiplier"`
	WoundDamage     int               `json:"wound_damage"`
	AppliedHP       int               `json:"applied_hp"`
	ShockPenalty    int               `json:"shock_penalty"`
	Fatal           bool              `json:"fatal"`
}

func ApplyDamage(target *entity.Entity, rawDamage int, dmgType items.DamageType, hitLoc items.HitLocation, dr float64, isFlexible bool) DamageApplication {
	app := DamageApplication{
		TargetID:        target.ID,
		RawDamage:       rawDamage,
		DR:              dr,
		DamageType:      dmgType,
		HitLocation:     hitLoc,
		WoundMultiplier: hitLoc.WoundMultiplier(),
	}

	netDmg := float64(rawDamage) - dr
	if netDmg < 0 {
		netDmg = 0
	}
	app.NetDamage = int(netDmg)

	woundDmg := int(float64(app.NetDamage) * app.WoundMultiplier)
	app.WoundDamage = woundDmg

	if woundDmg < 1 && app.NetDamage > 0 {
		woundDmg = 1
	}

	app.AppliedHP = woundDmg

	target.TakeDamage(woundDmg)
	if !target.Alive {
		app.Fatal = true
	}

	if woundDmg >= 2 {
		app.ShockPenalty = woundDmg / 2
		if app.ShockPenalty > 4 {
			app.ShockPenalty = 4
		}
	}

	return app
}

func ShockPenaltyForDamage(dmg int) int {
	if dmg < 2 {
		return 0
	}
	p := dmg / 2
	if p > 4 {
		return 4
	}
	return p
}
