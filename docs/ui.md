# Web UI (htmx)

## Approach

Server-rendered HTML with [htmx](https://htmx.org/) for interactivity. Gin serves Go `html/template` pages. No SPA framework, no npm build step. Static assets (CSS, JS, htmx) are embedded via `embed.FS`.

Real-time updates use Server-Sent Events (SSE) pushed from the tick loop.

---

## Architecture

```
┌──────────┐   HTTP/HTML    ┌──────────────┐
│ Browser  │ ←────────────→ │    Gin       │
│ (htmx)   │                │  templates/  │
│          │ ←── SSE ───── │  handlers/   │
└──────────┘                └──────┬───────┘
                                   │
                                   ▼
                           ┌───────────────┐
                           │   Engine      │
                           │  (Tick Loop)  │
                           └───────────────┘
```

- All pages are server-rendered Go templates
- htmx attributes replace full-page navigation with fragment swaps
- SSE pushes tick updates to the browser (world time, entity positions, combat events)
- No client-side state — the server is the source of truth

---

## Template Structure

```
internal/
└── web/
    ├── handler.go           # HTTP handlers for UI routes
    ├── sse.go               # SSE hub (broadcasts tick events)
    ├── templates/
    │   ├── base.html        # Layout shell (nav, sidebar, content slot)
    │   ├── dashboard.html   # World overview page
    │   ├── locations.html   # Location tree browser
    │   ├── location.html    # Single location detail
    │   ├── entities.html    # Entity list with filters
    │   ├── entity.html      # Single entity sheet
    │   ├── combat.html      # Active combat instances
    │   ├── quests.html      # Quest journal
    │   ├── ai.html          # AI debug console
    │   ├── fragments/       # HTMX fragments (partial templates)
    │   │   ├── world_clock.html
    │   │   ├── entity_list.html
    │   │   ├── combat_log.html
    │   │   ├── location_children.html
    │   │   ├── entity_hp_bar.html
    │   │   └── notification.html
    │   └── static/          # Embedded static assets
    │       ├── css/
    │       │   └── app.css
    │       └── js/
    │           └── app.js   # Tiny: htmx config, SSE reconnect
    ├── static.go            # embed.FS for static assets
    └── routes.go            # Route registration on Gin
```

### base.html Layout

```html
<!DOCTYPE html>
<html>
<head>
    <title>Simuz</title>
    <script src="/static/js/htmx.min.js"></script>
    <script src="/static/js/app.js"></script>
    <link rel="stylesheet" href="/static/css/app.css">
</head>
<body hx-ext="sse" sse-connect="/api/v1/ui/events">
    <nav>
        <a href="/">Dashboard</a>
        <a href="/locations">World</a>
        <a href="/entities">Entities</a>
        <a href="/combat">Combat</a>
        <a href="/quests">Quests</a>
        <a href="/ai">AI</a>
    </nav>
    <main>
        {{block "content" .}}{{end}}
    </main>
    <aside id="notifications"
           sse-swap="notification"
           hx-swap="beforeend">
    </aside>
</body>
</html>
```

---

## Page Descriptions

### Dashboard (`/`)

World summary — time, weather, entity counts, recent events.

```
┌─────────────────────────────────────────┐
│  Simuz                           [Tick] │
├──────────┬──────────────────────────────┤
│  Clock   │  Weather                     │
│  Day 14  │  ☁ Overcast, 12°C           │
│  14:32   │  Wind: 5km/h NE             │
├──────────┼──────────────────────────────┤
│  Stats   │  Recent Events               │
│  │ 126 entities │  ▶ Greta served ale   │
│  │ 47 locations │  ▶ Guard patrols gate │
│  │ 3 combats    │  ▶ Wolf killed boar   │
│  │ 12 active    │  ▶ Player entered     │
│  │   quests     │    Frosthold          │
└──────────┴──────────────────────────────┘
```

### World Browser (`/locations`)

Recursive tree of locations. Click to expand children via htmx.

```html
<ul>
    {{range .Locations}}
    <li>
        <span hx-get="/locations/{{.ID}}/children"
              hx-target="next ul"
              hx-trigger="click"
              hx-swap="innerHTML">
            {{.Icon}} {{.Name}}
        </span>
        <ul></ul>
    </li>
    {{end}}
</ul>
```

Clicking a leaf location loads its detail panel via `hx-get="/locations/{{.ID}}"`.

### Entity Sheet (`/entities/:id`)

Full entity view — stats, HP/FP bars, equipment, inventory, AI config, active quests.

```
┌──────────────────────────────────────┐
│  ◀ Back to entities                  │
├──────────┬───────────────────────────┤
│  Greta   │  Innkeeper                │
│  Human   │  Location: Frosthold Inn  │
│  LV 3    │  AI: guarded              │
├──────────┴───────────────────────────┤
│  STR ████████░░ 16  │  HP ██████░░ 28/40 │
│  DEX ██████░░░░ 12  │  FP ████░░░░ 18/26 │
│  CON ██████░░░░ 13  │                     │
│  INT ██████░░░░ 11  │  XP: 450/800       │
│  WIS ████████░░ 14  │                     │
│  CHA ████████░░ 15  │  Gold: 42           │
├──────────────────────┴──────────────────┤
│  Equipment           │  Inventory         │
│  Head: —             │  Ale x12           │
│  Body: Apron         │  Bread x5          │
│  Weapon: Dagger      │  Ragged Cloth x3   │
│  Offhand: —          │                    │
├──────────────────────┴──────────────────┤
│  Active Quests                          │
│  ☐ The Rat Problem — kill 3/8 rats      │
└──────────────────────────────────────────┘
```

HP/FP bars refresh via SSE or periodic polling:

```html
<div hx-get="/entities/{{.ID}}/vitals"
     hx-trigger="sse:vitals"
     hx-swap="innerHTML">
    {{ template "entity_hp_bar" . }}
</div>
```

### Combat View (`/combat`)

Lists active combat instances. Click to expand round-by-round log.

```
┌──────────────────────────────────────┐
│  Active Combats (3)                  │
├──────────────────────────────────────┤
│  ▶ Frosthold Gate                    │
│     Guard (12/20 HP) vs Wolf (4/18)  │
│     Round 4 — 2s ago                 │
│                                       │
│  ▶ Rat Infestation                   │
│     Player (34/40) vs Giant Rat x3   │
│     Round 2 — just now               │
│                                       │
│  ▶ Sewer Expedition                  │
│     Player (40/40) vs...             │
│     Round 1 — 3s ago                 │
└──────────────────────────────────────┘
```

### Quest Journal (`/quests`)

```
┌──────────────────────────────────────┐
│  Active Quests               [Filter]│
├──────────────────────────────────────┤
│  ☐ The Rat Problem          Lv 2     │
│     Investigate the cellar           │
│     Progress: 6/8 giant rats killed  │
│     Reward: 150 XP, 50 gold          │
│                                       │
│  ☐ Sewer Expedition        Lv 4     │
│     (locked — requires rat_problem)  │
│                                       │
├──────────────────────────────────────┤
│  Completed Quests                    │
│  ✓ Frosthold Arrival        Lv 1     │
│  ✓ First Lesson             Lv 1     │
└──────────────────────────────────────┘
```

### AI Debug Console (`/ai`)

For development: inspect entity AI state, execute Lua expressions, reload scripts.

```
┌──────────────────────────────────────┐
│  AI Scripts                    [Reload]│
├──────────────────────────────────────┤
│  merchant.lua       ✓ loaded         │
│  innkeeper.lua      ✓ loaded         │
│  patrol_guard.lua   ✓ loaded         │
│  dragon.lua         ✓ loaded         │
├──────────────────────────────────────┤
│  Entity AI Inspector                 │
│  Entity ID: [________________] [Go]  │
│                                       │
│  Type: scripted                      │
│  Script: merchant.lua                │
│  State: {"angry": false, ...}       │
│  [Edit State] [Eval Expression]     │
└──────────────────────────────────────┘
```

---

## HTMX Patterns

### Fragment Loading

Most page content is lazy-loaded via htmx to keep initial render fast.

```html
<!-- Entity list loads on page enter -->
<div hx-get="/entities"
     hx-trigger="load"
     hx-target="this"
     hx-swap="innerHTML">
    Loading entities...
</div>
```

### Infinite Scroll / Pagination

```html
<button hx-get="/entities?offset={{.NextOffset}}"
        hx-target="this"
        hx-swap="outerHTML">
    Load more
</button>
```

### Form Submission

```html
<form hx-post="/entities/{{.ID}}/action"
      hx-target="#action-result"
      hx-swap="innerHTML">
    <select name="action">
        <option value="move">Move</option>
        <option value="attack">Attack</option>
        <option value="use">Use Item</option>
    </select>
    <input name="target" placeholder="Target ID">
    <button type="submit">Execute</button>
</form>
```

### Real-Time Updates via SSE

The tick loop broadcasts events to an SSE hub:

```go
// internal/web/sse.go

type SSEHub struct {
    clients   map[chan SSEEvent]struct{}
    mu        sync.RWMutex
}

type SSEEvent struct {
    Event string      // Event name (used in sse-swap)
    Data  string      // JSON or HTML fragment
}

func (h *SSEHub) Broadcast(event string, data string) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    for ch := range h.clients {
        select {
        case ch <- SSEEvent{Event: event, Data: data}:
        default:
            // drop slow client
        }
    }
}
```

The tick loop calls `hub.Broadcast` for various event types:

```go
// Engine tick loop integration
func (tl *TickLoop) Tick() {
    tl.tick++
    // ... process world ...

    // Broadcast periodic updates
    if tl.tick%5 == 0 {
        tl.sseHub.Broadcast("clock", renderClockFragment(tl.world.Time))
    }
    if tl.tick%10 == 0 {
        tl.sseHub.Broadcast("vitals", renderVitalsFragment(tl.entities.Active()))
    }
    // On interesting events
    for _, event := range tl.world.DrainEvents() {
        tl.sseHub.Broadcast("notification", renderNotification(event))
    }
}
```

Browser receives SSE and htmx swaps matching elements:

```html
<div sse-swap="clock" hx-swap="innerHTML">
    <!-- auto-updated with World Clock fragment -->
</div>

<div id="notifications"
     sse-swap="notification"
     hx-swap="beforeend">
    <!-- notifications accumulate here -->
</div>
```

### Hyperscript for Sparse Client Logic

For minor client-side behavior that isn't worth a server round-trip, use [hyperscript](https://hyperscript.org/) (sibling project to htmx):

```html
<button _="on click toggle .collapsed on the next <ul/>">
    Toggle Children
</button>

<div _="on sse:notification add @class('flash') to me
         then wait 2s then remove @class('flash')">
</div>
```

Included via a tiny `<script>` tag. Optional — skip if not needed.

---

## API Routes (UI)

Mounted at `/` (separate from `/api/v1`):

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Dashboard |
| GET | `/locations` | World / location browser |
| GET | `/locations/:id` | Location detail page |
| GET | `/locations/:id/children` | HTMX fragment: child list |
| GET | `/entities` | Entity list with filters |
| GET | `/entities/:id` | Entity detail page |
| GET | `/entities/:id/vitals` | HTMX fragment: HP/FP bars |
| GET | `/combat` | Combat instances view |
| GET | `/combat/:id` | Combat detail with log |
| GET | `/quests` | Quest journal |
| GET | `/ai` | AI debug console |
| GET | `/api/v1/ui/events` | SSE endpoint for real-time updates |

---

## Static Assets

All assets embedded at compile time:

```go
// internal/web/static.go

package web

import "embed"

//go:embed templates/static/*
var StaticFS embed.FS

//go:embed templates/base.html
//go:embed templates/dashboard.html
//go:embed templates/locations.html
//go:embed templates/entity.html
//go:embed templates/fragments/*.html
var TemplateFS embed.FS
```

Gin setup:

```go
func SetupRoutes(router *gin.Engine, sim *engine.Simulation) {
    tmpl := template.Must(template.ParseFS(TemplateFS, "templates/*.html", "templates/fragments/*.html"))
    router.SetHTMLTemplate(tmpl)

    staticFS := gin.FS(StaticFS)
    router.StaticFS("/static", staticFS)

    handler := &Handler{sim: sim}
    router.GET("/", handler.Dashboard)
    router.GET("/locations", handler.LocationList)
    router.GET("/locations/:id", handler.LocationDetail)
    router.GET("/locations/:id/children", handler.LocationChildren)
    router.GET("/entities", handler.EntityList)
    router.GET("/entities/:id", handler.EntityDetail)
    router.GET("/entities/:id/vitals", handler.EntityVitals)
    router.GET("/combat", handler.CombatList)
    router.GET("/quests", handler.QuestJournal)
    router.GET("/ai", handler.AIConsole)
    router.GET("/api/v1/ui/events", handler.SSEEvents)
}
```

---

## Project Layout Additions

```
simuz/
├── internal/
│   ├── web/                 # Web UI
│   │   ├── handler.go
│   │   ├── sse.go
│   │   ├── routes.go
│   │   ├── static.go
│   │   └── templates/
│   │       ├── base.html
│   │       ├── dashboard.html
│   │       ├── locations.html
│   │       ├── location.html
│   │       ├── entities.html
│   │       ├── entity.html
│   │       ├── combat.html
│   │       ├── quests.html
│   │       ├── ai.html
│   │       ├── admin.html
│   │       ├── fragments/
│   │       │   ├── world_clock.html
│   │       │   ├── entity_list.html
│   │       │   ├── entity_vitals.html
│   │       │   ├── location_children.html
│   │       │   ├── combat_log.html
│   │       │   └── notification.html
│   │       └── static/
│   │           ├── css/app.css
│   │           └── js/app.js
│   ├── api/                 # JSON API handlers (unchanged)
│   └── ...
└── ...
```

---

## CSS Approach

No framework. Single `app.css` (~200 lines) with:

- CSS custom properties for theming (`--bg`, `--text`, `--accent`, `--hp-color`, etc.)
- Dark mode via `prefers-color-scheme`
- Utility classes for bars, cards, trees, tables, badges
- Responsive sidebar layout (nav collapses on mobile)

Keeps the binary small and avoids dependency churn.

---

## Dependencies Added

```
require (
    github.com/gin-gonic/gin          # Already in stack
    github.com/bigskysoftware/htmx    # Not a Go pkg — downloaded as static asset
)
```

HTMX is fetched at build time and embedded:

```go
//go:embed static/js/htmx.min.js
var htmxJS embed.FS
```

Alternatively, hotlink from CDN during dev:

```html
<script src="https://unpkg.com/htmx.org@2.0.0"></script>
```
