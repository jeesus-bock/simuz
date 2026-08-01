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
│   ├── templates/       # YAML generation templates
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
                              └─────┬─────────┘
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
- Cron-like: fire at specific simulated times (e.g., "every dawn", "every 15 min")

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

### Base Path

All API routes are mounted under `/api/v1`.

### Response Envelope

All API responses use a consistent JSON envelope:

```json
{
  "ok": true,
  "data": { ... },
  "error": ""
}
```

On error:

```json
{
  "ok": false,
  "data": null,
  "error": "human-readable error message"
}
```

### World Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/world` | Current world state (tick, time, weather, speed) |
| GET | `/api/v1/world/locations/:id` | Single location by ID, including children and exits |
| POST | `/api/v1/world/tick` | Advance simulation by N ticks |
| PUT | `/api/v1/world/speed` | Change simulation speed (game-minutes per tick) |

**GET `/api/v1/world` Response:**

```json
{
  "ok": true,
  "data": {
    "tick": 14235,
    "day": 3,
    "hour": 14,
    "minute": 30,
    "phase": "day",
    "season": "summer",
    "weather": { "type": "clear", "temperature": 18, "visibility": 10 },
    "speed": 24,
    "ticks_per_game_day": 60
  },
  "error": ""
}
```

**POST `/api/v1/world/tick` Request Body:**

```json
{ "ticks": 10 }
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ticks` | uint64 | yes | Number of ticks to advance (1–1000) |

**PUT `/api/v1/world/speed` Request Body:**

```json
{ "speed": 48 }
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `speed` | int | yes | Game-minutes per tick (1–1440) |

### Entity Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/entities` | List entities with optional filters |
| GET | `/api/v1/entities/:id` | Full entity details by ID |
| POST | `/api/v1/entities` | Spawn a new entity |
| POST | `/api/v1/entities/:id/action` | Queue an action for an entity |

**GET `/api/v1/entities` — Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `location_id` | string | no | Filter by location |
| `species` | string | no | Filter by species |
| `faction` | string | no | Filter by faction |
| `profession` | string | no | Filter by profession |
| `alive` | bool | no | Filter by alive status |
| `limit` | int | no | Max results (default 100, max 1000) |
| `offset` | int | no | Pagination offset (default 0) |

**GET `/api/v1/entities/:id` — Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | Entity ID |

**Response:**

```json
{
  "ok": true,
  "data": {
    "id": "player_001",
    "name": "Aldric",
    "species": "human",
    "type": "creature",
    "location_id": "frosthold_inn",
    "travel_state": null,
    "attributes": { "str": 14, "dex": 12, "con": 13, "int": 10, "wis": 12, "cha": 11 },
    "derived": {
      "max_hp": 40, "max_fp": 26, "base_speed": 5,
      "initiative": 6, "carry_capacity": 70, "natural_dr": 0
    },
    "skills": { "sword": 12, "archery": 8 },
    "hit_points": { "current": 28, "max": 40 },
    "fatigue_points": { "current": 18, "max": 26 },
    "status_effects": [],
    "inventory": ["item_001", "item_002"],
    "equipped": { "head": null, "body": "leather_jacket", "weapon": "shortsword", "offhand": null },
    "behavior": null,
    "circadian": "diurnal",
    "age": { "years": 25, "mature": 16, "max_lifespan": 80 },
    "faction": "frosthold",
    "knowledge": { "lore_local": 3, "persuasion": 5 }
  },
  "error": ""
}
```

**POST `/api/v1/entities` — Request Body:**

