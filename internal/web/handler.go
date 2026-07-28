// Package web provides HTTP handlers and template rendering for the Simuz web UI.
package web

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"io/fs"
	"sort"
	"strings"

	"simuz/internal/combat"
	"simuz/internal/engine"
	"simuz/internal/entity"
	"simuz/internal/events"
	"simuz/internal/items"
	"simuz/internal/quest"
	"simuz/internal/world"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Sim    *engine.Simulation
	Tmpls  *template.Template
	Static fs.FS
}

func NewHandler(sim *engine.Simulation, tmpls *template.Template, static fs.FS) *Handler {
	return &Handler{
		Sim:    sim,
		Tmpls:  tmpls,
		Static: static,
	}
}

func (h *Handler) Dashboard(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	if err := h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":     "Simuz",
		"page":      "dashboard",
		"tick":      h.Sim.Tick,
		"time":      h.Sim.Time.String(),
		"phase":     h.Sim.Time.Phase().String(),
		"season":    h.Sim.Time.Season().String(),
		"entities":  len(h.Sim.Entities.All()),
		"locations": len(h.Sim.World.AllLocations()),
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) EntitiesPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	all := h.Sim.Entities.All()
	sortEntitiesForDisplay(all)
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":         "Entities",
		"page":          "entities",
		"tick":          h.Sim.Tick,
		"time":          h.Sim.Time.String(),
		"phase":         h.Sim.Time.Phase().String(),
		"season":        h.Sim.Time.Season().String(),
		"entities":      len(all),
		"entities_list": all,
		"locations":     len(h.Sim.World.AllLocations()),
	}); err != nil {
		_ = c.Error(err)
	}
}

type locNode struct {
	Location    *world.Location
	Children    []locNode
	Depth       int
	EntityCount int
}

type exitView struct {
	TargetID   string
	TargetName string
	Direction  string
	Distance   float64
}

type combatantView struct {
	ID         string
	Name       string
	Species    string
	Faction    string
	HP         int
	MaxHP      int
	WeaponName string
	State      string
	StateRank  int
	StateLabel string
	StateIcon  string
	StateClass string
}

type combatGroup struct {
	Faction    string
	Label      string
	Combatants []combatantView
}

type questProgressView struct {
	Def    *quest.QuestDef
	Entity *entity.Entity
	State  *quest.EntityQuestState
}

func (h *Handler) activeQuestViews() []questProgressView {
	var active []questProgressView
	for _, ent := range h.Sim.Entities.All() {
		for _, state := range h.Sim.Quests.EntityStates(ent.ID) {
			if state.State != quest.StateActive {
				continue
			}
			if def := h.Sim.Quests.GetDef(state.QuestID); def != nil {
				active = append(active, questProgressView{Def: def, Entity: ent, State: state})
			}
		}
	}
	sort.SliceStable(active, func(i, j int) bool {
		if active[i].Entity.Name != active[j].Entity.Name {
			return active[i].Entity.Name < active[j].Entity.Name
		}
		return active[i].State.AcceptedTick < active[j].State.AcceptedTick
	})
	return active
}

func entityDisplayRank(e *entity.Entity) int {
	switch {
	case e == nil:
		return 3
	case !e.Alive:
		return 2
	case !e.Conscious:
		return 1
	default:
		return 0
	}
}

func sortEntitiesForDisplay(entities []*entity.Entity) {
	sort.SliceStable(entities, func(i, j int) bool {
		ri := entityDisplayRank(entities[i])
		rj := entityDisplayRank(entities[j])
		if ri != rj {
			return ri < rj
		}
		ni := strings.ToLower(entities[i].Name)
		nj := strings.ToLower(entities[j].Name)
		if !strings.EqualFold(entities[i].Name, entities[j].Name) {
			return ni < nj
		}
		return entities[i].ID < entities[j].ID
	})
}

func sortLocationsForDisplay(locs []*world.Location) {
	sort.SliceStable(locs, func(i, j int) bool {
		if locs[i] == nil {
			return false
		}
		if locs[j] == nil {
			return true
		}
		ni := strings.ToLower(locs[i].Name)
		nj := strings.ToLower(locs[j].Name)
		if !strings.EqualFold(locs[i].Name, locs[j].Name) {
			return ni < nj
		}
		return locs[i].ID < locs[j].ID
	})
}

func sortInventoryForDisplay(items []items.ItemInstance) {
	sort.SliceStable(items, func(i, j int) bool {
		ni := ""
		nj := ""
		if items[i].Def != nil {
			ni = strings.ToLower(items[i].Def.Name)
		}
		if items[j].Def != nil {
			nj = strings.ToLower(items[j].Def.Name)
		}
		if !strings.EqualFold(ni, nj) {
			return ni < nj
		}
		di := items[i].DefID
		dj := items[j].DefID
		if di != dj {
			return di < dj
		}
		return items[i].ID < items[j].ID
	})
}

func sortEffectsForDisplay(effects []entity.ActiveEffect) {
	sort.SliceStable(effects, func(i, j int) bool {
		if !strings.EqualFold(effects[i].Name, effects[j].Name) {
			return strings.ToLower(effects[i].Name) < strings.ToLower(effects[j].Name)
		}
		if effects[i].StartTick != effects[j].StartTick {
			return effects[i].StartTick < effects[j].StartTick
		}
		return effects[i].ID < effects[j].ID
	})
}

func sortCombatZonesForDisplay(zones []combatZone) {
	sort.SliceStable(zones, func(i, j int) bool {
		ni := strings.ToLower(zones[i].LocationName)
		nj := strings.ToLower(zones[j].LocationName)
		if !strings.EqualFold(zones[i].LocationName, zones[j].LocationName) {
			return ni < nj
		}
		return zones[i].LocationID < zones[j].LocationID
	})
}

func hostileFactionMixExists(factions map[string]int) bool {
	if len(factions) < 2 {
		return false
	}
	keys := make([]string, 0, len(factions))
	for faction := range factions {
		keys = append(keys, faction)
	}
	sortStringsByFoldAndRaw(keys)
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			if combat.Relation(keys[i], keys[j]) == combat.Hostile {
				return true
			}
		}
	}
	return false
}

