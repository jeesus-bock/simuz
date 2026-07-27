package api

import (
	"simuz/internal/engine"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Sim *engine.Simulation
}

func RegisterRoutes(rg *gin.RouterGroup, sim *engine.Simulation) {
	h := &Handler{Sim: sim}

	rg.GET("/world", h.GetWorld)
	rg.GET("/world/locations/:id", h.GetLocation)
	rg.POST("/world/tick", h.PostTick)

	rg.GET("/entities", h.ListEntities)
	rg.GET("/entities/:id", h.GetEntity)
	rg.POST("/entities", h.CreateEntity)
	rg.POST("/entities/:id/action", h.EntityAction)

	rg.GET("/combat", h.ListCombats)

	rg.GET("/bestiary", h.GetBestiary)
	rg.GET("/bestiary/:species", h.GetSpecies)
	rg.GET("/items", h.GetItems)
	rg.GET("/items/:id", h.GetItem)

	rg.GET("/ai/archetypes", h.GetAIArchetypes)
	rg.GET("/ai/scripts", h.GetAIScripts)

	rg.GET("/entities/:id/ai", h.GetEntityAI)
	rg.PUT("/entities/:id/ai", h.UpdateEntityAI)

	rg.GET("/quests", h.ListQuests)
	rg.GET("/quests/:id", h.GetQuest)
	rg.POST("/quests/:id/accept", h.AcceptQuest)
	rg.GET("/entities/:id/quests", h.GetEntityQuests)

	rg.POST("/admin/generate", h.AdminGenerate)
	rg.GET("/admin/generate/status", h.AdminGenerateStatus)
}
