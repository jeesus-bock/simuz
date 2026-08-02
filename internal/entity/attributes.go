// Package entity defines the simulation entities, their attributes, and related behaviors.
package entity

type Attributes struct {
	STR int `json:"str"`
	DEX int `json:"dex"`
	CON int `json:"con"`
	INT int `json:"int"`
	WIS int `json:"wis"`
	CHA int `json:"cha"`
}

func NewAttributes(str, dex, con, intel, wis, cha int) Attributes {
	return Attributes{
		STR: clampAttr(str),
		DEX: clampAttr(dex),
		CON: clampAttr(con),
		INT: clampAttr(intel),
		WIS: clampAttr(wis),
		CHA: clampAttr(cha),
	}
}

func RandomAttributes(rng func(int) int) Attributes {
	return Attributes{
		STR: rng(18) + 3,
		DEX: rng(18) + 3,
		CON: rng(18) + 3,
		INT: rng(18) + 3,
		WIS: rng(18) + 3,
		CHA: rng(18) + 3,
	}
}

func (a Attributes) Modifier(score int) int {
	return (score - 10) / 2
}

func (a Attributes) STRMod() int { return a.Modifier(a.STR) }
func (a Attributes) DEXMod() int { return a.Modifier(a.DEX) }
func (a Attributes) CONMod() int { return a.Modifier(a.CON) }
func (a Attributes) INTMod() int { return a.Modifier(a.INT) }
func (a Attributes) WISMod() int { return a.Modifier(a.WIS) }
func (a Attributes) CHAMod() int { return a.Modifier(a.CHA) }

func clampAttr(v int) int {
	if v < 3 {
		return 3
	}
	if v > 20 {
		return 20
	}
	return v
}

// AverageAttrs returns the element-wise average of two attribute sets,
// used when computing child attributes from parents.
func AverageAttrs(a, b Attributes) Attributes {
	avg := func(x, y int) int {
		return (x + y) / 2
	}
	return Attributes{
		STR: avg(a.STR, b.STR),
		DEX: avg(a.DEX, b.DEX),
		CON: avg(a.CON, b.CON),
		INT: avg(a.INT, b.INT),
		WIS: avg(a.WIS, b.WIS),
		CHA: avg(a.CHA, b.CHA),
	}
}
