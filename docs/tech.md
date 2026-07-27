# Tech Stack & Architecture

## Stack Overview

| Layer        | Choice        | Rationale |
|-------------|---------------|-----------|
| Language    | Go 1.22+      | Concurrency, performance, single binary deploy |
| HTTP Router | Gin           | Fast, well-documented, middleware ecosystem |
| Database    | SQLite        | Embedded, zero-config, good enough for simulation |
| Wire Format | JSON          | Universal, debuggable, good-enough perf |
| Auth        | None (MVP)    | Local/dev-only initially |

## Project Layout

```
simuz/
├── cmd/
│   └── simuz/           # Main entrypoint
├── internal/
│   ├── ai/              # AI & scripting system
│   │   ├── archetypes.go
│   │   ├── scripted.go
│   │   ├── perception.go
│   │   ├── sandbox.go
│   │   └── api.go
│   ├── api/             # HTTP handlers (Gin routes)
│   │   ├── router.go
│   │   ├── world.go
│   │   ├── entities.go
│   │   ├── ai.go
│   │   ├── combat.go
│   │   └── quests.go
│   ├── engine/          # Core simulation loop
│   │   ├── tick.go
│   │   ├── scheduler.go
│   │   └── time.go
│   ├── web/             # Web UI (htmx + Go templates)
│   │   ├── handler.go
│   │   ├── sse.go
│   │   ├── routes.go
│   │   ├── static.go
│   │   └── templates/
│   │       ├── base.html
│   │       ├── dashboard.html
│   │       ├── locations.html
│   │       ├── entity.html
│   │       ├── combat.html
│   │       ├── quests.html
│   │       ├── ai.html
│   │       ├── admin.html
│   │       ├── fragments/
│   │       └── static/
│   ├── gen/             # World generation & seed data
│   │   ├── world.go
│   │   ├── biome.go
│   │   ├── region.go
│   │   ├── settlement.go
│   │   ├── dungeon.go
│   │   ├── populate.go
│   │   ├── items.go
│   │   ├── quests.go
│   │   ├── templates.go
│   │   ├── rng.go
│   │   └── output.go
│   ├── world/           # World model
│   │   ├── location.go
│   │   ├── travel.go
│   │   ├── weather.go
│   │   └── time.go
│   ├── entity/          # Entity/creature model
│   │   ├── entity.go
│   │   ├── attributes.go
│   │   ├── skills.go
│   │   └── behavior.go
│   ├── combat/          # Combat resolution
│   │   ├── attack.go
│   │   ├── damage.go
│   │   ├── armor.go
│   │   └── wounds.go
│   ├── items/           # Equipment & items
│   │   ├── item.go
│   │   ├── weapon.go
│   │   ├── armor.go
│   │   └── crafting.go
│   ├── quest/           # Quest engine
│   │   ├── definition.go
│   │   ├── engine.go
│   │   ├── conditions.go
│   │   ├── objectives.go
│   │   ├── triggers.go
│   │   └── store.go
│   └── storage/         # Persistence layer
│       ├── interface.go
│       └── sqlite.go
├── data/
│   ├── templates/       # YAML generation templates (biomes, settlements, dungeons, etc.)
│   ├── worlds/          # Generated world output (gitignored)
│   ├── seed/            # Hand-crafted seed data overrides
│   └── names/           # Name corpora for generation
├── quests/              # YAML quest definitions
├── scripts/             # Lua scripts
│   └── ai/              # AI scripts + lib/
├── docs/                # Design documents
├── go.mod
└── go.sum
```

## Architecture

```
┌──────────┐     HTTP/JSON      ┌───────────┐
│  Client   │ ──────────────→   │   Gin     │
│ (CLI/GUI) │ ←──────────────   │   API     │
└──────────┘                    └─────┬─────┘
                                      │
                                      ▼
                              ┌───────────────┐
                              │   Engine      │
                              │  (Tick Loop)  │
                              └───────┬───────┘
                                      │
              ┌───────────┬───────────┼───────────┬───────────┐
              ▼           ▼           ▼           ▼           ▼
        ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
        │ World   │ │ Entity  │ │ Combat  │ │ Items   │ │Scheduler│
        │ Model   │ │ Model   │ │ Engine  │ │ Model   │ │(Events) │
        └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘
              │           │           │           │
              ▼           ▼           ▼           ▼
        ┌─────────────────────────────────────────────┐
        │           Storage Interface                 │
        │       (Optional persistence toggle)         │
        └───────────────────┬─────────────────────────┘
                            ▼
                    ┌──────────────┐
                    │   SQLite     │
                    └──────────────┘
```

