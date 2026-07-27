package items

type HitLocation int

const (
	HitTorso HitLocation = iota
	HitHead
	HitLeftArm
	HitRightArm
	HitLeftLeg
	HitRightLeg
	HitHand
	HitFoot
	HitNeck
	HitVitals
	HitGroin
	HitFace
	HitSkull
)

func (hl HitLocation) String() string {
	switch hl {
	case HitTorso:
		return "torso"
	case HitHead:
		return "head"
	case HitLeftArm:
		return "left_arm"
	case HitRightArm:
		return "right_arm"
	case HitLeftLeg:
		return "left_leg"
	case HitRightLeg:
		return "right_leg"
	case HitHand:
		return "hand"
	case HitFoot:
		return "foot"
	case HitNeck:
		return "neck"
	case HitVitals:
		return "vitals"
	case HitGroin:
		return "groin"
	case HitFace:
		return "face"
	case HitSkull:
		return "skull"
	default:
		return "torso"
	}
}

func (hl HitLocation) WoundMultiplier() float64 {
	switch hl {
	case HitTorso:
		return 1.0
	case HitHead, HitNeck, HitVitals:
		return 2.0
	case HitFace, HitGroin:
		return 1.5
	case HitSkull:
		return 4.0
	case HitLeftArm, HitRightArm, HitLeftLeg, HitRightLeg:
		return 1.0
	case HitHand, HitFoot:
		return 0.5
	default:
		return 1.0
	}
}

type ArmorCoverage int

const (
	CoversNone ArmorCoverage = 0
	CoversTorso ArmorCoverage = 1 << iota
	CoversHead
	CoversArms
	CoversLegs
	CoversHands
	CoversFeet
	CoversNeck
	CoversGroin
	CoversAll = CoversTorso | CoversHead | CoversArms | CoversLegs | CoversHands | CoversFeet | CoversNeck | CoversGroin
)

type ArmorDef struct {
	DR       float64        `json:"dr"`
	Coverage ArmorCoverage  `json:"coverage"`
	Flexible bool           `json:"flexible"`
	Weight   float64        `json:"weight"`
	Penalty  int            `json:"penalty"`
}

func (ac ArmorCoverage) Covers(loc HitLocation) bool {
	switch loc {
	case HitTorso:
		return ac&CoversTorso != 0
	case HitHead, HitFace, HitSkull:
		return ac&CoversHead != 0
	case HitLeftArm, HitRightArm:
		return ac&CoversArms != 0
	case HitLeftLeg, HitRightLeg:
		return ac&CoversLegs != 0
	case HitHand:
		return ac&CoversHands != 0
	case HitFoot:
		return ac&CoversFeet != 0
	case HitNeck:
		return ac&CoversNeck != 0
	case HitGroin:
		return ac&CoversGroin != 0
	default:
		return false
	}
}
