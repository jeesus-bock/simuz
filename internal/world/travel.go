package world

import "math"

type TravelStatus int

const (
	TravelNotTraveling TravelStatus = iota
	TravelInProgress
	TravelArrived
	TravelInterrupted
)

type TravelState struct {
	EntityID     string       `json:"entity_id"`
	FromID       string       `json:"from_id"`
	ToID         string       `json:"to_id"`
	Mode         TravelMode   `json:"mode"`
	TotalTicks   int          `json:"total_ticks"`
	ElapsedTicks int          `json:"elapsed_ticks"`
	Status       TravelStatus `json:"status"`
}

func TravelTime(from, to *Location, mode TravelMode) int {
	dist := from.DistanceTo(to)
	speed := mode.Speed()
	if speed <= 0 {
		speed = 5
	}
	ticks := int(math.Ceil(dist / speed))
	if ticks < 1 {
		ticks = 1
	}
	return ticks
}

func TravelTimeWithWeather(from, to *Location, mode TravelMode, weather *Weather) int {
	ticks := TravelTime(from, to, mode)
	if weather != nil {
		mod := weather.TravelSpeedModifier()
		if mod > 0 && mod < 1 {
			ticks = int(math.Ceil(float64(ticks) / mod))
		}
	}
	if ticks < 1 {
		ticks = 1
	}
	if ticks > 500 {
		ticks = 500
	}
	return ticks
}

func (ts *TravelState) Progress() float64 {
	if ts.TotalTicks == 0 {
		return 1
	}
	return float64(ts.ElapsedTicks) / float64(ts.TotalTicks)
}

func (ts *TravelState) Tick() {
	if ts.Status != TravelInProgress {
		return
	}
	ts.ElapsedTicks++
	if ts.ElapsedTicks >= ts.TotalTicks {
		ts.Status = TravelArrived
	}
}

func (ts *TravelState) Interrupt() {
	ts.Status = TravelInterrupted
}

func NewTravel(entityID, fromID, toID string, totalTicks int, mode TravelMode) *TravelState {
	if totalTicks < 1 {
		totalTicks = 1
	}
	return &TravelState{
		EntityID:     entityID,
		FromID:       fromID,
		ToID:         toID,
		Mode:         mode,
		TotalTicks:   totalTicks,
		ElapsedTicks: 0,
		Status:       TravelInProgress,
	}
}
