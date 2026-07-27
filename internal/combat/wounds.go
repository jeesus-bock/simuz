package combat

import (
	"simuz/internal/entity"
	"simuz/internal/items"
)

type Wound struct {
	Location    items.HitLocation `json:"location"`
	Damage      int               `json:"damage"`
	DamageType  items.DamageType  `json:"damage_type"`
	Crippled    bool              `json:"crippled"`
	BleedingRate int              `json:"bleeding_rate"`
	TickApplied uint64            `json:"tick_applied"`
}

type CrippleThreshold int

const (
	CrippleHand  CrippleThreshold = 3
	CrippleFoot  CrippleThreshold = 3
	CrippleArm   CrippleThreshold = 5
	CrippleLeg   CrippleThreshold = 5
	CrippleNeck  CrippleThreshold = 7
	CrippleTorso CrippleThreshold = 10
)

func CrippleThresholdForLocation(loc items.HitLocation) int {
	switch loc {
	case items.HitHand:
		return int(CrippleHand)
	case items.HitFoot:
		return int(CrippleFoot)
	case items.HitLeftArm, items.HitRightArm:
		return int(CrippleArm)
	case items.HitLeftLeg, items.HitRightLeg:
		return int(CrippleLeg)
	case items.HitNeck:
		return int(CrippleNeck)
	default:
		return int(CrippleTorso)
	}
}

func IsCrippling(loc items.HitLocation, damage int) bool {
	return damage >= CrippleThresholdForLocation(loc)
}

func BleedingRateForDamage(damage int, dmgType items.DamageType) int {
	if dmgType == items.DamageCrush || dmgType == items.DamageBurn {
		return 0
	}

	switch {
	case damage >= 10:
		return 3
	case damage >= 5:
		return 2
	case damage >= 1:
		return 1
	default:
		return 0
	}
}

type WoundManager struct {
	wounds map[string][]Wound
}

func NewWoundManager() *WoundManager {
	return &WoundManager{
		wounds: make(map[string][]Wound),
	}
}

func (wm *WoundManager) AddWound(entityID string, wound Wound) {
	wm.wounds[entityID] = append(wm.wounds[entityID], wound)
}

func (wm *WoundManager) WoundsFor(id string) []Wound {
	return wm.wounds[id]
}

func (wm *WoundManager) ProcessBleeding(entity *entity.Entity, tick uint64) int {
	totalBleed := 0
	wounds := wm.wounds[entity.ID]
	for i, w := range wounds {
		if w.BleedingRate > 0 {
			totalBleed += w.BleedingRate
			wounds[i].Damage += w.BleedingRate
		}
	}
	if totalBleed > 0 {
		entity.TakeDamage(totalBleed)
	}
	return totalBleed
}
