package world

import (
	"math"
	"sort"
	"strings"
)

type LocationType int

const (
	LocWorld LocationType = iota
	LocRegion
	LocCity
	LocDistrict
	LocBuilding
	LocRoom
	LocRealm
)

func (lt LocationType) String() string {
	switch lt {
	case LocWorld:
		return "world"
	case LocRegion:
		return "region"
	case LocCity:
		return "city"
	case LocDistrict:
		return "district"
	case LocBuilding:
		return "building"
	case LocRoom:
		return "room"
	case LocRealm:
		return "realm"
	default:
		return "unknown"
	}
}

func ParseLocationType(s string) LocationType {
	switch s {
	case "world":
		return LocWorld
	case "region":
		return LocRegion
	case "city":
		return LocCity
	case "district":
		return LocDistrict
	case "building":
		return LocBuilding
	case "room":
		return LocRoom
	case "realm":
		return LocRealm
	default:
		return LocRoom
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

type World struct {
	locations map[string]*Location
	rootID    string
}

func NewWorld() *World {
	return &World{
		locations: make(map[string]*Location),
	}
}

func (w *World) AddLocation(loc *Location) {
	w.locations[loc.ID] = loc
	if loc.ParentID != "" {
		if parent, ok := w.locations[loc.ParentID]; ok {
			parent.Children = append(parent.Children, loc.ID)
		}
	}
	if w.rootID == "" && loc.Type == LocWorld {
		w.rootID = loc.ID
	}
}

func (w *World) Location(id string) *Location {
	return w.locations[id]
}

func (w *World) RootLocation() *Location {
	return w.locations[w.rootID]
}

func (w *World) AllLocations() []*Location {
	result := make([]*Location, 0, len(w.locations))
	for _, loc := range w.locations {
		result = append(result, loc)
	}
	return result
}

func (w *World) ChildLocations(parentID string) []*Location {
	parent := w.locations[parentID]
	if parent == nil {
		return nil
	}
	result := make([]*Location, 0, len(parent.Children))
	for _, cid := range parent.Children {
		if child := w.locations[cid]; child != nil {
			result = append(result, child)
		}
	}
	return result
}

func (w *World) Route(fromID, toID string) []string {
	if fromID == "" || toID == "" {
		return nil
	}
	if fromID == toID {
		return []string{fromID}
	}
	if w.locations[fromID] == nil || w.locations[toID] == nil {
		return nil
	}

	type node struct {
		id   string
		dist int
	}

	queue := []node{{id: fromID, dist: 0}}
	prev := map[string]string{fromID: ""}
	visited := map[string]bool{fromID: true}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.id == toID {
			break
		}

		neighbors := w.adjacentLocationIDs(cur.id)
		for _, next := range neighbors {
			if visited[next] {
				continue
			}
			visited[next] = true
			prev[next] = cur.id
			queue = append(queue, node{id: next, dist: cur.dist + 1})
		}
	}

	if !visited[toID] {
		return nil
	}

	var route []string
	for cur := toID; cur != ""; cur = prev[cur] {
		route = append(route, cur)
		if cur == fromID {
			break
		}
	}
	if len(route) == 0 || route[len(route)-1] != fromID {
		return nil
	}
	for i, j := 0, len(route)-1; i < j; i, j = i+1, j-1 {
		route[i], route[j] = route[j], route[i]
	}
	return route
}

func (w *World) adjacentLocationIDs(id string) []string {
	loc := w.locations[id]
	if loc == nil {
		return nil
	}

	type entry struct {
		id   string
		name string
	}

	var entries []entry
	seen := make(map[string]struct{})
	add := func(nextID string) {
		if nextID == "" || nextID == id {
			return
		}
		if _, ok := seen[nextID]; ok {
			return
		}
		if next := w.locations[nextID]; next != nil {
			seen[nextID] = struct{}{}
			entries = append(entries, entry{id: nextID, name: next.Name})
		}
	}

	if loc.ParentID != "" {
		add(loc.ParentID)
	}
	for _, childID := range loc.Children {
		add(childID)
	}
	for _, ex := range loc.Exits {
		add(ex.TargetID)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		li := strings.ToLower(entries[i].name)
		lj := strings.ToLower(entries[j].name)
		if li != lj {
			return li < lj
		}
		return entries[i].id < entries[j].id
	})

	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.id)
	}
	return out
}

func (w *World) IsInside(id string) bool {
	loc := w.locations[id]
	if loc == nil {
		return false
	}
	return !loc.IsOutside
}

func (w *World) AncestorOfType(id string, t LocationType) *Location {
	cur := w.locations[id]
	for cur != nil {
		if cur.Type == t {
			return cur
		}
		if cur.ParentID == "" {
			break
		}
		cur = w.locations[cur.ParentID]
	}
	return nil
}

func (w *World) RegionOf(id string) *Location {
	if loc := w.locations[id]; loc != nil && loc.Type == LocRegion {
		return loc
	}
	return w.AncestorOfType(id, LocRegion)
}

func (w *World) EffectiveWeather(id string) *Weather {
	cur := w.locations[id]
	for cur != nil {
		if cur.Weather != nil {
			return cur.Weather
		}
		if cur.ParentID == "" {
			break
		}
		cur = w.locations[cur.ParentID]
	}
	return nil
}

func (w *World) AddExit(fromID, toID, direction string, mode TravelMode, distance float64) {
	from := w.locations[fromID]
	to := w.locations[toID]
	if from == nil || to == nil {
		return
	}
	if distance <= 0 {
		distance = from.DistanceTo(to)
		if distance < 1 {
			distance = 1
		}
	}
	for _, e := range from.Exits {
		if e.TargetID == toID {
			return
		}
	}
	from.Exits = append(from.Exits, Exit{
		TargetID:   toID,
		Direction:  direction,
		TravelMode: mode,
		Distance:   distance,
	})
}

func (w *World) AddBidirectionalExit(a, b, dirAB, dirBA string) {
	la, lb := w.locations[a], w.locations[b]
	if la == nil || lb == nil {
		return
	}
	dist := la.DistanceTo(lb)
	if dist < 1 {
		dist = 1
	}
	w.AddExit(a, b, dirAB, TravelWalk, dist)
	w.AddExit(b, a, dirBA, TravelWalk, dist)
}

func (w *World) SameRegion(a, b string) bool {
	ra, rb := w.RegionOf(a), w.RegionOf(b)
	if ra == nil || rb == nil {
		return a == b
	}
	return ra.ID == rb.ID
}

func (w *World) CanInstantMove(fromID, toID string) bool {
	if fromID == toID {
		return true
	}
	from, to := w.locations[fromID], w.locations[toID]
	if from == nil || to == nil {
		return false
	}
	if from.ParentID == toID || to.ParentID == fromID {
		return true
	}
	if from.ParentID != "" && from.ParentID == to.ParentID {
		return true
	}
	if ca := w.AncestorOfType(fromID, LocCity); ca != nil {
		if cb := w.AncestorOfType(toID, LocCity); cb != nil && ca.ID == cb.ID {
			return true
		}
	}
	return w.SameRegion(fromID, toID)
}

func (l *Location) HasTag(tag string) bool {
	for _, t := range l.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
