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

// GetQuest returns the details of a specific quest by its ID.
// Currently returns a placeholder "not implemented" message.
func (h *Handler) GetQuest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":      c.Param("id"),
		"message": "quest not implemented",
	})
}

// AcceptQuest allows an entity to accept a quest by its ID.
// Currently returns a placeholder "not implemented" message.
func (h *Handler) AcceptQuest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"quest_id": c.Param("id"),
		"status":   "accepted",
		"message":  "quest system not implemented",
	})
}

// GetEntityQuests returns the list of quests accepted by a specific entity.
// Currently returns an empty list as a placeholder.
func (h *Handler) GetEntityQuests(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{})
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
