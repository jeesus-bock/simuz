package entity

import "math/rand"

// ProfessionBonus returns the static attribute bonus for a given profession.
// The berzerker gets +5-10 to STR and DEX.
func ProfessionBonus(profession string) (strBonus, dexBonus int) {
	switch profession {
	case "berzerker":
		b := rand.Intn(6) + 5 // 5-10
		return b, b
	default:
		return 0, 0
	}
}