func relationLabel(rel combat.FactionRelation) string {
	switch rel {
	case combat.Friendly:
		return "friendly"
	case combat.Hostile:
		return "hostile"
	default:
		return "neutral"
	}
}

func buildFactionRelationNotes(factions []string) []string {
	if len(factions) < 2 {
		return nil
	}
	notes := make([]string, 0, len(factions))
	for i := range factions {
		for j := i + 1; j < len(factions); j++ {
			rel := combat.Relation(factions[i], factions[j])
			if rel == combat.Hostile {
				continue
			}
			notes = append(notes, factions[i]+" + "+factions[j]+" ("+relationLabel(rel)+")")
		}
	}
	sortStringsByFoldAndRaw(notes)
	return notes
}

func sortExitsForDisplay(exits []exitView) {
	sort.SliceStable(exits, func(i, j int) bool {
		ni := strings.ToLower(exits[i].TargetName)
		nj := strings.ToLower(exits[j].TargetName)
		if !strings.EqualFold(exits[i].TargetName, exits[j].TargetName) {
			return ni < nj
		}
		if exits[i].Direction != exits[j].Direction {
			return strings.ToLower(exits[i].Direction) < strings.ToLower(exits[j].Direction)
		}
		return exits[i].TargetID < exits[j].TargetID
	})
}

func sortStringsByFoldAndRaw(values []string) {
	sort.SliceStable(values, func(i, j int) bool {
		li := strings.ToLower(values[i])
		lj := strings.ToLower(values[j])
		if !strings.EqualFold(values[i], values[j]) {
			return li < lj
		}
		return values[i] < values[j]
	})
}

func sortTravelersForDisplay(travelers []travelerView) {
	sort.SliceStable(travelers, func(i, j int) bool {
		ni := strings.ToLower(travelers[i].Name)
		nj := strings.ToLower(travelers[j].Name)
		if !strings.EqualFold(travelers[i].Name, travelers[j].Name) {
			return ni < nj
		}
		return travelers[i].EntityID < travelers[j].EntityID
	})
}

func combatantState(e *entity.Entity) (state, label, icon, class string) {
	if e == nil {
		return "unknown", "unknown", "?", "combat-unknown"
	}
	switch {
	case !e.Alive:
		return "dead", "dead", "☠", "combat-dead"
	case !e.Conscious:
		return "knocked_out", "knocked out", "✖", "combat-knocked"
	default:
		return "active", "active", "●", "combat-active"
	}
}

func combatantWeaponName(e *entity.Entity) string {
	if e == nil || e.Equipment.Weapon == nil || e.Equipment.Weapon.Def == nil {
		return "fists"
	}
	return e.Equipment.Weapon.Def.Name
}

func buildCombatantView(e *entity.Entity) combatantView {
	state, label, icon, class := combatantState(e)
	rank := 0
	switch state {
	case "active":
		rank = 0
	case "knocked_out":
		rank = 1
	case "dead":
		rank = 2
	default:
		rank = 3
	}
	faction := e.Faction
	if faction == "" {
		faction = "unknown"
	}
	return combatantView{
		ID:         e.ID,
		Name:       e.Name,
		Species:    e.Species,
		Faction:    faction,
		HP:         e.HP,
		MaxHP:      e.MaxHP,
		WeaponName: combatantWeaponName(e),
		State:      state,
		StateRank:  rank,
		StateLabel: label,
		StateIcon:  icon,
		StateClass: class,
	}
}

func buildCombatGroups(entities []*entity.Entity) (active []combatGroup, downed []combatGroup, activeCount, downedCount int) {
	activeByFaction := make(map[string][]combatantView)
	downedByFaction := make(map[string][]combatantView)

	for _, e := range entities {
		view := buildCombatantView(e)
		switch view.State {
		case "active":
			activeByFaction[view.Faction] = append(activeByFaction[view.Faction], view)
			activeCount++
		default:
			downedByFaction[view.Faction] = append(downedByFaction[view.Faction], view)
			downedCount++
		}
	}

	sortedFactionKeys := func(m map[string][]combatantView) []string {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return strings.ToLower(keys[i]) < strings.ToLower(keys[j])
		})
		return keys
	}

	sortViews := func(vs []combatantView) {
		sort.SliceStable(vs, func(i, j int) bool {
			if vs[i].StateRank != vs[j].StateRank {
				return vs[i].StateRank < vs[j].StateRank
			}
			if !strings.EqualFold(vs[i].Name, vs[j].Name) {
				return strings.ToLower(vs[i].Name) < strings.ToLower(vs[j].Name)
			}
			return vs[i].ID < vs[j].ID
		})
	}

	for _, faction := range sortedFactionKeys(activeByFaction) {
		vs := activeByFaction[faction]
		sortViews(vs)
		active = append(active, combatGroup{
			Faction:    faction,
			Label:      titleStr(faction),
			Combatants: vs,
		})
	}

	for _, faction := range sortedFactionKeys(downedByFaction) {
		vs := downedByFaction[faction]
		sortViews(vs)
		downed = append(downed, combatGroup{
			Faction:    faction,
			Label:      titleStr(faction),
			Combatants: vs,
		})
	}

	return active, downed, activeCount, downedCount
}

func combatDetailData(h *Handler, locID string) gin.H {
	loc := h.Sim.World.Location(locID)
	name := locID
	if loc != nil {
		name = loc.Name
	}

	all := h.Sim.Entities.All()
	var combatants []*entity.Entity
	for _, e := range all {
		if e.LocationID == locID {
			combatants = append(combatants, e)
		}
	}

	activeGroups, downedGroups, activeCount, downedCount := buildCombatGroups(combatants)
	events := combat.LocationEvents(locID, 100)

	return gin.H{
		"loc_id":           locID,
		"loc_name":         name,
		"tick":             h.Sim.Tick,
		"time":             h.Sim.Time.String(),
		"combatants_total": len(combatants),
		"active_count":     activeCount,
		"downed_count":     downedCount,
		"active_groups":    activeGroups,
		"downed_groups":    downedGroups,
		"combat_events":    events,
	}
}

