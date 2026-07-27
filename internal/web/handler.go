package web

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"io/fs"

	"simuz/internal/combat"
	"simuz/internal/engine"
	"simuz/internal/entity"
	"simuz/internal/items"
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

	h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":     "Simuz",
		"page":      "dashboard",
		"tick":      h.Sim.Tick,
		"time":      h.Sim.Time.String(),
		"phase":     h.Sim.Time.Phase().String(),
		"season":    h.Sim.Time.Season().String(),
		"entities":  len(h.Sim.Entities.All()),
		"locations": len(h.Sim.World.AllLocations()),
	})
}

func (h *Handler) EntitiesPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	all := h.Sim.Entities.All()
	h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":         "Entities",
		"page":          "entities",
		"tick":          h.Sim.Tick,
		"time":          h.Sim.Time.String(),
		"phase":         h.Sim.Time.Phase().String(),
		"season":        h.Sim.Time.Season().String(),
		"entities":      len(all),
		"entities_list": all,
		"locations":     len(h.Sim.World.AllLocations()),
	})
}

type locNode struct {
	Location *world.Location
	Children []locNode
	Depth    int
	EntityCount int
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
	var nodes []locNode
	for _, loc := range children {
		ec := len(entities.ByLocation(loc.ID))
		node := locNode{
			Location: loc,
			Depth:    depth,
			EntityCount: ec,
			Children: buildChildren(w, entities, loc.ID, depth+1),
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

	h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":      "Locations",
		"page":       "locations",
		"tick":       h.Sim.Tick,
		"time":       h.Sim.Time.String(),
		"phase":      h.Sim.Time.Phase().String(),
		"season":     h.Sim.Time.Season().String(),
		"entities":   len(all),
		"locations":  len(h.Sim.World.AllLocations()),
		"loc_tree":   buildLocationTree(h.Sim.World, h.Sim.Entities),
		"root_name":  h.Sim.World.RootLocation().Name,
		"realms":     realms,
	})
}

func (h *Handler) CombatPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	all := h.Sim.Entities.All()

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

	var zones []combatZone
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
		zones = append(zones, combatZone{LocationName: name, LocationID: locID, Factions: fnames, EntityCount: total})
	}

	h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":       "Combat",
		"page":        "combat",
		"tick":        h.Sim.Tick,
		"time":        h.Sim.Time.String(),
		"phase":       h.Sim.Time.Phase().String(),
		"season":      h.Sim.Time.Season().String(),
		"entities":    len(all),
		"locations":   len(h.Sim.World.AllLocations()),
		"combat_zones": zones,
		"combat_log":  combat.RecentLog(50),
	})
}

func (h *Handler) QuestsPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	defs := h.Sim.Quests.AllDefs()
	h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":      "Quests",
		"page":       "quests",
		"tick":       h.Sim.Tick,
		"time":       h.Sim.Time.String(),
		"phase":      h.Sim.Time.Phase().String(),
		"season":     h.Sim.Time.Season().String(),
		"entities":   len(h.Sim.Entities.All()),
		"locations":  len(h.Sim.World.AllLocations()),
		"quest_defs": defs,
	})
}

func (h *Handler) AIPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":     "AI",
		"page":      "ai",
		"tick":      h.Sim.Tick,
		"time":      h.Sim.Time.String(),
		"phase":     h.Sim.Time.Phase().String(),
		"season":    h.Sim.Time.Season().String(),
		"entities":  len(h.Sim.Entities.All()),
		"locations": len(h.Sim.World.AllLocations()),
	})
}

func (h *Handler) EventsPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	evList := h.Sim.Events.RecentEvents(100)
	h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":     "Events",
		"page":      "events",
		"tick":      h.Sim.Tick,
		"time":      h.Sim.Time.String(),
		"phase":     h.Sim.Time.Phase().String(),
		"season":    h.Sim.Time.Season().String(),
		"entities":  len(h.Sim.Entities.All()),
		"locations": len(h.Sim.World.AllLocations()),
		"events":    evList,
	})
}

func (h *Handler) EventsFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	evList := h.Sim.Events.RecentEvents(50)
	h.Tmpls.ExecuteTemplate(c.Writer, "events_list", gin.H{
		"events": evList,
		"tick":   h.Sim.Tick,
	})
}

func (h *Handler) DashboardFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	h.Tmpls.ExecuteTemplate(c.Writer, "dashboard_stats", gin.H{
		"tick":      h.Sim.Tick,
		"time":      h.Sim.Time.String(),
		"phase":     h.Sim.Time.Phase().String(),
		"season":    h.Sim.Time.Season().String(),
		"entities":  len(h.Sim.Entities.All()),
		"locations": len(h.Sim.World.AllLocations()),
	})
}

