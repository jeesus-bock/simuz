package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListQuests(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	defs := h.Sim.Quests.AllDefs()
	c.JSON(http.StatusOK, defs)
}

func (h *Handler) GetQuest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":      c.Param("id"),
		"message": "quest not implemented",
	})
}

func (h *Handler) AcceptQuest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"quest_id": c.Param("id"),
		"status":   "accepted",
		"message":  "quest system not implemented",
	})
}

func (h *Handler) GetEntityQuests(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{})
}

func (h *Handler) AdminGenerate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "generation not implemented",
	})
}

func (h *Handler) AdminGenerateStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "idle",
	})
}