func buildLocationTree(w *world.World, entities *entity.Manager) []locNode {
	root := w.RootLocation()
	if root == nil {
		return nil
	}
	return buildChildren(w, entities, root.ID, 0)
}

func buildChildren(w *world.World, entities *entity.Manager, parentID string, depth int) []locNode {
	children := w.ChildLocations(parentID)
	sortLocationsForDisplay(children)
	var nodes []locNode
	for _, loc := range children {
		ec := len(entities.ByLocation(loc.ID))
		node := locNode{
			Location:    loc,
			Depth:       depth,
			EntityCount: ec,
			Children:    buildChildren(w, entities, loc.ID, depth+1),
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func (h *Handler) LocationsPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	all := h.Sim.Entities.All()

	var realms []*world.Location
	for _, loc := range h.Sim.World.AllLocations() {
		if loc.Type == world.LocRealm {
			realms = append(realms, loc)
		}
	}
	sortLocationsForDisplay(realms)

	if err := h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":     "Locations",
		"page":      "locations",
		"tick":      h.Sim.Tick,
		"time":      h.Sim.Time.String(),
		"phase":     h.Sim.Time.Phase().String(),
		"season":    h.Sim.Time.Season().String(),
		"entities":  len(all),
		"locations": len(h.Sim.World.AllLocations()),
		"loc_tree":  buildLocationTree(h.Sim.World, h.Sim.Entities),
		"travelers": buildTravelerViews(h.Sim, ""),
		"root_name": h.Sim.World.RootLocation().Name,
		"realms":    realms,
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) CombatPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	all := h.Sim.Entities.All()
	activeZones, contestedZones := buildCombatOverviewZones(h.Sim)

	if err := h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":           "Combat",
		"page":            "combat",
		"tick":            h.Sim.Tick,
		"time":            h.Sim.Time.String(),
		"phase":           h.Sim.Time.Phase().String(),
		"season":          h.Sim.Time.Season().String(),
		"entities":        len(all),
		"locations":       len(h.Sim.World.AllLocations()),
		"combat_zones":    activeZones,
		"contested_zones": contestedZones,
		"combat_log":      combat.RecentLog(50),
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) QuestsPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	defs := h.Sim.Quests.AllDefs()
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":         "Quests",
		"page":          "quests",
		"tick":          h.Sim.Tick,
		"time":          h.Sim.Time.String(),
		"phase":         h.Sim.Time.Phase().String(),
		"season":        h.Sim.Time.Season().String(),
		"entities":      len(h.Sim.Entities.All()),
		"locations":     len(h.Sim.World.AllLocations()),
		"quest_defs":    defs,
		"active_quests": h.activeQuestViews(),
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) AIPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":     "AI",
		"page":      "ai",
		"tick":      h.Sim.Tick,
		"time":      h.Sim.Time.String(),
		"phase":     h.Sim.Time.Phase().String(),
		"season":    h.Sim.Time.Season().String(),
		"entities":  len(h.Sim.Entities.All()),
		"locations": len(h.Sim.World.AllLocations()),
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) EventsPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	evList := h.Sim.Events.RecentEvents(100)
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":     "Events",
		"page":      "events",
		"tick":      h.Sim.Tick,
		"time":      h.Sim.Time.String(),
		"phase":     h.Sim.Time.Phase().String(),
		"season":    h.Sim.Time.Season().String(),
		"entities":  len(h.Sim.Entities.All()),
		"locations": len(h.Sim.World.AllLocations()),
		"events":    evList,
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) EventsFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	evList := h.Sim.Events.RecentEvents(50)
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "events_list", gin.H{
		"events": evList,
		"tick":   h.Sim.Tick,
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) DashboardFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "dashboard_stats", gin.H{
		"tick":      h.Sim.Tick,
		"time":      h.Sim.Time.String(),
		"phase":     h.Sim.Time.Phase().String(),
		"season":    h.Sim.Time.Season().String(),
		"entities":  len(h.Sim.Entities.All()),
		"locations": len(h.Sim.World.AllLocations()),
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) EntitiesFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	all := h.Sim.Entities.All()
	sortEntitiesForDisplay(all)
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "entities_table", gin.H{
		"tick":          h.Sim.Tick,
		"time":          h.Sim.Time.String(),
		"entities":      len(all),
		"entities_list": all,
		"locations":     len(h.Sim.World.AllLocations()),
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) LocationsFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	all := h.Sim.Entities.All()
	var realms []*world.Location
	for _, loc := range h.Sim.World.AllLocations() {
		if loc.Type == world.LocRealm {
			realms = append(realms, loc)
		}
	}
	sortLocationsForDisplay(realms)
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "locations_tree", gin.H{
		"tick":      h.Sim.Tick,
		"time":      h.Sim.Time.String(),
		"entities":  len(all),
		"locations": len(h.Sim.World.AllLocations()),
		"loc_tree":  buildLocationTree(h.Sim.World, h.Sim.Entities),
		"travelers": buildTravelerViews(h.Sim, ""),
		"root_name": h.Sim.World.RootLocation().Name,
		"realms":    realms,
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) locationDetailData(locID string) (gin.H, bool) {
	loc := h.Sim.World.Location(locID)
	if loc == nil {
		return nil, false
	}
	parentName := ""
	if loc.ParentID != "" {
		if p := h.Sim.World.Location(loc.ParentID); p != nil {
			parentName = p.Name
		} else {
			parentName = loc.ParentID
		}
	}
	ents := h.Sim.Entities.ByLocation(locID)
	sortEntitiesForDisplay(ents)
	type entRow struct {
		ID, Name, Species, Faction string
		Level, HP, MaxHP           int
		Alive                      bool
	}
	var rows []entRow
	for _, e := range ents {
		rows = append(rows, entRow{
			ID: e.ID, Name: e.Name, Species: e.Species, Faction: e.Faction,
			Level: e.Level, HP: e.HP, MaxHP: e.MaxHP, Alive: e.Alive,
		})
	}
	wth := h.Sim.World.EffectiveWeather(locID)
	children := h.Sim.World.ChildLocations(locID)
	sortLocationsForDisplay(children)
	events := combat.LocationEvents(locID, 30)
	travelers := buildTravelerViews(h.Sim, locID)
	exits := buildLocationExits(h.Sim.World, loc)
	return gin.H{
		"title":         "Location: " + loc.Name,
		"page":          "location_detail",
		"tick":          h.Sim.Tick,
		"time":          h.Sim.Time.String(),
		"phase":         h.Sim.Time.Phase().String(),
		"season":        h.Sim.Time.Season().String(),
		"entities":      len(h.Sim.Entities.All()),
		"locations":     len(h.Sim.World.AllLocations()),
		"loc":           loc,
		"parent_name":   parentName,
		"exits":         exits,
		"entity_rows":   rows,
		"weather":       wth,
		"children":      children,
		"combat_events": events,
		"travelers":     travelers,
	}, true
}

