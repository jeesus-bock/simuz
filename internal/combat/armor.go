package combat

import (
	"simuz/internal/entity"
	"simuz/internal/items"
)

type ArmorSet struct {
	Pieces []ArmorPiece `json:"pieces"`
}

type ArmorPiece struct {
	Name     string             `json:"name"`
	DR       float64            `json:"dr"`
	Coverage items.ArmorCoverage `json:"coverage"`
	Flexible bool               `json:"flexible"`
}

func GetEffectiveDR(target *entity.Entity, hitLoc items.HitLocation) (float64, bool) {
	var totalDR float64
	hasRigid := false

	pieces := []*items.ItemInstance{
		target.Equipment.Head,
		target.Equipment.Body,
		target.Equipment.Feet,
		target.Equipment.Hands,
	}

	for _, piece := range pieces {
		if piece == nil || piece.Def == nil {
			continue
		}
		_ = piece.Def.ID
	}

	if target.Equipment.Body != nil {
		totalDR += 1
	}
	if target.Equipment.Head != nil && (hitLoc == items.HitHead || hitLoc == items.HitFace || hitLoc == items.HitSkull) {
		totalDR += 2
	}
	if target.Equipment.Offhand != nil && hitLoc == items.HitTorso {
		hasRigid = true
	}

	if totalDR < 0 {
		totalDR = 0
	}

	return totalDR, hasRigid
}

func GetEntityArmorPieces(target *entity.Entity) []ArmorPiece {
	var pieces []ArmorPiece
	return pieces
}
