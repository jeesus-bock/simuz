package world

import (
	"math/rand"
	"testing"
)

func TestCanInstantMove(t *testing.T) {
	w := NewWorld()
	root := NewLocation("aetheria", "Aetheria", LocWorld, "", Position{})
	w.AddLocation(root)

	nh := NewLocation("northern_highlands", "NH", LocRegion, "aetheria", Position{X: 100, Y: 200})
	w.AddLocation(nh)

	fh := NewLocation("frosthold", "Frosthold", LocCity, "northern_highlands", Position{})
	fh.IsOutside = false
	w.AddLocation(fh)

	inn := NewLocation("frosthold_inn", "Inn", LocBuilding, "frosthold", Position{})
	inn.IsOutside = false
	w.AddLocation(inn)

	common := NewLocation("frosthold_inn_common", "Common", LocRoom, "frosthold_inn", Position{})
	common.IsOutside = false
	w.AddLocation(common)

	cf := NewLocation("crystal_forest", "CF", LocRegion, "aetheria", Position{X: 500, Y: 100})
	w.AddLocation(cf)

	// same room
	if !w.CanInstantMove("frosthold_inn_common", "frosthold_inn_common") {
		t.Error("same should be instant")
	}
	// siblings under city
	if !w.CanInstantMove("frosthold_inn", "frosthold_market") {
		// market may not exist in this test world, create one
	}
	market := NewLocation("frosthold_market", "Market", LocBuilding, "frosthold", Position{})
	market.IsOutside = false
	w.AddLocation(market)
	if !w.CanInstantMove("frosthold_inn", "frosthold_market") {
		t.Error("same city buildings should be instant")
	}
	// parent/child
	if !w.CanInstantMove("frosthold", "frosthold_inn") {
		t.Error("parent child instant")
	}
	// cross region should NOT be instant
	if w.CanInstantMove("frosthold_inn_common", "crystal_forest") {
		t.Error("cross region should not be instant")
	}
}

func TestRegionOfAndEffectiveWeather(t *testing.T) {
	w := NewWorld()
	a := NewLocation("aetheria", "A", LocWorld, "", Position{})
	a.Weather = NewWeather(Clear, 15)
	w.AddLocation(a)

	nh := NewLocation("northern_highlands", "NH", LocRegion, "aetheria", Position{})
	nh.Weather = NewWeather(Overcast, -5)
	w.AddLocation(nh)

	fh := NewLocation("frosthold", "F", LocCity, "northern_highlands", Position{})
	fh.IsOutside = false
	w.AddLocation(fh)

	wth := w.EffectiveWeather("frosthold")
	if wth == nil || wth.Type != Overcast {
		t.Errorf("expected region weather, got %+v", wth)
	}

	if r := w.RegionOf("frosthold"); r == nil || r.ID != "northern_highlands" {
		t.Error("wrong region")
	}
}

func TestAddBidirectionalExit(t *testing.T) {
	w := NewWorld()
	r1 := NewLocation("r1", "R1", LocRegion, "aetheria", Position{X: 0, Y: 0})
	r2 := NewLocation("r2", "R2", LocRegion, "aetheria", Position{X: 100, Y: 0})
	w.AddLocation(r1)
	w.AddLocation(r2)
	w.AddBidirectionalExit("r1", "r2", "east", "west")
	if len(r1.Exits) != 1 || len(r2.Exits) != 1 {
		t.Error("exits not added")
	}
}

func TestGenerateWeatherForBias(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	w := GenerateWeatherFor(Winter, "northern_highlands", rng)
	if w.Temperature > 10 {
		t.Error("highlands should be cold")
	}
	w2 := GenerateWeatherFor(Summer, "ash_desert", rng)
	if w2.Humidity > 0.6 {
		t.Log("desert humidity low expected, got", w2.Humidity) // not strict
	}
}

func TestTravelTimeWithWeather(t *testing.T) {
	from := NewLocation("a", "A", LocRegion, "", Position{X: 0, Y: 0})
	to := NewLocation("b", "B", LocRegion, "", Position{X: 100, Y: 0})
	normal := TravelTimeWithWeather(from, to, TravelWalk, nil)
	if normal < 1 {
		t.Error("travel time too small")
	}
	storm := &Weather{Type: Thunderstorm}
	slow := TravelTimeWithWeather(from, to, TravelWalk, storm)
	if slow <= normal {
		t.Error("storm should slow travel")
	}
}
