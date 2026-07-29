// Package api provides the HTTP API router and handlers for world, entity, and quest data.
package api

import (
	"net/http"
	"sort"
	"strings"

	"simuz/internal/combat"
	"simuz/internal/entity"

	"github.com/gin-gonic/gin"
)

// ListCombats returns a summary of active combat zones (locations with
// multiple hostile factions) along with the most recent combat log entries.
func (h *Handler) ListCombats(c *gin.Context) {
	all := h.Sim.Entities.All()

	type zoneInfo struct {
		LocationID   string   `json:"location_id"`
		LocationName string   `json:"location_name"`
		Factions     []string `json:"factions"`
		EntityCount  int      `json:"entity_count"`
	}

	locFactions := make(map[string]map[string]int)
	for _, e := range all {
		if !e.Alive {
			continue
		}
		if locFactions[e.LocationID] == nil {
			locFactions[e.LocationID] = make(map[string]int)
		}
		locFactions[e.LocationID][e.Faction]++
	}

	var zones []zoneInfo
	for locID, factions := range locFactions {
		if len(factions) < 2 {
			continue
		}
		total := 0
		fnames := make([]string, 0, len(factions))
		for f, c := range factions {
			fnames = append(fnames, f)
			total += c
		}
		loc := h.Sim.World.Location(locID)
		name := locID
		if loc != nil {
			name = loc.Name
		}
		sort.SliceStable(fnames, func(i, j int) bool {
			li := strings.ToLower(fnames[i])
			lj := strings.ToLower(fnames[j])
			if li != lj {
				return li < lj
			}
			return fnames[i] < fnames[j]
		})
		zones = append(zones, zoneInfo{LocationID: locID, LocationName: name, Factions: fnames, EntityCount: total})
	}
	sort.SliceStable(zones, func(i, j int) bool {
		li := strings.ToLower(zones[i].LocationName)
		lj := strings.ToLower(zones[j].LocationName)
		if li != lj {
			return li < lj
		}
		return zones[i].LocationID < zones[j].LocationID
	})

	c.JSON(http.StatusOK, gin.H{
		"zones": zones,
		"log":   combat.RecentLog(50),
	})
}

// GetBestiary returns a list of all known creature species and their stats.
// Currently returns a placeholder "not implemented" message.
func (h *Handler) GetBestiary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "bestiary not implemented",
	})
}

// GetSpecies returns detailed information about a specific creature species.
// Currently returns a placeholder "not implemented" message.
func (h *Handler) GetSpecies(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"species": c.Param("species"),
		"message": "species details not implemented",
	})
}

// GetItems returns a list of all items available in the game world.
// Currently returns an empty list as a placeholder.
func (h *Handler) GetItems(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{})
}

// GetItem returns detailed information about a specific item by its ID.
// Currently returns a placeholder "not implemented" message.
func (h *Handler) GetItem(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":      c.Param("id"),
		"message": "item not implemented",
	})
}

// GetAIArchetypes returns a list of all available AI archetype names that
// entities can be assigned.
func (h *Handler) GetAIArchetypes(c *gin.Context) {
	c.JSON(http.StatusOK, []string{"passive", "aggressive", "territorial", "cowardly", "greedy", "noble", "curious", "guarded", "patrol", "scripted", "dormant"})
}

// GetAIScripts returns a list of all available AI Lua scripts that can be
// assigned to entities. Currently returns an empty list as a placeholder.
func (h *Handler) GetAIScripts(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{})
}

// GetEntityAI returns the AI configuration for a specific entity by its ID.
// Returns a 404 error if the entity is not found.
func (h *Handler) GetEntityAI(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	id := c.Param("id")
	e := h.Sim.Entities.Get(id)
	if e == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
		return
	}
	c.JSON(http.StatusOK, e.AI)
}

// UpdateEntityAI updates the AI configuration for a specific entity by its ID.
// The new AI config is read from the request body. Returns a 404 error if the
// entity is not found, or a 400 error if the request body is invalid.
func (h *Handler) UpdateEntityAI(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	id := c.Param("id")
	e := h.Sim.Entities.Get(id)
	if e == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
		return
	}

	var ai entity.EntityAI
	if err := c.ShouldBindJSON(&ai); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	e.AI = ai
	c.JSON(http.StatusOK, e.AI)
}
