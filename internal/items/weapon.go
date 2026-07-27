package items

type DamageType int

const (
	DamageCrush DamageType = iota
	DamageCut
	DamageImpale
	DamagePiMinus
	DamagePi
	DamagePiPlus
	DamageBurn
)

func (dt DamageType) String() string {
	switch dt {
	case DamageCrush:
		return "cr"
	case DamageCut:
		return "cut"
	case DamageImpale:
		return "imp"
	case DamagePiMinus:
		return "pi-"
	case DamagePi:
		return "pi"
	case DamagePiPlus:
		return "pi+"
	case DamageBurn:
		return "burn"
	default:
		return "cr"
	}
}

func (dt DamageType) WoundMultiplier() float64 {
	switch dt {
	case DamageCrush:
		return 1.0
	case DamageCut:
		return 1.5
	case DamageImpale:
		return 2.0
	case DamagePiMinus:
		return 0.5
	case DamagePi:
		return 1.0
	case DamagePiPlus:
		return 1.5
	case DamageBurn:
		return 1.0
	default:
		return 1.0
	}
}

type DamageRoll struct {
	Dice  int `json:"dice"`
	Sides int `json:"sides"`
	Flat  int `json:"flat"`
}

func (d DamageRoll) Max() int {
	return d.Dice*d.Sides + d.Flat
}

func (d DamageRoll) Average() int {
	return (d.Dice*(d.Sides+1))/2 + d.Flat
}

type MeleeWeaponDef struct {
	Skill         string     `json:"skill"`
	Damage        DamageRoll `json:"damage"`
	DamageType    DamageType `json:"damage_type"`
	Reach         int        `json:"reach"`
	Parry         int        `json:"parry"`
	MinST         int        `json:"min_st"`
	Weight        float64    `json:"weight"`
}

type RangedWeaponDef struct {
	Skill      string     `json:"skill"`
	Damage     DamageRoll `json:"damage"`
	DamageType DamageType `json:"damage_type"`
	Acc        int        `json:"acc"`
	RangeMax   int        `json:"range_max"`
	ROF        int        `json:"rof"`
	Shots      int        `json:"shots"`
	MinST      int        `json:"min_st"`
}