func (h *Handler) EntitiesFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	all := h.Sim.Entities.All()
	h.Tmpls.ExecuteTemplate(c.Writer, "entities_table", gin.H{
		"tick":          h.Sim.Tick,
		"time":          h.Sim.Time.String(),
		"entities":      len(all),
		"entities_list": all,
		"locations":     len(h.Sim.World.AllLocations()),
	})
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
	h.Tmpls.ExecuteTemplate(c.Writer, "locations_tree", gin.H{
		"tick":      h.Sim.Tick,
		"time":      h.Sim.Time.String(),
		"entities":  len(all),
		"locations": len(h.Sim.World.AllLocations()),
		"loc_tree":  buildLocationTree(h.Sim.World, h.Sim.Entities),
		"root_name": h.Sim.World.RootLocation().Name,
		"realms":    realms,
	})
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
	events := combat.LocationEvents(locID, 30)
	var travelers []gin.H
	if h.Sim.Traveling != nil {
		for _, ts := range h.Sim.Traveling {
			if ts != nil && (ts.FromID == locID || ts.ToID == locID) && ts.Status == world.TravelInProgress {
				name := ts.EntityID
				if e := h.Sim.Entities.Get(ts.EntityID); e != nil {
					name = e.Name
				}
				travelers = append(travelers, gin.H{
					"entity_id": ts.EntityID,
					"name":      name,
					"from":      ts.FromID,
					"to":        ts.ToID,
					"progress":  int(ts.Progress() * 100),
					"eta":       ts.TotalTicks - ts.ElapsedTicks,
				})
			}
		}
	}
	return gin.H{
		"title":     "Location: " + loc.Name,
		"page":      "location_detail",
		"tick":      h.Sim.Tick,
		"time":      h.Sim.Time.String(),
		"phase":     h.Sim.Time.Phase().String(),
		"season":    h.Sim.Time.Season().String(),
		"entities":  len(h.Sim.Entities.All()),
		"locations": len(h.Sim.World.AllLocations()),
		"loc":       loc,
		"parent_name": parentName,
		"entity_rows": rows,
		"weather":   wth,
		"children":  children,
		"combat_events": events,
		"travelers": travelers,
	}, true
}

func (h *Handler) LocationDetailPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	data, ok := h.locationDetailData(c.Param("id"))
	if !ok {
		c.String(404, "Location not found")
		return
	}
	h.Tmpls.ExecuteTemplate(c.Writer, "base.html", data)
}

func (h *Handler) LocationDetailFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	data, ok := h.locationDetailData(c.Param("id"))
	if !ok {
		c.String(404, "Location not found")
		return
	}
	h.Tmpls.ExecuteTemplate(c.Writer, "location_detail_status", data)
}

type combatZone struct {
	LocationName string
	LocationID   string
	Factions     []string
	EntityCount  int
}

func (h *Handler) CombatFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	all := h.Sim.Entities.All()

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

	var zones []combatZone
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
		zones = append(zones, combatZone{LocationName: name, LocationID: locID, Factions: fnames, EntityCount: total})
	}

	h.Tmpls.ExecuteTemplate(c.Writer, "combat_status", gin.H{
		"tick":        h.Sim.Tick,
		"time":        h.Sim.Time.String(),
		"entities":    len(all),
		"locations":   len(h.Sim.World.AllLocations()),
		"combat_zones": zones,
		"combat_log":  combat.RecentLog(50),
	})
}

func (h *Handler) QuestsFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	defs := h.Sim.Quests.AllDefs()
	h.Tmpls.ExecuteTemplate(c.Writer, "quests_list", gin.H{
		"tick":       h.Sim.Tick,
		"time":       h.Sim.Time.String(),
		"entities":   len(h.Sim.Entities.All()),
		"locations":  len(h.Sim.World.AllLocations()),
		"quest_defs": defs,
	})
}

func (h *Handler) AIFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	all := h.Sim.Entities.All()
	var activeEntities []*entity.Entity
	for _, e := range all {
		if e.AI.Type == "scripted" || e.AI.Type == "aggressive" || e.AI.Type == "dormant" {
			activeEntities = append(activeEntities, e)
		}
	}
	h.Tmpls.ExecuteTemplate(c.Writer, "ai_status", gin.H{
		"tick":          h.Sim.Tick,
		"time":          h.Sim.Time.String(),
		"entities":      len(all),
		"locations":     len(h.Sim.World.AllLocations()),
		"entities_list": activeEntities,
	})
}

