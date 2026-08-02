// Package web provides the HTTP routes and view wiring for the Simuz web UI.
package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"simuz/internal/engine"

	"github.com/gin-gonic/gin"
)

func percent(a, b int) int {
	if b == 0 {
		return 0
	}
	return a * 100 / b
}

func mul(a, b int) int {
	return a * b
}

func titleStr(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func joinStrings(s []string, sep string) string {
	return strings.Join(s, sep)
}

func mulF(a float64, b float64) float64 {
	return a * b
}

//go:embed templates/*.html
var templateFS embed.FS

//go:embed templates/static
var staticFS embed.FS

func SetupRoutes(router *gin.Engine, sim *engine.Simulation) {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"percent":           percent,
		"mul":               mul,
		"mulF":              mulF,
		"join":              joinStrings,
		"titleStr":          titleStr,
		"renderLocationMap": renderLocationMap,
	}).ParseFS(templateFS, "templates/*.html"))

	h := NewHandler(sim, tmpl, staticFS)

	router.GET("/", h.Dashboard)
	router.POST("/speed", h.SetSpeedPost)
	router.GET("/entities", h.EntitiesPage)
	router.GET("/locations", h.LocationsPage)
	router.GET("/location/:id", h.LocationDetailPage)
	router.GET("/combat", h.CombatPage)
	router.GET("/combat/:location", h.CombatDetailPage)
	router.GET("/entity/:id", h.EntityDetailPage)
	router.GET("/quests", h.QuestsPage)
	router.GET("/quest/:id", h.QuestDetailPage)
	router.POST("/quest/:id/accept", h.AcceptQuestPost)
	router.GET("/events", h.EventsPage)
	router.GET("/ai", h.AIPage)
	router.GET("/pregnancies", h.PregnanciesPage)
	router.GET("/deities", h.DeitiesPage)
	router.GET("/factions", h.FactionsPage)
	router.GET("/species", h.SpeciesPage)
	router.GET("/api/v1/ui/events", h.SSEEvents)
	router.GET("/api/v1/ui/fragments/dashboard", h.DashboardFragment)
	router.GET("/api/v1/ui/fragments/entities", h.EntitiesFragment)
	router.GET("/api/v1/ui/fragments/locations", h.LocationsFragment)
	router.GET("/api/v1/ui/fragments/location/:id", h.LocationDetailFragment)
	router.GET("/api/v1/ui/fragments/combat", h.CombatFragment)
	router.GET("/api/v1/ui/fragments/combat/:location", h.CombatDetailFragment)
	router.GET("/api/v1/ui/fragments/entity/:id", h.EntityDetailFragment)
	router.GET("/api/v1/ui/fragments/quests", h.QuestsFragment)
	router.GET("/api/v1/ui/fragments/quest/:id", h.QuestDetailFragment)
	router.GET("/api/v1/ui/fragments/pregnancies", h.PregnanciesFragment)
	router.GET("/api/v1/ui/fragments/ai", h.AIFragment)
	router.GET("/api/v1/ui/fragments/events", h.EventsFragment)
	router.GET("/api/v1/ui/fragments/deities", h.DeitiesFragment)
	router.GET("/api/v1/ui/fragments/factions", h.FactionsFragment)
	router.GET("/api/v1/ui/fragments/species", h.SpeciesFragment)

	staticSub, _ := fs.Sub(staticFS, "templates/static")
	router.StaticFS("/static", http.FS(staticSub))
}