func buildLocationExits(w *world.World, loc *world.Location) []exitView {
	if loc == nil || len(loc.Exits) == 0 {
		return nil
	}
	exits := make([]exitView, 0, len(loc.Exits))
	for _, ex := range loc.Exits {
		targetName := ex.TargetID
		if target := w.Location(ex.TargetID); target != nil {
			targetName = target.Name
		}
		exits = append(exits, exitView{
			TargetID:   ex.TargetID,
			TargetName: targetName,
			Direction:  ex.Direction,
			Distance:   ex.Distance,
		})
	}
	sortExitsForDisplay(exits)
	return exits
}

func (h *Handler) LocationDetailPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	data, ok := h.locationDetailData(c.Param("id"))
	if !ok {
		c.String(404, "Location not found")
		return
	}
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "base.html", data); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) LocationDetailFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	data, ok := h.locationDetailData(c.Param("id"))
	if !ok {
		c.String(404, "Location not found")
		return
	}
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "location_detail_status", data); err != nil {
		_ = c.Error(err)
	}
}

type combatZone struct {
	LocationName  string
	LocationID    string
	Factions      []string
	RelationNotes []string
	EntityCount   int
	Href          string
}

func (h *Handler) CombatFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	all := h.Sim.Entities.All()
	activeZones, contestedZones := buildCombatOverviewZones(h.Sim)

	if err := h.Tmpls.ExecuteTemplate(c.Writer, "combat_status", gin.H{
		"tick":            h.Sim.Tick,
		"time":            h.Sim.Time.String(),
		"entities":        len(all),
		"locations":       len(h.Sim.World.AllLocations()),
		"combat_zones":    activeZones,
		"contested_zones": contestedZones,
		"combat_log":      combat.RecentLog(50),
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) QuestsFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	defs := h.Sim.Quests.AllDefs()
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "quests_list", gin.H{
		"tick":          h.Sim.Tick,
		"time":          h.Sim.Time.String(),
		"entities":      len(h.Sim.Entities.All()),
		"locations":     len(h.Sim.World.AllLocations()),
		"quest_defs":    defs,
		"active_quests": h.activeQuestViews(),
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) AIFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	all := h.Sim.Entities.All()
	sortEntitiesForDisplay(all)
	var activeEntities []*entity.Entity
	for _, e := range all {
		if e.AI.Type == "scripted" || e.AI.Type == "aggressive" || e.AI.Type == "dormant" {
			activeEntities = append(activeEntities, e)
		}
	}
	sortEntitiesForDisplay(activeEntities)
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "ai_status", gin.H{
		"tick":          h.Sim.Tick,
		"time":          h.Sim.Time.String(),
		"entities":      len(all),
		"locations":     len(h.Sim.World.AllLocations()),
		"entities_list": activeEntities,
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) CombatDetailPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	locID := c.Param("location")
	all := h.Sim.Entities.All()
	data := combatDetailData(h, locID)
	data["title"] = "Combat: " + data["loc_name"].(string)
	data["page"] = "combat_detail"
	data["phase"] = h.Sim.Time.Phase().String()
	data["season"] = h.Sim.Time.Season().String()
	data["entities"] = len(all)
	data["locations"] = len(h.Sim.World.AllLocations())

	if err := h.Tmpls.ExecuteTemplate(c.Writer, "base.html", data); err != nil {
		_ = c.Error(err)
	}
}

func buildCombatOverviewZones(sim *engine.Simulation) (activeZones []combatZone, contestedZones []combatZone) {
	if sim == nil {
		return nil, nil
	}
	locFactions := make(map[string]map[string]int)
	for _, e := range sim.Entities.All() {
		if !e.Alive {
			continue
		}
		if locFactions[e.LocationID] == nil {
			locFactions[e.LocationID] = make(map[string]int)
		}
		locFactions[e.LocationID][e.Faction]++
	}

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
		sortStringsByFoldAndRaw(fnames)
		loc := sim.World.Location(locID)
		name := locID
		if loc != nil {
			name = loc.Name
		}
		zone := combatZone{
			LocationName:  name,
			LocationID:    locID,
			Factions:      fnames,
			RelationNotes: buildFactionRelationNotes(fnames),
			EntityCount:   total,
		}
		if hostileFactionMixExists(factions) {
			zone.Href = "/combat/" + locID
			activeZones = append(activeZones, zone)
		} else {
			zone.Href = "/location/" + locID
			contestedZones = append(contestedZones, zone)
		}
	}

	sortCombatZonesForDisplay(activeZones)
	sortCombatZonesForDisplay(contestedZones)
	return activeZones, contestedZones
}

func (h *Handler) CombatDetailFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	locID := c.Param("location")
	if err := h.Tmpls.ExecuteTemplate(c.Writer, "combat_detail_status", combatDetailData(h, locID)); err != nil {
		_ = c.Error(err)
	}
}

type equipSlot struct {
	Slot string
	Item *items.ItemInstance
}

type travelStepView struct {
	ID      string
	Name    string
	Current bool
}

type travelRouteView struct {
	FromID      string
	ToID        string
	Steps       []travelStepView
	CurrentStep int
	TotalSteps  int
	Progress    int
}

