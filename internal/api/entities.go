package api

import (
	"net/http"

	"simuz/internal/entity"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListEntities(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	all := h.Sim.Entities.All()
	result := make([]gin.H, len(all))
	for i, e := range all {
		result[i] = entityToSummary(e)
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetEntity(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	id := c.Param("id")
	e := h.Sim.Entities.Get(id)
	if e == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
		return
	}

	c.JSON(http.StatusOK, entityToDetail(e))
}

func (h *Handler) CreateEntity(c *gin.Context) {
	var req struct {
		ID        string `json:"id" binding:"required"`
		Name      string `json:"name" binding:"required"`
		Species   string `json:"species"`
		Level     int    `json:"level"`
		LocationID string `json:"location_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attrs := entity.RandomAttributes(func(n int) int { return h.Sim.RNG.Intn(n) })
	ent := entity.NewEntity(req.ID, req.Name, req.Species, attrs, req.Level)
	if req.LocationID != "" {
		ent.LocationID = req.LocationID
	}

	h.Sim.Entities.Add(ent)

	c.JSON(http.StatusCreated, entityToDetail(ent))
}

func (h *Handler) EntityAction(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	id := c.Param("id")
	ent := h.Sim.Entities.Get(id)
	if ent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
		return
	}

	var req struct {
		Action string `json:"action" binding:"required"`
		Target string `json:"target"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id": id,
		"action":    req.Action,
		"target":    req.Target,
		"status":    "queued",
	})
}

func entityToSummary(e *entity.Entity) gin.H {
	return gin.H{
		"id":          e.ID,
		"name":        e.Name,
		"species":     e.Species,
		"level":       e.Level,
		"alive":       e.Alive,
		"immortal":    e.Immortal,
		"hp":          e.HP,
		"max_hp":      e.MaxHP,
		"location_id": e.LocationID,
		"activity":    e.Activity.Type,
		"ai_type":     e.AI.Type,
		"ai_scripts":  e.AI.ScriptIDs,
	}
}

func entityToDetail(e *entity.Entity) gin.H {
	skills := make(map[string]int)
	if e.Skills != nil {
		for k, v := range e.Skills {
			skills[k] = v
		}
	}
	return gin.H{
		"id":         e.ID,
		"name":       e.Name,
		"species":    e.Species,
		"level":      e.Level,
		"alive":      e.Alive,
		"attributes": e.Attributes,
		"skills":     skills,
		"hp":         e.HP,
		"max_hp":     e.MaxHP,
		"fp":         e.FP,
		"max_fp":     e.MaxFP,
		"xp":         e.XP,
		"location_id": e.LocationID,
		"position":   e.Position,
		"equipment":  e.Equipment,
		"inventory":  e.Inventory,
		"activity":   e.Activity,
		"ai":         e.AI,
		"faction":    e.Faction,
		"encumbrance": e.Encumbrance(),
	}
}
