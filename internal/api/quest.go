// Package api provides the HTTP API router and handlers for world, entity, and quest data.
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListQuests returns all available quest definitions sorted by ID and title.
func (h *Handler) ListQuests(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	defs := h.Sim.Quests.AllDefs()
	c.JSON(http.StatusOK, defs)
}

// GetQuest returns the details of a specific quest by its ID, including
// the definition and all entities currently pursuing it.
func (h *Handler) GetQuest(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	id := c.Param("id")
	def := h.Sim.Quests.GetDef(id)
	if def == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quest not found"})
		return
	}

	type pursuerInfo struct {
		EntityID string `json:"entity_id"`
		Name     string `json:"name"`
		State    string `json:"state"`
		Stage    string `json:"stage"`
	}

	var pursuers []pursuerInfo
	for _, ent := range h.Sim.Entities.All() {
		for _, state := range h.Sim.Quests.EntityStates(ent.ID) {
			if state.QuestID != id {
				continue
			}
			pursuers = append(pursuers, pursuerInfo{
				EntityID: ent.ID,
				Name:     ent.Name,
				State:    string(state.State),
				Stage:    state.CurrentStage,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"definition": def,
		"pursuers":   pursuers,
	})
}

// AcceptQuest allows an entity to accept a quest by its ID.
// Requires "entity_id" in the request body.
func (h *Handler) AcceptQuest(c *gin.Context) {
	h.Sim.Lock()
	defer h.Sim.Unlock()

	questID := c.Param("id")
	var req struct {
		EntityID string `json:"entity_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ent := h.Sim.Entities.Get(req.EntityID)
	if ent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
		return
	}

	if !h.Sim.Quests.CanAccept(req.EntityID, questID, ent.Level) {
		c.JSON(http.StatusConflict, gin.H{"error": "entity cannot accept this quest (already active, level requirement not met, or quest not found)"})
		return
	}

	ok := h.Sim.Quests.Accept(req.EntityID, questID, ent.Level, h.Sim.Tick)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to accept quest"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"quest_id":   questID,
		"entity_id":  req.EntityID,
		"status":     "accepted",
		"accepted_at": h.Sim.Tick,
	})
}

// GetEntityQuests returns the list of quests accepted by a specific entity.
func (h *Handler) GetEntityQuests(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	id := c.Param("id")
	ent := h.Sim.Entities.Get(id)
	if ent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
		return
	}

	states := h.Sim.Quests.EntityStates(id)
	type questInfo struct {
		QuestID      string `json:"quest_id"`
		Title        string `json:"title"`
		State        string `json:"state"`
		CurrentStage string `json:"current_stage"`
		AcceptedTick uint64 `json:"accepted_tick"`
	}

	result := make([]questInfo, 0, len(states))
	for _, s := range states {
		title := s.QuestID
		if def := h.Sim.Quests.GetDef(s.QuestID); def != nil {
			title = def.Title
		}
		result = append(result, questInfo{
			QuestID:      s.QuestID,
			Title:        title,
			State:        string(s.State),
			CurrentStage: s.CurrentStage,
			AcceptedTick: s.AcceptedTick,
		})
	}

	c.JSON(http.StatusOK, result)
}

// AdminGenerate triggers world generation with the provided parameters.
// Currently returns a placeholder "not implemented" message.
func (h *Handler) AdminGenerate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "generation not implemented",
	})
}

// AdminGenerateStatus returns the current status of world generation.
// Currently returns a placeholder "idle" status.
func (h *Handler) AdminGenerateStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "idle",
	})
}