func buildTravelRouteView(sim *engine.Simulation, ent *entity.Entity) *travelRouteView {
	if sim == nil || ent == nil {
		return nil
	}
	ts := sim.TravelState(ent.ID)
	if ts == nil || ts.Status != world.TravelInProgress {
		return nil
	}
	route := ts.Route
	if len(route) == 0 {
		route = []string{ts.FromID, ts.ToID}
	}
	if len(route) == 0 {
		return nil
	}
	currentIdx := max(ts.RouteIndex, 0)
	if currentIdx >= len(route) {
		currentIdx = len(route) - 1
	}
	steps := make([]travelStepView, 0, len(route))
	for i, id := range route {
		name := id
		if loc := sim.World.Location(id); loc != nil {
			name = loc.Name
		}
		steps = append(steps, travelStepView{
			ID:      id,
			Name:    name,
			Current: i == currentIdx,
		})
	}
	progress := 0
	if len(route) > 1 {
		progress = currentIdx * 100 / (len(route) - 1)
	}
	return &travelRouteView{
		FromID:      ts.FromID,
		ToID:        ts.ToID,
		Steps:       steps,
		CurrentStep: currentIdx,
		TotalSteps:  len(route) - 1,
		Progress:    progress,
	}
}

type travelerView struct {
	EntityID    string
	Name        string
	From        string
	To          string
	Progress    int
	Eta         int
	Route       []travelStepView
	CurrentStep int
	TotalSteps  int
}

func buildTravelerViews(sim *engine.Simulation, locID string) []travelerView {
	if sim == nil {
		return nil
	}
	var travelers []travelerView
	if sim.Traveling == nil {
		return travelers
	}
	for _, ts := range sim.Traveling {
		if ts == nil || ts.Status != world.TravelInProgress {
			continue
		}
		if locID != "" && ts.FromID != locID && ts.ToID != locID {
			continue
		}
		name := ts.EntityID
		if e := sim.Entities.Get(ts.EntityID); e != nil {
			name = e.Name
		}
		route := ts.Route
		if len(route) == 0 {
			route = []string{ts.FromID, ts.ToID}
		}
		currentIdx := max(ts.RouteIndex, 0)
		if len(route) > 0 && currentIdx >= len(route) {
			currentIdx = len(route) - 1
		}
		steps := make([]travelStepView, 0, len(route))
		for i, id := range route {
			stepName := id
			if loc := sim.World.Location(id); loc != nil {
				stepName = loc.Name
			}
			steps = append(steps, travelStepView{
				ID:      id,
				Name:    stepName,
				Current: i == currentIdx,
			})
		}
		progress := 0
		if len(route) > 1 {
			progress = currentIdx * 100 / (len(route) - 1)
		}
		travelers = append(travelers, travelerView{
			EntityID:    ts.EntityID,
			Name:        name,
			From:        ts.FromID,
			To:          ts.ToID,
			Progress:    progress,
			Eta:         ts.TotalTicks - ts.ElapsedTicks,
			Route:       steps,
			CurrentStep: currentIdx,
			TotalSteps:  len(route) - 1,
		})
	}
	sortTravelersForDisplay(travelers)
	return travelers
}

func (h *Handler) EntityDetailPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	entID := c.Param("id")
	ent := h.Sim.Entities.Get(entID)
	if ent == nil {
		c.String(404, "Entity not found")
		return
	}
	all := h.Sim.Entities.All()
	loc := h.Sim.World.Location(ent.LocationID)
	locName := ent.LocationID
	if loc != nil {
		locName = loc.Name
	}

	events := combat.LocationEvents(ent.LocationID, 200)
	var entEvents []combat.Event
	for _, e := range events {
		if e.AttackerID == ent.ID || e.DefenderID == ent.ID {
			entEvents = append(entEvents, e)
		}
	}
	if len(entEvents) > 50 {
		entEvents = entEvents[len(entEvents)-50:]
	}

	flagsJSON := "{}"
	if len(ent.Flags) > 0 {
		b, err := json.Marshal(ent.Flags)
		if err == nil {
			flagsJSON = string(b)
		}
	}
	inventory := make([]items.ItemInstance, len(ent.Inventory))
	copy(inventory, ent.Inventory)
	sortInventoryForDisplay(inventory)
	effects := make([]entity.ActiveEffect, len(ent.Effects))
	copy(effects, ent.Effects)
	sortEffectsForDisplay(effects)

	eff := ent.EffectiveAttrs()

	xpForNext := ent.Level * 100
	xpPercent := 0
	canLevelUp := false
	if xpForNext > 0 && entity.CanLevelUp(ent.Species) {
		xpPercent = ent.XP * 100 / xpForNext
		if ent.XP >= xpForNext {
			canLevelUp = true
		}
	}

	if err := h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":         "Entity: " + ent.Name,
		"page":          "entity_detail",
		"tick":          h.Sim.Tick,
		"time":          h.Sim.Time.String(),
		"phase":         h.Sim.Time.Phase().String(),
		"season":        h.Sim.Time.Season().String(),
		"entities":      len(all),
		"locations":     len(h.Sim.World.AllLocations()),
		"entity":        ent,
		"loc_name":      locName,
		"equip_slots":   getEquipSlots(ent),
		"inventory":     inventory,
		"effects":       effects,
		"travel_route":  buildTravelRouteView(h.Sim, ent),
		"combat_events": entEvents,
		"flags_json":    flagsJSON,
		"effective_str": eff.STR,
		"effective_dex": eff.DEX,
		"effective_con": eff.CON,
		"effective_int": eff.INT,
		"effective_wis": eff.WIS,
		"effective_cha": eff.CHA,
		"xp_for_next":   xpForNext,
		"xp_percent":    xpPercent,
		"can_level_up":  canLevelUp,
		"mood_mods_str": moodModsString(ent),
		"skills":        buildSkillInfo(ent),
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) EntityDetailFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	entID := c.Param("id")
	ent := h.Sim.Entities.Get(entID)
	if ent == nil {
		c.String(404, "Entity not found")
		return
	}
	loc := h.Sim.World.Location(ent.LocationID)
	locName := ent.LocationID
	if loc != nil {
		locName = loc.Name
	}

	events := combat.LocationEvents(ent.LocationID, 200)
	var entEvents []combat.Event
	for _, e := range events {
		if e.AttackerID == ent.ID || e.DefenderID == ent.ID {
			entEvents = append(entEvents, e)
		}
	}
	if len(entEvents) > 50 {
		entEvents = entEvents[len(entEvents)-50:]
	}

	flagsJSON := "{}"
	if len(ent.Flags) > 0 {
		b, err := json.Marshal(ent.Flags)
		if err == nil {
			flagsJSON = string(b)
		}
	}
	inventory := make([]items.ItemInstance, len(ent.Inventory))
	copy(inventory, ent.Inventory)
	sortInventoryForDisplay(inventory)
	effects := make([]entity.ActiveEffect, len(ent.Effects))
	copy(effects, ent.Effects)
	sortEffectsForDisplay(effects)

	eff := ent.EffectiveAttrs()

	xpForNext := ent.Level * 100
	xpPercent := 0
	canLevelUp := false
	if xpForNext > 0 && entity.CanLevelUp(ent.Species) {
		xpPercent = ent.XP * 100 / xpForNext
		if ent.XP >= xpForNext {
			canLevelUp = true
		}
	}

	if err := h.Tmpls.ExecuteTemplate(c.Writer, "entity_detail_status", gin.H{
		"entity":        ent,
		"loc_name":      locName,
		"equip_slots":   getEquipSlots(ent),
		"inventory":     inventory,
		"effects":       effects,
		"travel_route":  buildTravelRouteView(h.Sim, ent),
		"combat_events": entEvents,
		"flags_json":    flagsJSON,
		"effective_str": eff.STR,
		"effective_dex": eff.DEX,
		"effective_con": eff.CON,
		"effective_int": eff.INT,
		"effective_wis": eff.WIS,
		"effective_cha": eff.CHA,
		"xp_for_next":   xpForNext,
		"xp_percent":    xpPercent,
		"can_level_up":  canLevelUp,
		"mood_mods_str": moodModsString(ent),
		"quest_states":  h.Sim.Quests.EntityStates(ent.ID),
		"skills":        buildSkillInfo(ent),
	}); err != nil {
		_ = c.Error(err)
	}
}

