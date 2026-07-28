// Package api provides the HTTP API router and handlers for world, entity, and quest data.
package api

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetWorld(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	allEntities := h.Sim.Entities.All()
	alive := 0
	for _, e := range allEntities {
		if e.Alive {
			alive++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tick":   h.Sim.Tick,
		"time":   h.Sim.Time.String(),
		"day":    h.Sim.Time.Day,
		"hour":   h.Sim.Time.Hour,
		"minute": h.Sim.Time.Minute,
		"speed":  h.Sim.Time.Speed,
		"phase":  h.Sim.Time.Phase().String(),
		"season": h.Sim.Time.Season().String(),
		"entities": gin.H{
			"total": len(allEntities),
			"alive": alive,
		},
		"locations": len(h.Sim.World.AllLocations()),
	})
}

func (h *Handler) GetLocation(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	id := c.Param("id")
	loc := h.Sim.World.Location(id)
	if loc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
		return
	}

	children := h.Sim.World.ChildLocations(id)
	sort.SliceStable(children, func(i, j int) bool {
		if children[i] == nil {
			return false
		}
		if children[j] == nil {
			return true
		}
		li := strings.ToLower(children[i].Name)
		lj := strings.ToLower(children[j].Name)
		if li != lj {
			return li < lj
		}
		return children[i].ID < children[j].ID
	})
	childIDs := make([]string, len(children))
	for i, ch := range children {
		childIDs[i] = ch.ID
	}

	entities := h.Sim.Entities.ByLocation(id)
	sort.SliceStable(entities, func(i, j int) bool {
		switch {
		case entities[i] == nil:
			return false
		case entities[j] == nil:
			return true
		case entities[i].Alive != entities[j].Alive:
			return entities[i].Alive && !entities[j].Alive
		case entities[i].Conscious != entities[j].Conscious:
			return entities[i].Conscious && !entities[j].Conscious
		default:
			li := strings.ToLower(entities[i].Name)
			lj := strings.ToLower(entities[j].Name)
			if li != lj {
				return li < lj
			}
			return entities[i].ID < entities[j].ID
		}
	})
	entityIDs := make([]string, len(entities))
	for i, e := range entities {
		entityIDs[i] = e.ID
	}

	exits := make([]gin.H, 0, len(loc.Exits))
	for _, ex := range loc.Exits {
		targetName := ex.TargetID
		if target := h.Sim.World.Location(ex.TargetID); target != nil {
			targetName = target.Name
		}
		exits = append(exits, gin.H{
			"target_id":   ex.TargetID,
			"target_name": targetName,
			"direction":   ex.Direction,
			"travel_mode": ex.TravelMode,
			"distance":    ex.Distance,
		})
	}
	sort.SliceStable(exits, func(i, j int) bool {
		li, _ := exits[i]["target_name"].(string)
		lj, _ := exits[j]["target_name"].(string)
		li = strings.ToLower(li)
		lj = strings.ToLower(lj)
		if li != lj {
			return li < lj
		}
		di, _ := exits[i]["direction"].(string)
		dj, _ := exits[j]["direction"].(string)
		if !strings.EqualFold(di, dj) {
			return strings.ToLower(di) < strings.ToLower(dj)
		}
		ti, _ := exits[i]["target_id"].(string)
		tj, _ := exits[j]["target_id"].(string)
		return ti < tj
	})

	c.JSON(http.StatusOK, gin.H{
		"id":         loc.ID,
		"name":       loc.Name,
		"type":       loc.Type.String(),
		"parent_id":  loc.ParentID,
		"children":   childIDs,
		"position":   loc.Position,
		"area":       loc.Area,
		"is_outside": loc.IsOutside,
		"weather":    loc.Weather,
		"exits":      exits,
		"entities":   entityIDs,
	})
}

func (h *Handler) PostTick(c *gin.Context) {
	var req struct {
		Count int `json:"count" form:"count"`
	}
	if err := c.ShouldBind(&req); err != nil || req.Count < 1 {
		req.Count = 1
	}
	if req.Count > 1000 {
		req.Count = 1000
	}

	for i := 0; i < req.Count; i++ {
		h.Sim.TickOnce()
	}

	c.JSON(http.StatusOK, gin.H{
		"ticks": req.Count,
		"tick":  h.Sim.Tick,
		"time":  h.Sim.Time.String(),
	})
}
