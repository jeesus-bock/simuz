package world

import "fmt"

type Season int

const (
	Spring Season = iota
	Summer
	Autumn
	Winter
)

func (s Season) String() string {
	switch s {
	case Spring:
		return "spring"
	case Summer:
		return "summer"
	case Autumn:
		return "autumn"
	case Winter:
		return "winter"
	default:
		return "unknown"
	}
}

type DayPhase int

const (
	Dawn DayPhase = iota
	Day
	Dusk
	Night
)

func (dp DayPhase) String() string {
	switch dp {
	case Dawn:
		return "dawn"
	case Day:
		return "day"
	case Dusk:
		return "dusk"
	case Night:
		return "night"
	default:
		return "unknown"
	}
}

type GameTime struct {
	Tick  uint64 `json:"tick"`
	Day   int    `json:"day"`
	Hour  int    `json:"hour"`
	Minute int   `json:"minute"`
	Speed int    `json:"speed"`
}

func NewGameTime(speed int) GameTime {
	return GameTime{
		Tick:   0,
		Day:    1,
		Hour:   6,
		Minute: 0,
		Speed:  speed,
	}
}

func (gt *GameTime) Advance() {
	gt.Tick++
	gameMinutes := gt.Speed
	gt.Minute += gameMinutes
	for gt.Minute >= 60 {
		gt.Minute -= 60
		gt.Hour++
	}
	for gt.Hour >= 24 {
		gt.Hour -= 24
		gt.Day++
	}
}

func (gt *GameTime) AdvanceN(ticks int) {
	for i := 0; i < ticks; i++ {
		gt.Advance()
	}
}

func (gt *GameTime) Phase() DayPhase {
	switch {
	case gt.Hour >= 5 && gt.Hour < 7:
		return Dawn
	case gt.Hour >= 7 && gt.Hour < 19:
		return Day
	case gt.Hour >= 19 && gt.Hour < 21:
		return Dusk
	default:
		return Night
	}
}

func (gt *GameTime) Season() Season {
	dayInYear := gt.Day % 360
	switch {
	case dayInYear < 90:
		return Spring
	case dayInYear < 180:
		return Summer
	case dayInYear < 270:
		return Autumn
	default:
		return Winter
	}
}

func (gt *GameTime) IsDaytime() bool {
	return gt.Hour >= 6 && gt.Hour < 20
}

// DayOfWeek returns 1-7 (Monday-Sunday)
func (gt *GameTime) DayOfWeek() int {
	return (gt.Day-1)%7 + 1
}

// DayOfWeekName returns the name of the day
func (gt *GameTime) DayOfWeekName() string {
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	return days[gt.DayOfWeek()-1]
}

func (gt *GameTime) String() string {
	return fmt.Sprintf("Day %d, %02d:%02d", gt.Day, gt.Hour, gt.Minute)
}
