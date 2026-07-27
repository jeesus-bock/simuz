package api

import (
	"net/http"

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
	childIDs := make([]string, len(children))
	for i, ch := range children {
		childIDs[i] = ch.ID
	}

	entities := h.Sim.Entities.ByLocation(id)
	entityIDs := make([]string, len(entities))
	for i, e := range entities {
		entityIDs[i] = e.ID
	}

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
		"exits":      loc.Exits,
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
