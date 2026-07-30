// Package world defines the simulation world model, locations, travel rules, and weather systems.
package world

import (
	"math"
)

type LocationType int

const (
	LocWorld LocationType = iota
	LocRegion
	LocCity
	LocBuilding
	LocRoom
	LocRealm
	LocWildSite
	LocSubTerranean
)

func (lt LocationType) String() string {
	switch lt {
	case LocWorld:
		return "world"
	case LocRegion:
		return "region"
	case LocCity:
		return "city"
	case LocBuilding:
		return "building"
	case LocRoom:
		return "room"
	case LocWildSite:
		return "wild_site"
	case LocSubTerranean:
		return "subterranean"
	case LocRealm:
		return "realm"
	default:
		return "unknown"
	}
}

type TravelMode int

const (
	TravelWalk TravelMode = iota
	TravelRide
	TravelSail
	TravelFly
	TravelTeleport
)

func (tm TravelMode) Speed() float64 {
	switch tm {
	case TravelWalk:
		return 5
	case TravelRide:
		return 15
	case TravelSail:
		return 10
	case TravelFly:
		return 30
	case TravelTeleport:
		return 1e9
	default:
		return 5
	}
}

type Exit struct {
	TargetID   string     `json:"target_id"`
	Direction  string     `json:"direction"`
	TravelMode TravelMode `json:"travel_mode"`
	Distance   float64    `json:"distance"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Location struct {
	ID                 string       `json:"id"`
	Name               string       `json:"name"`
	Type               LocationType `json:"type"`
	ParentID           string       `json:"parent_id,omitempty"`
	Children           []string     `json:"children,omitempty"`
	Position           Position     `json:"position"`
	Area               float64      `json:"area"`
	IsOutside          bool         `json:"is_outside"`
	Weather            *Weather     `json:"weather,omitempty"`
	Exits              []Exit       `json:"exits,omitempty"`
	Tags               []string     `json:"tags,omitempty"`
	ControllingFaction string       `json:"controlling_faction,omitempty"`
	ControlStrength    int          `json:"control_strength,omitempty"`
}

func NewLocation(id, name string, locType LocationType, parentID string, pos Position) *Location {
	return &Location{
		ID:        id,
		Name:      name,
		Type:      locType,
		ParentID:  parentID,
		Children:  make([]string, 0),
		Position:  pos,
		IsOutside: locType <= LocCity,
		Tags:      make([]string, 0),
	}
}

func (l *Location) DistanceTo(other *Location) float64 {
	dx := l.Position.X - other.Position.X
	dy := l.Position.Y - other.Position.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func (l *Location) HasTag(tag string) bool {
	for _, t := range l.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