func moodModsString(e *entity.Entity) string {
	if len(e.MoodModifiers) == 0 {
		return ""
	}
	var parts []string
	for _, m := range e.MoodModifiers {
		remaining := int(m.DecayAtTick) - e.Age
		if remaining < 0 {
			remaining = 0
		}
		parts = append(parts, fmt.Sprintf("%s:%dt", m.Mood, remaining))
	}
	return joinStr(parts, ", ")
}

func joinStr(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += sep + p
	}
	return out
}

func getEquipSlots(e *entity.Entity) []equipSlot {
	return []equipSlot{
		{"Head", e.Equipment.Head},
		{"Body", e.Equipment.Body},
		{"Weapon", e.Equipment.Weapon},
		{"Offhand", e.Equipment.Offhand},
		{"Feet", e.Equipment.Feet},
		{"Hands", e.Equipment.Hands},
		{"Neck", e.Equipment.Neck},
		{"Finger 1", e.Equipment.Finger1},
		{"Finger 2", e.Equipment.Finger2},
	}
}

type mapNode struct {
	loc         *world.Location
	depth       int
	count       int
	travelCount int
	routeNotes  []string
	x, y        float64
	w, h        float64
	children    []mapNode
}

func buildMapLayout(nodes []locNode, depth int) []mapNode {
	var out []mapNode
	for _, n := range nodes {
		mn := mapNode{
			loc:   n.Location,
			depth: depth,
			count: n.EntityCount,
		}
		mn.children = buildMapLayout(n.Children, depth+1)
		out = append(out, mn)
	}
	return out
}

func annotateMapTravelers(nodes []mapNode, travelers []travelerView) {
	if len(nodes) == 0 || len(travelers) == 0 {
		return
	}
	notesByLoc := make(map[string][]string)
	countsByLoc := make(map[string]int)
	for _, tv := range travelers {
		if len(tv.Route) == 0 {
			continue
		}
		currentIdx := tv.CurrentStep
		if currentIdx < 0 {
			currentIdx = 0
		}
		if currentIdx >= len(tv.Route) {
			currentIdx = len(tv.Route) - 1
		}
		routeIDs := make([]string, 0, len(tv.Route))
		for i, step := range tv.Route {
			label := step.Name
			if label == "" {
				label = step.ID
			}
			if i == currentIdx {
				label = "[" + label + "]"
			}
			routeIDs = append(routeIDs, label)
		}
		routeText := strings.Join(routeIDs, " -> ")
		for _, step := range tv.Route {
			if step.ID == "" {
				continue
			}
			countsByLoc[step.ID]++
			notesByLoc[step.ID] = append(notesByLoc[step.ID], fmt.Sprintf("%s: %s (%d/%d, %dt)",
				tv.Name,
				routeText,
				currentIdx+1,
				len(tv.Route),
				tv.Eta,
			))
		}
	}
	var walk func([]mapNode)
	walk = func(ns []mapNode) {
		for i := range ns {
			if ns[i].loc != nil {
				ns[i].travelCount = countsByLoc[ns[i].loc.ID]
				ns[i].routeNotes = notesByLoc[ns[i].loc.ID]
			}
			if len(ns[i].children) > 0 {
				walk(ns[i].children)
			}
		}
	}
	walk(nodes)
}

func assignMapPositions(nodes []mapNode, startX, levelW float64, levelY float64, vSpacing float64) (float64, float64) {
	if len(nodes) == 0 {
		return startX, levelY
	}
	leafCount := 0
	for i := range nodes {
		if len(nodes[i].children) == 0 {
			leafCount++
		} else {
			leafCount += countLeaves(nodes[i].children)
		}
	}
	totalW := float64(leafCount)*levelW + float64(leafCount-1)*8
	cx := startX + totalW/2

	x := startX
	for i := range nodes {
		if len(nodes[i].children) == 0 {
			nodes[i].x = x
			nodes[i].y = levelY
			nodes[i].w = levelW
			nodes[i].h = 36
			x += levelW + 8
		} else {
			childX, _ := assignMapPositions(nodes[i].children, x, levelW, levelY+vSpacing, vSpacing)
			childW := childX - x
			if childW < levelW {
				childW = levelW
			}
			nodes[i].x = x + (childW-levelW)/2
			nodes[i].y = levelY
			nodes[i].w = levelW
			nodes[i].h = 36
			x = childX
		}
	}

	return cx, levelY
}

