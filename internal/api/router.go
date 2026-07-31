// Package api provides the HTTP API router and handlers for world, entity, and quest data.
package api

import (
	"simuz/internal/engine"

	"github.com/gin-gonic/gin"
)

// Handler holds references to the simulation and template engine needed
// by all API route handlers.
type Handler struct {
	// Sim is a pointer to the active simulation instance, providing access
	// to the world, entities, quests, and time state.
	Sim *engine.Simulation
}

// RegisterRoutes registers all API endpoints under the provided gin RouterGroup.
// Each route is bound to its corresponding handler method on the Handler struct.
func RegisterRoutes(rg *gin.RouterGroup, sim *engine.Simulation) {
	h := &Handler{Sim: sim}

	rg.GET("/world", h.GetWorld)
	rg.GET("/world/locations/:id", h.GetLocation)
	rg.POST("/world/tick", h.PostTick)
	rg.PUT("/world/speed", h.SetSpeed)

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