## Tick Loop

- Main loop runs at 1 Hz (one tick per second)
- Each tick:
  1. Process scheduled events (real-time triggers)
  2. Update entity positions for in-progress travel
  3. Process entity AI/behavior
  4. Resolve queued combat actions
  5. Evaluate active quest objectives and failure conditions
  6. Apply environmental effects (weather, hunger, etc.)
  7. Save state (if persistence enabled)

```
type TickLoop struct {
    tick      uint64
    scheduler *Scheduler
    world     *World
    entities  *EntityManager
    combat    *CombatEngine
    quests    *quest.QuestManager
    storage   Storage
}

func (tl *TickLoop) Run() {
    ticker := time.NewTicker(1 * time.Second)
    for range ticker.C {
        tl.tick++
        tl.scheduler.ProcessDue(tl.tick)
        tl.world.Tick()
        tl.entities.Tick()
        tl.combat.Tick()
        tl.quests.ProcessTick(tl.tick)
        if tl.tick%saveInterval == 0 {
            tl.storage.Save()
        }
    }
}
```

## Scheduler (Real-Time Triggers)

```
type ScheduledEvent struct {
    ID        string
    Tick      uint64    // First occurrence
    Interval  uint64    // 0 = one-shot
    Action    func()
}

type Scheduler struct {
    events []ScheduledEvent
}
```

Events can be:
- One-shot: fire at tick X
- Repeating: fire every N ticks
- Cron-like: fire at specific simulated times (e.g., "every dawn")

## Database

SQLite via `modernc.org/sqlite` (pure Go, no CGO) or `mattn/go-sqlite3`.

Key tables:
- `locations` — hierarchical world locations
- `entities` — all creatures and objects (includes `ai_type`, `script_id`, `ai_data` JSON columns)
- `items` — inventory items
- `combat_log` — combat history
- `world_state` — time, weather, global state
- `entity_quests` — per-entity quest state (current stage, objective progress, flags)
- `quest_flags` — global quest-related flags

The storage layer uses an interface so the DB can be disabled for ephemeral sessions:

```
type Storage interface {
    Save(w *World) error
    Load() (*World, error)
    Enabled() bool
}
```

## API Design

Base path: `/api/v1`

| Method | Path | Description |
|--------|------|-------------|
| GET    | /world         | World state, time, weather |
| GET    | /world/locations/:id | Location details |
| POST   | /world/tick    | Advance N ticks (for manual control) |
| GET    | /entities      | List entities with filters (location, species, etc.) |
| GET    | /entities/:id  | Entity details, stats, inventory |
| POST   | /entities      | Spawn a new entity |
| POST   | /entities/:id/action | Queue an action (move, attack, use item) |
| GET    | /combat        | Active combat instances |
| GET    | /bestiary      | Species reference data |
| GET    | /bestiary/:species | Single species entry |
| GET    | /items         | Item definitions |
| GET    | /items/:id     | Single item definition |
| GET    | /ai/archetypes | List built-in AI archetypes |
| GET    | /ai/scripts    | List available Lua scripts |
| POST   | /ai/scripts/reload | Reload all scripts from disk |
| GET    | /entities/:id/ai | Get entity AI config |
| PUT    | /entities/:id/ai | Update entity AI config |
| GET    | /quests        | Available quests (visible to entity) |
| GET    | /quests/:id    | Quest definition and details |
| POST   | /quests/:id/accept | Accept a quest |
| GET    | /entities/:id/quests | Entity's active/completed quests |
| POST   | /admin/generate     | Trigger world generation with seed/params |
| GET    | /admin/generate/status | Generation progress |

UI routes (HTML, mounted at `/`):

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Dashboard |
| GET | `/locations` | World / location browser |
| GET | `/locations/:id` | Location detail page |
| GET | `/entities` | Entity list with filters |
| GET | `/entities/:id` | Entity detail page |
| GET | `/combat` | Combat instances view |
| GET | `/quests` | Quest journal |
| GET | `/ai` | AI debug console |
| GET | `/api/v1/ui/events` | SSE endpoint for real-time updates |

All JSON API responses follow `{"error": "message"}` error format.

## Concurrency

- The tick loop runs in a single goroutine (no concurrent world mutations)
- API reads are served from a RWMutex-protected snapshot or use a read-only reference
- Writes via API queue into a channel consumed by the tick loop
- This avoids the need for fine-grained locking

## Dependencies

```
require (
    github.com/gin-gonic/gin
    modernc.org/sqlite         // or github.com/mattn/go-sqlite3
)
```

Standard library only for everything else (encoding/json, net/http, time, sync, etc.).