```json
{
  "species": "human",
  "name": "New NPC",
  "location_id": "frosthold_inn",
  "faction": "frosthold",
  "profession": "innkeeper",
  "level": 3,
  "attributes": { "str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10 },
  "ai_type": "guarded",
  "faction_id": "frosthold"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `species` | string | yes | Species key (e.g. "human", "orc", "elf") |
| `name` | string | yes | Display name |
| `location_id` | string | yes | Starting location ID |
| `faction` | string | no | Faction membership |
| `profession` | string | no | Profession (e.g. "innkeeper", "guard") |
| `level` | int | no | Starting level (default 1) |
| `attributes` | object | no | STR/DEX/CON/INT/WIS/CHA overrides |
| `ai_type` | string | no | AI archetype (default "passive") |
| `faction_id` | string | no | Faction ID for relations |

**POST `/api/v1/entities/:id/action` — Request Body:**

```json
{
  "action": "move",
  "target": "frosthold_gate",
  "params": {}
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `action` | string | yes | Action type: `move`, `attack`, `use_item`, `talk_to`, `deliver_item`, `pickup`, `drop`, `equip`, `unequip`, `use_skill` |
| `target` | string | depends | Target entity ID, location ID, or item ID depending on action |
| `params` | object | no | Additional action parameters |

**Response:**

```json
{
  "ok": true,
  "data": {
    "action": "move",
    "entity_id": "player_001",
    "from": "frosthold_inn",
    "to": "frosthold_gate",
    "status": "queued",
    "eta_tick": 14250
  },
  "error": ""
}
```

### Combat Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/combat` | List all active combat instances |
| GET | `/api/v1/combat/:id` | Get a specific combat instance with round log |

**GET `/api/v1/combat` — Response:**

```json
{
  "ok": true,
  "data": {
    "combats": [
      {
        "id": "combat_001",
        "location_id": "frosthold_gate",
        "participants": ["player_001", "wolf_003"],
        "round": 4,
        "started_tick": 14230,
        "status": "active"
      }
    ]
  },
  "error": ""
}
```

**GET `/api/v1/combat/:id` — Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | Combat instance ID |

**Response:**

```json
{
  "ok": true,
  "data": {
    "id": "combat_001",
    "location_id": "frosthold_gate",
    "participants": [
      { "entity_id": "player_001", "hp": {"current": 28, "max": 40}, "fp": {"current": 18, "max": 26}, "status": "active" },
      { "entity_id": "wolf_003", "hp": {"current": 4, "max": 18}, "fp": {"current": 0, "max": 10}, "status": "active" }
    ],
    "round": 4,
    "log": [
      { "tick": 14231, "attacker": "player_001", "defender": "wolf_003", "hit": true, "damage": 6, "location": "torso" },
      { "tick": 14232, "attacker": "wolf_003", "defender": "player_001", "hit": true, "damage": 3, "location": "leg" }
    ],
    "status": "active"
  },
  "error": ""
}
```

### Bestiary Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/bestiary` | List all species with optional filters |
| GET | `/api/v1/bestiary/:species` | Get details for a single species |

**GET `/api/v1/bestiary` — Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `habitat` | string | no | Filter by habitat (e.g. "forest", "underground") |
| `size` | string | no | Filter by size (e.g. "small", "large", "huge") |
| `civilized` | bool | no | Filter by civilized vs wild |

**GET `/api/v1/bestiary/:species` — Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `species` | string | yes | Species key (e.g. "human", "orc", "goblin") |

**Response:**

```json
{
  "ok": true,
  "data": {
    "id": "human",
    "name": "Human",
    "attributes": { "str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10 },
    "hp": 20,
    "speed": 5,
    "lifespan": "80 years",
    "traits": ["versatile"],
    "habitat": "all",
    "culture": "Diverse, ambitious",
    "subraces": []
  },
  "error": ""
}
```

### Items Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/items` | List all item definitions with optional filters |
| GET | `/api/v1/items/:id` | Get a single item definition by ID |

**GET `/api/v1/items` — Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | string | no | Filter by item type (weapon, armor, consumable, etc.) |
| `slot` | string | no | Filter by equipment slot |

**GET `/api/v1/items/:id` — Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | Item definition ID |

**Response:**

```json
{
  "ok": true,
  "data": {
    "id": "shortsword",
    "name": "Shortsword",
    "type": "weapon",
    "damage_dice": "1d6",
    "damage_type": "cut",
    "strength_min": 8,
    "reach": 1,
    "weight": 1.5,
    "value": 150,
    "description": "A short, light blade suitable for thrusting or swinging."
  },
  "error": ""
}
```

### AI Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/ai/archetypes` | List all built-in AI archetypes with descriptions |
| GET | `/api/v1/ai/scripts` | List available Lua scripts |
| GET | `/api/v1/ai/scripts/:id` | Get script source by name |
| POST | `/api/v1/ai/scripts/:id/reload` | Hot-reload a single script from disk |
| POST | `/api/v1/ai/scripts/reload` | Reload all scripts |
| GET | `/api/v1/entities/:id/ai` | Get entity AI config |
| PUT | `/api/v1/entities/:id/ai` | Update entity AI config |
| POST | `/api/v1/entities/:id/ai/debug/eval` | Run a Lua expression in the entity's VM |

**GET `/api/v1/ai/archetypes` — Response:**

```json
{
  "ok": true,
  "data": [
    { "id": "passive", "name": "Passive", "description": "Ignores all entities, follows daily routine" },
    { "id": "aggressive", "name": "Aggressive", "description": "Attacks any hostile entity in perception range" },
    { "id": "scripted", "name": "Scripted", "description": "Delegates all decisions to a Lua script" }
  ],
  "error": ""
}
```

**GET `/api/v1/ai/scripts` — Response:**

```json
{
  "ok": true,
  "data": [
    { "id": "merchant_lev", "name": "merchant_lev.lua", "description": "Custom haggle, flee, restock" },
    { "id": "rat_king", "name": "rat_king.lua", "description": "Boss: summon minions, enrage phase" }
  ],
  "error": ""
}
```

**GET `/api/v1/ai/scripts/:id` — Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | Script name without `.lua` extension |

**Response:**

```json
{
  "ok": true,
  "data": { "id": "rat_king", "source": "-- Boss AI with phases\n\nfunction tick()\n    ..." },
  "error": ""
}
```

**POST `/api/v1/ai/scripts/:id/reload` — Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | Script name without `.lua` extension |

**Response:**

```json
{ "ok": true, "data": { "id": "rat_king", "status": "reloaded" }, "error": "" }
```

**GET `/api/v1/entities/:id/ai` — Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | Entity ID |

**Response:**

```json
{
  "ok": true,
  "data": {
    "entity_id": "goblin_001",
    "ai_type": "scripted",
    "script_ids": ["goblin_ambush"],
    "faction_id": "goblin_warband",
    "aggro_range": 15.0,
    "aggro_warning": 0,
    "home_location": "goblin_cave_entrance",
    "territory_radius": 50.0
  },
  "error": ""
}
```

**PUT `/api/v1/entities/:id/ai` — Request Body:**

```json
{
  "ai_type": "scripted",
  "script_ids": ["goblin_ambush"],
  "faction_id": "goblin_warband",
  "aggro_range": 15.0,
  "aggro_warning": 0,
  "home_location": "goblin_cave_entrance",
  "territory_radius": 50.0
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ai_type` | string | no | AI archetype ("passive", "aggressive", "scripted", etc.) |
| `script_ids` | array[string] | no | Lua script basenames (used when ai_type is "scripted") |
| `faction_id` | string | no | Faction for relationship calculations |
| `aggro_range` | float64 | no | Perception range for aggro in meters (0 = use species default) |
| `aggro_warning` | int | no | Ticks before attack (for guarded/territorial), 0 = instant |
| `home_location` | string | no | Territory center (for territorial archetype) |
| `territory_radius` | float64 | no | Territory size in meters |

**POST `/api/v1/entities/:id/ai/debug/eval` — Request Body:**

```json
{ "expression": "self.hp / self.max_hp" }
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `expression` | string | yes | Lua expression to evaluate in the entity's VM |

**Response:**

```json
{
  "ok": true,
  "data": {
    "expression": "self.hp / self.max_hp",
    "result": 0.75,
    "result_type": "number"
  },
  "error": ""
}
```

### Quest Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/quests` | List available quests visible to a specific entity |
| GET | `/api/v1/quests/:id` | Get quest definition and details by quest ID |
| POST | `/api/v1/quests/:id/accept` | Accept a quest |
| GET | `/api/v1/entities/:id/quests` | Get all quest states for an entity |

**GET `/api/v1/quests` — Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `entity_id` | string | yes | Entity ID to filter visible quests |
| `type` | string | no | Filter by quest type (main, side, faction, daily) |
| `state` | string | no | Filter by state (inactive, active, completed, failed) |

**Response:**

```json
{
  "ok": true,
  "data": [
    {
      "id": "rat_problem",
      "title": "The Rat Problem",
      "type": "side",
      "level": 2,
      "state": "active",
      "current_stage": "kill_rats",
      "objectives": {
        "kill_rats_main": { "description": "Kill 8 giant rats", "progress": 5, "count": 8, "optional": false },
        "kill_rat_king": { "description": "Slay the Rat King", "progress": 0, "count": 1, "optional": true }
      },
      "rewards": { "experience": 150, "gold": 50 }
    }
  ],
  "error": ""
}
```

**GET `/api/v1/quests/:id` — Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | Quest ID |

**Response:**

```json
{
  "ok": true,
  "data": {
    "id": "rat_problem",
    "title": "The Rat Problem",
    "type": "side",
    "level": 2,
    "description": "Innkeeper Greta has a rat infestation in her cellar.",
    "source": { "type": "npc", "npc_id": "innkeeper_greta", "location_id": "frosthold_inn" },
    "stages": [
      {
        "id": "investigate",
        "name": "Investigate the Cellar",
        "description": "Head down to the inn's cellar.",
        "requirements": [],
        "objectives": [
          { "id": "enter_cellar", "type": "visit_location", "location_id": "frosthold_inn_cellar", "count": 1, "progress": 1 }
        ]
      }
    ],
    "rewards": { "experience": 150, "gold": 50, "items": [{ "template": "health_potion", "count": 2 }] },
    "failure_conditions": [{ "type": "time_limit", "hours": 72 }]
  },
  "error": ""
}
```

**POST `/api/v1/quests/:id/accept` — Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | Quest ID |

**Request Body:**

```json
{ "entity_id": "player_001" }
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `entity_id` | string | yes | The entity accepting the quest |

**Response:**

```json
{
  "ok": true,
  "data": {
    "quest_id": "rat_problem",
    "entity_id": "player_001",
    "state": "active",
    "current_stage": "investigate",
    "accepted_tick": 14235
  },
  "error": ""
}
```

**GET `/api/v1/entities/:id/quests` — Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | Entity ID |

**Response:**

```json
{
  "ok": true,
  "data": {
    "entity_id": "player_001",
    "quests": {
      "rat_problem": {
        "state": "ACTIVE",
        "current_stage": "kill_rats",
        "completed_stages": ["investigate"],
        "objectives": { "kill_rats_main": 5, "kill_rat_king": 0 },
        "variables": {},
        "accepted_tick": 14235
      }
    }
  },
  "error": ""
}
```

### Server-Sent Events (SSE)

The tick loop broadcasts events to connected SSE clients at `/api/v1/ui/events`.

**Connection:**

```
GET /api/v1/ui/events
Accept: text/event-stream
```

**Event Types:**

| Event | Data Format | Frequency |
|-------|-------------|-----------|
| `clock` | `{tick, day, hour, minute, phase, season}` | Every 5 ticks |
| `vitals` | `{entity_id, hp_current, hp_max, fp_current, fp_max}` | Every 10 ticks |
| `notification` | `{type, message, tick}` | On interesting events |
| `combat_update` | `{combat_id, round, log_entry}` | On combat actions |
| `quest_update` | `{entity_id, quest_id, stage, objective_progress}` | On quest state change |

**SSE Event Stream Format:**

```
event: clock
data: {"tick":14235,"day":3,"hour":14,"minute":30,"phase":"day","season":"summer"}

event: notification
data: {"type":"combat","message":"Wolf attacks Aldric!","tick":14235}
```

### Error Responses

All error responses use the standard envelope with `ok: false`:

```json
{
  "ok": false,
  "data": null,
  "error": "quest 'rat_problem' not found"
}
```

**Common HTTP Status Codes:**

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad request — missing or invalid parameters |
| 404 | Not found — resource does not exist |
| 409 | Conflict — action cannot be performed (e.g., quest already accepted) |
| 422 | Unprocessable entity — prerequisites not met |
| 500 | Internal server error |

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