func countLeaves(nodes []mapNode) int {
	c := 0
	for _, n := range nodes {
		if len(n.children) == 0 {
			c++
		} else {
			c += countLeaves(n.children)
		}
	}
	return c
}

func styleForLocType(locType world.LocationType) (fill, stroke, text string) {
	switch locType {
	case world.LocWorld:
		return "#2d2d5e", "#ffd700", "#ffd700"
	case world.LocRegion:
		return "#1a3a5c", "#4a9eff", "#c8e6ff"
	case world.LocCity:
		return "#1a3a2e", "#4caf50", "#a5d6a7"
	case world.LocBuilding:
		return "#2a2a3e", "#888", "#ccc"
	case world.LocRoom:
		return "#222233", "#666", "#999"
	case world.LocRealm:
		return "#2a1a3e", "#b388ff", "#d1c4e9"
	default:
		return "#1a1a2e", "#555", "#888"
	}
}

func renderNodeSVG(n mapNode, entCount int) string {
	fill, stroke, textClr := styleForLocType(n.loc.Type)
	w := n.w
	if w < 80 {
		w = 80
	}
	h := n.h
	name := n.loc.Name
	if len(name) > 16 {
		name = name[:14] + "…"
	}
	link := fmt.Sprintf(`/combat/%s`, n.loc.ID)
	title := fmt.Sprintf("%s (%s)\n%d entities", n.loc.Name, n.loc.ID, entCount)
	if n.travelCount > 0 {
		title += fmt.Sprintf("\n%d traveler(s) in transit", n.travelCount)
		for _, note := range n.routeNotes {
			title += "\n" + note
		}
	}

	var svg string
	svg += fmt.Sprintf(`<a xlink:href="%s" target="_top">`, link)
	svg += fmt.Sprintf(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="6" ry="6" fill="%s" stroke="%s" stroke-width="1.5"/>`,
		n.x-w/2, n.y, w, h, fill, stroke)
	svg += fmt.Sprintf(`<text x="%.0f" y="%.0f" text-anchor="middle" fill="%s" font-size="11" font-family="monospace">%s</text>`,
		n.x, n.y+h/2+4, textClr, name)
	if n.travelCount > 0 {
		badge := fmt.Sprintf(`<g><rect x="%.0f" y="%.0f" width="18" height="14" rx="4" ry="4" fill="#e94560" opacity="0.95"/><text x="%.0f" y="%.0f" text-anchor="middle" fill="#fff" font-size="10" font-family="monospace">%d</text></g>`,
			n.x+w/2-20, n.y+4, n.x+w/2-11, n.y+14, n.travelCount)
		svg += badge
	}
	svg += fmt.Sprintf(`<title>%s</title>`, html.EscapeString(title))
	svg += `</a>`

	for _, c := range n.children {
		cx := n.x
		cy := n.y + h
		px := c.x
		py := c.y
		svg += fmt.Sprintf(`<path d="M%.0f,%.0f C%.0f,%.0f %.0f,%.0f %.0f,%.0f" fill="none" stroke="#444" stroke-width="1.5"/>`,
			cx, cy, cx, cy+(py-cy)/2, px, cy+(py-cy)/2, px, py)
		svg += renderNodeSVG(c, c.count)
	}

	return svg
}

func renderLocationMap(nodes []locNode, travelers []travelerView) template.HTML {
	layout := buildMapLayout(nodes, 0)
	if len(layout) == 0 {
		return template.HTML("")
	}
	annotateMapTravelers(layout, travelers)

	levelW := 110.0
	vSpacing := 80.0
	padX := 40.0
	padY := 30.0

	assignMapPositions(layout, padX, levelW, padY, vSpacing)

	width := 0.0
	height := 0.0
	var walk func([]mapNode)
	walk = func(ns []mapNode) {
		for _, n := range ns {
			r := n.x + n.w/2 + padX
			if r > width {
				width = r
			}
			b := n.y + n.h + vSpacing
			if b > height {
				height = b
			}
			walk(n.children)
		}
	}
	walk(layout)
	if width < 600 {
		width = 600
	}

	svg := fmt.Sprintf(`<svg viewBox="0 0 %.0f %.0f" xmlns="http://www.w3.org/2000/svg" style="width:100%%;max-width:%dpx;background:#111;border-radius:8px;font-family:monospace">`,
		width, height, int(width)+40)

	var legend string
	types := []struct {
		label string
		color string
		loc   world.LocationType
	}{
		{"World", "#ffd700", world.LocWorld},
		{"Region", "#4a9eff", world.LocRegion},
		{"City", "#4caf50", world.LocCity},
		{"Building", "#888", world.LocBuilding},
		{"Room", "#666", world.LocRoom},
		{"Realm", "#b388ff", world.LocRealm},
	}
	lx := width - 160
	ly := padY + 10
	for _, t := range types {
		legend += fmt.Sprintf(`<rect x="%.0f" y="%.0f" width="12" height="12" rx="2" fill="%s"/>`, lx, ly, t.color)
		legend += fmt.Sprintf(`<text x="%.0f" y="%.0f" fill="#aaa" font-size="10">%s</text>`, lx+18, ly+10, t.label)
		ly += 18
	}

	svg += legend

	for _, n := range layout {
		svg += renderNodeSVG(n, n.count)
	}

	svg += `</svg>`
	return template.HTML(svg)
}

type skillInfo struct {
	Name  string
	Level int
	XP    int
}

func buildSkillInfo(e *entity.Entity) []skillInfo {
	var out []skillInfo
	for _, name := range entity.SkillNames {
		level := e.SkillLevel(name)
		xp := level * entity.SkillXPPerLevel
		out = append(out, skillInfo{Name: name, Level: level, XP: xp})
	}
	return out
}

func (h *Handler) SSEEvents(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ch := make(chan uint64, 64)
	h.Sim.OnEvent(func(evt events.SimEvent) {
		select {
		case ch <- evt.Tick:
		default:
		}
	})

	c.Stream(func(w io.Writer) bool {
		tick, ok := <-ch
		if !ok {
			return false
		}
		c.SSEvent("tick", fmt.Sprintf(`{"tick":%d}`, tick))
		return true
	})
}

// --- Pregnancies view ---

type pregnantEntityView struct {
	EntityID       string
	EntityName     string
	Species        string
	Faction        string
	Level          int
	Progress       int
	TicksRemaining int
	LocationID     string
	LocationName   string
}

type recentBirthView struct {
	OffspringID   string
	OffspringName string
	Species       string
	Gender        string
	ParentID      string
	ParentName    string
	Tick          uint64
}

type relationshipView struct {
	EntityAID   string
	EntityAName string
	EntityBID   string
	EntityBName string
	Type        string
	SinceTick   uint64
}

// SpeciesGestationTicks returns the gestation period in ticks for a given species.
func SpeciesGestationTicks(species string) int {
	switch species {
	case "human":
		return 200
	case "orc":
		return 150
	case "elf":
		return 250
	case "dwarf":
		return 180
	case "goblin":
		return 120
	case "fey":
		return 160
	case "rat_king":
		return 100
	case "kobold":
		return 90
	case "vampire":
		return 0
	case "hag":
		return 140
	default:
		return 200
	}
}

func (h *Handler) PregnanciesPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	pregnantEntities := buildPregnantEntities(h.Sim)
	recentBirths := buildRecentBirths(h.Sim)
	relationships := buildRelationships(h.Sim)

	if err := h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":             "Pregnancies & Births",
		"page":              "pregnancies",
		"tick":              h.Sim.Tick,
		"time":              h.Sim.Time.String(),
		"phase":             h.Sim.Time.Phase().String(),
		"season":            h.Sim.Time.Season().String(),
		"entities":          len(h.Sim.Entities.All()),
		"locations":         len(h.Sim.World.AllLocations()),
		"pregnant_entities": pregnantEntities,
		"recent_births":     recentBirths,
		"relationships":     relationships,
	}); err != nil {
		_ = c.Error(err)
	}
}

func (h *Handler) PregnanciesFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()

	pregnantEntities := buildPregnantEntities(h.Sim)
	recentBirths := buildRecentBirths(h.Sim)
	relationships := buildRelationships(h.Sim)

	if err := h.Tmpls.ExecuteTemplate(c.Writer, "pregnancies_list", gin.H{
		"tick":              h.Sim.Tick,
		"pregnant_entities": pregnantEntities,
		"recent_births":     recentBirths,
		"relationships":     relationships,
	}); err != nil {
		_ = c.Error(err)
	}
}

func buildPregnantEntities(sim *engine.Simulation) []pregnantEntityView {
	var out []pregnantEntityView
	for _, e := range sim.Entities.All() {
		if !e.Pregnant {
			continue
		}
		gestation := SpeciesGestationTicks(e.Species)
		if gestation <= 0 {
			gestation = 200
		}
		// PregnancyTick not yet available on entity.Entity; progress unavailable.
		progress := 0
		remaining := gestation

		locName := e.LocationID
		if loc := sim.World.Location(e.LocationID); loc != nil {
			locName = loc.Name
		}

		out = append(out, pregnantEntityView{
			EntityID:       e.ID,
			EntityName:     e.Name,
			Species:        e.Species,
			Faction:        e.Faction,
			Level:          e.Level,
			Progress:       progress,
			TicksRemaining: remaining,
			LocationID:     e.LocationID,
			LocationName:   locName,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].EntityName != out[j].EntityName {
			return out[i].EntityName < out[j].EntityName
		}
		return out[i].EntityID < out[j].EntityID
	})
	return out
}

func buildRecentBirths(sim *engine.Simulation) []recentBirthView {
	var out []recentBirthView
	for _, evt := range sim.EventsCopy() {
		if evt.Type != events.EventEntityBorn {
			continue
		}
		data := evt.Data
		var parentID string
		var offspringID string
		if mother, ok := data["mother"].(string); ok {
			parentID = mother
		} else if father, ok := data["father"].(string); ok {
			parentID = father
		}
		if sourceID, ok := data["child"].(string); ok {
			offspringID = sourceID
		} else {
			offspringID = evt.Source
		}
		parent := sim.Entities.Get(parentID)
		parentName := parentID
		if parent != nil {
			parentName = parent.Name
		}
		offspring := sim.Entities.Get(offspringID)
		offspringName := offspringID
		species := ""
		gender := ""
		if offspring != nil {
			offspringName = offspring.Name
			species = offspring.Species
			gender = offspring.Gender
		}
		out = append(out, recentBirthView{
			OffspringID:   offspringID,
			OffspringName: offspringName,
			Species:       species,
			Gender:        gender,
			ParentID:      parentID,
			ParentName:    parentName,
			Tick:          evt.Tick,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Tick > out[j].Tick
	})
	if len(out) > 50 {
		out = out[:50]
	}
	return out
}

func buildRelationships(sim *engine.Simulation) []relationshipView {
	var out []relationshipView
	seen := make(map[string]bool)
	for _, e := range sim.Entities.All() {
		for _, rel := range e.Relationships {
			key := rel.OtherID + ":" + e.ID
			reverseKey := e.ID + ":" + rel.OtherID
			if seen[key] || seen[reverseKey] {
				continue
			}
			seen[key] = true
			seen[reverseKey] = true

			other := sim.Entities.Get(rel.OtherID)
			otherName := rel.OtherID
			if other != nil {
				otherName = other.Name
			}
			out = append(out, relationshipView{
				EntityAID:   e.ID,
				EntityAName: e.Name,
				EntityBID:   rel.OtherID,
				EntityBName: otherName,
				Type:        string(rel.Type),
				SinceTick:   rel.SinceTick,
			})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].EntityAName != out[j].EntityAName {
			return out[i].EntityAName < out[j].EntityAName
		}
		if out[i].EntityBName != out[j].EntityBName {
			return out[i].EntityBName < out[j].EntityBName
		}
		return out[i].SinceTick < out[j].SinceTick
	})
	return out
}