func (h *Handler) CombatDetailPage(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	locID := c.Param("location")
	loc := h.Sim.World.Location(locID)
	name := locID
	if loc != nil {
		name = loc.Name
	}

	all := h.Sim.Entities.All()
	var combatants []*entity.Entity
	for _, e := range all {
		if e.LocationID == locID && e.Alive {
			combatants = append(combatants, e)
		}
	}

	events := combat.LocationEvents(locID, 100)

	h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":       "Combat: " + name,
		"page":        "combat_detail",
		"tick":        h.Sim.Tick,
		"time":        h.Sim.Time.String(),
		"phase":       h.Sim.Time.Phase().String(),
		"season":      h.Sim.Time.Season().String(),
		"entities":    len(all),
		"locations":   len(h.Sim.World.AllLocations()),
		"loc_id":      locID,
		"loc_name":    name,
		"combatants":  combatants,
		"combat_events": events,
	})
}

func (h *Handler) CombatDetailFragment(c *gin.Context) {
	h.Sim.RLock()
	defer h.Sim.RUnlock()
	locID := c.Param("location")
	loc := h.Sim.World.Location(locID)
	name := locID
	if loc != nil {
		name = loc.Name
	}

	all := h.Sim.Entities.All()
	var combatants []*entity.Entity
	for _, e := range all {
		if e.LocationID == locID && e.Alive {
			combatants = append(combatants, e)
		}
	}

	events := combat.LocationEvents(locID, 100)

	h.Tmpls.ExecuteTemplate(c.Writer, "combat_detail_status", gin.H{
		"tick":         h.Sim.Tick,
		"time":         h.Sim.Time.String(),
		"loc_id":       locID,
		"loc_name":     name,
		"combatants":   combatants,
		"combat_events": events,
	})
}

type equipSlot struct {
	Slot string
	Item *items.ItemInstance
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

	h.Tmpls.ExecuteTemplate(c.Writer, "base.html", gin.H{
		"title":        "Entity: " + ent.Name,
		"page":         "entity_detail",
		"tick":         h.Sim.Tick,
		"time":         h.Sim.Time.String(),
		"phase":        h.Sim.Time.Phase().String(),
		"season":       h.Sim.Time.Season().String(),
		"entities":     len(all),
		"locations":    len(h.Sim.World.AllLocations()),
		"entity":       ent,
		"loc_name":     locName,
		"equip_slots":  getEquipSlots(ent),
		"combat_events": entEvents,
		"flags_json":   flagsJSON,
		"effective_str":  eff.STR,
		"effective_dex":  eff.DEX,
		"effective_con":  eff.CON,
		"effective_int":  eff.INT,
		"effective_wis":  eff.WIS,
		"effective_cha":  eff.CHA,
		"xp_for_next":   xpForNext,
		"xp_percent":    xpPercent,
		"can_level_up":  canLevelUp,
		"mood_mods_str": moodModsString(ent),
		"skills":        buildSkillInfo(ent),
	})
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

	h.Tmpls.ExecuteTemplate(c.Writer, "entity_detail_status", gin.H{
		"entity":       ent,
		"loc_name":     locName,
		"equip_slots":  getEquipSlots(ent),
		"combat_events": entEvents,
		"flags_json":   flagsJSON,
		"effective_str":  eff.STR,
		"effective_dex":  eff.DEX,
		"effective_con":  eff.CON,
		"effective_int":  eff.INT,
		"effective_wis":  eff.WIS,
		"effective_cha":  eff.CHA,
		"xp_for_next":   xpForNext,
		"xp_percent":    xpPercent,
		"can_level_up":  canLevelUp,
		"mood_mods_str": moodModsString(ent),
		"quest_states":  h.Sim.Quests.EntityStates(ent.ID),
		"skills":        buildSkillInfo(ent),
	})
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
	loc      *world.Location
	depth    int
	count    int
	x, y     float64
	w, h     float64
	children []mapNode
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
	title := fmt.Sprintf(`%s (%s)\n%d entities`, n.loc.Name, n.loc.ID, entCount)

	var svg string
	svg += fmt.Sprintf(`<a xlink:href="%s" target="_top">`, link)
	svg += fmt.Sprintf(`<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="6" ry="6" fill="%s" stroke="%s" stroke-width="1.5"/>`,
		n.x-w/2, n.y, w, h, fill, stroke)
	svg += fmt.Sprintf(`<text x="%.0f" y="%.0f" text-anchor="middle" fill="%s" font-size="11" font-family="monospace">%s</text>`,
		n.x, n.y+h/2+4, textClr, name)
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

func renderLocationMap(nodes []locNode) template.HTML {
	layout := buildMapLayout(nodes, 0)
	if len(layout) == 0 {
		return template.HTML("")
	}

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
	h.Sim.OnEvent(func(evt engine.SimEvent) {
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
