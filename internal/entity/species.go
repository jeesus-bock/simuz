package entity

func SpeciesMaxAge(species string) int {
	switch species {
	case "deity":
		return 0
	case "human":
		return 30000
	case "elf":
		return 60000
	case "orc":
		return 22000
	case "goblin":
		return 12000
	case "fey":
		return 120000
	case "rat":
		return 3600
	case "rat_king":
		return 7200
	case "wolf":
		return 6000
	case "bear":
		return 12000
	case "boar":
		return 6000
	case "bat":
		return 2400
	case "spider":
		return 1800
	case "badger":
		return 7200
	case "chicken":
		return 2400
	case "pig":
		return 7200
	case "cow":
		return 14400
	case "sheep":
		return 10800
	case "goat":
		return 10800
	case "dog":
		return 12000
	case "kobold":
		return 6000
	case "vampire":
		return 0
	case "hag":
		return 90000
	default:
		return 30000
	}
}

func SpeciesStarvationThreshold(species string) int {
	maxAge := SpeciesMaxAge(species)
	if maxAge <= 0 {
		return 0
	}
	return maxAge / 3
}

func ShouldAutoFeed(species string) bool {
	switch species {
	case "chicken", "pig", "cow", "sheep", "goat":
		return false
	case "kobold", "vampire", "hag":
		return true
	default:
		return true
	}
}

func StarvationDamageInterval() int {
	return 10
}

func CanLevelUp(species string) bool {
	switch species {
	case "human", "orc", "elf", "goblin", "fey", "rat_king",
		"kobold", "vampire", "hag", "deity":
		return true
	default:
		return false
	}
}
