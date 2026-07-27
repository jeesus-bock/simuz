# AI & Scripting System

## Overview

Creature AI uses a two-tier system:

1. **Built-in archetypes** — common behaviors implemented in Go, set via `entity.ai_type`
2. **Scripted AI** — custom behaviors defined in Lua, enabled when `ai_type = "scripted"`

This keeps the hot path (90% of creatures) fast in Go while allowing full customization for unique NPCs, bosses, and quest entities.

---

## Built-in Archetypes

Archetypes are a single DB field on the entity. The engine dispatches to the correct Go behavior function each tick.

| Archetype | Enum Value | Behavior |
|-----------|-----------|----------|
| Passive | `passive` | Ignores all entities, follows daily routine (eat/sleep/wander). Default for civilians. |
| Aggressive | `aggressive` | Attacks any hostile entity in perception range. No warning. |
| Territorial | `territorial` | Attacks entities that enter a defined zone (location subtree). Issues warning first if `aggro_warning` is set. |
| Cowardly | `cowardly` | Flees when HP < threshold or facing stronger foe. May return with reinforcements. |
| Greedy | `greedy` | Prioritizes loot over combat. Picks up items, may attack弱者. Has betrayal chance. |
| Noble | `noble` | Protects weaker entities. Intervenes in nearby combat to help outnumbered side. Shares resources. |
| Curious | `curious` | Investigates sounds, lights, disturbances. Easily distracted. Low aggro threshold. |
| Guarded | `guarded` | Stationary at a post. Issues verbal warning, then counts down before attacking. |
| Patrol | `patrol` | Walks a predefined path (waypoints array). Reports intruders to nearby guards. |
| Scripted | `scripted` | Delegates all decisions to a Lua script. See Scripted AI section. |

### Entity Fields for AI

```go
type EntityAI struct {
    Type           AIArchetype   `json:"ai_type"`            // "aggressive", "scripted", etc.
    ScriptIDs      []string      `json:"script_ids,omitempty"` // Lua file basenames, used when Type == "scripted"
    FactionID      string        `json:"faction_id"`          // Determines hostile/ally relationships
    AggroRange     float64       `json:"aggro_range"`         // Perception range for aggro (meters), 0 = use species default
    AggroWarning   int           `json:"aggro_warning"`       // Ticks before attack (for guarded/territorial), 0 = instant
    Waypoints      []Waypoint    `json:"waypoints,omitempty"` // Patrol path
    HomeLocation   string        `json:"home_location"`       // Territory center (for territorial)
    TerritoryRadius float64      `json:"territory_radius"`    // Territory size
    ScriptState    json.RawMessage `json:"script_state,omitempty"` // Per-entity Lua state blob
}
```

### Tick Integration

```
func (e *Entity) TickAI() {
    switch e.AI.Type {
    case Aggressive:
        aiAggressive(e)
    case Cowardly:
        aiCowardly(e)
    case Scripted:
        aiScripted(e)   // Calls Lua VM
    // ...
    }
}
```

Each tick, the active AI function:
1. Checks perception — entities, items, sounds in range
2. Evaluates goals — attack, flee, investigate, patrol, rest
3. Queues actions — move, attack, use item, speak
4. Returns control to the tick loop

Actions are queued (not executed immediately) so the tick loop can process movement and combat in a deterministic order.

---

## Scripted AI (Lua)

### Why Lua

- **Pure Go** via [gopher-lua](https://github.com/yuin/gopher-lua) — no cgo
- **Tiny and fast** — VM startup ~microseconds, suitable per-tick
- **Game-proven** — WoW, Factorio, Roblox, Garry's Mod all use Lua for modding
- **Sandboxable** — limit execution time, memory, available globals
- **Familiar syntax** — low barrier for content creators

### Script Location & Lifecycle

```
scripts/
├── ai/
│   ├── merchant_lev.lua        # Unique NPC: custom haggle, flee, restock
│   ├── rat_king.lua            # Boss: summon minions, enrage phase
│   ├── dragon_elder.lua        # Boss: fly, breathe fire, land phases
│   └── lib/
│       ├── pathfinding.lua     # Shared: A* movement
│       ├── dialog.lua          # Shared: conditional speech
│       └── combat.lua          # Shared: target selection helpers
```

- Scripts are text files, loaded once at startup and compiled/cached in a `map[string]*lua.Function`
- Entity references a script by `script_id` (basename without `.lua`): `entity.ai.script_id = "merchant_lev"`
- Scripts can `require` other scripts from `scripts/ai/lib/` for shared utilities
- Hot-reload via API endpoint or SIGHUP for live editing

### Lua API Surface

The engine exposes a restricted set of Go functions to Lua. No file I/O, no OS calls, no network.

#### Perception

```lua
-- Get visible entities within range
local entities = ai.get_visible_entities(range)   -- returns array of entity tables

-- Get entity by ID
local target = ai.get_entity(entity_id)            -- returns entity table or nil

-- Check if entity can see position
local can_see = ai.can_see(x, y, z)

-- Get current location
local loc = ai.get_location()                      -- returns {id, name, type, parent_id}
```

#### Self

```lua
-- Read own stats (read-only)
local hp = self.hp
local max_hp = self.max_hp
local fp = self.fp
local pos = self.position          -- {x, y, z}
local attrs = self.attributes      -- {str, dex, con, int, wis, cha}
local skills = self.skills         -- {sword=12, archery=14, ...}
local inventory = self.inventory   -- array of item tables
local ai_state = self.state        -- custom Lua state (persisted as JSON)

-- Write to custom state (persisted)
self.state.angry = true
self.state.hunt_target = entity_id
```

#### Actions

```lua
-- Movement
ai.move_to(x, y, z)                              -- start travel path
ai.move_to_entity(entity_id, distance)            -- approach entity
ai.flee_from(entity_id, distance)                 -- run away
ai.wander(radius)                                  -- random movement within radius
ai.pause(ticks)                                    -- stand still for N ticks

-- Combat
ai.attack(entity_id)                              -- queue melee/ranged attack
ai.attack_location(entity_id, hit_location)       -- called shot
ai.use_skill(skill_id, target_id)                 -- e.g., "intimidate", "heal"

-- Equipment
ai.equip(item_id)                                 -- equip from inventory
ai.unequip(slot)                                  -- unequip a slot

-- Communication
ai.say(text)                                      -- speak (broadcast to location)
ai.emote(text)                                    -- *emote text*

-- Items
ai.pickup(item_entity_id)
ai.drop(item_id)
ai.use_item(item_id, target_id)                   -- use consumable, apply to target

-- World
ai.get_time()              -- returns {hour, minute, day, season, phase}
ai.get_weather()            -- returns {type, temperature, visibility}
ai.get_entities_at(location_id)  -- all entities in a location
```

#### Conditionals / Queries

```lua
-- Faction checks
ai.is_hostile(entity_id)           -- bool
ai.is_ally(entity_id)              -- bool
ai.get_reputation(faction_id)      -- int

-- Distance
ai.distance_to(entity_id)          -- meters
ai.distance_to_pos(x, y, z)        -- meters

-- Pathfinding
ai.has_path_to(x, y, z)            -- bool (reachable?)
ai.path_length_to(x, y, z)         -- estimated distance via path

-- Random (seeded per entity for reproducibility)
ai.random()                        -- float 0-1
ai.random_int(min, max)            -- int
ai.random_choice({"a", "b", "c"})

-- Math (std math library also available in Lua)
```

### Sandboxing

```go
// Each script runs in a new Lua VM instance (pooled for reuse)
// Restrictions enforced per VM:
// - Max execution time: 5ms per tick (killed if exceeded)
// - Max memory: 64KB Lua heap
// - No require("io"), require("os"), require("debug")
// - No loadstring, loadfile, dofile
// - No coroutines (single-shot per tick)
// - Available globals: math, table, string, ai.*, self
```

If a script exceeds limits, it's killed, logged, and the entity falls back to `passive` AI for the remainder of the tick.

### Script Example

```lua
-- scripts/ai/rat_king.lua
-- Boss AI with phases

function tick()
    local enemies = ai.get_visible_entities(20)
    local target = select_target(enemies)

    if not target then
        -- No enemies: patrol lair
        if not self.state.patrol_index then
            self.state.patrol_index = 1
        end
        local wp = self.waypoints[self.state.patrol_index]
        ai.move_to(wp.x, wp.y, wp.z)
        if ai.distance_to_pos(wp.x, wp.y, wp.z) < 1 then
            self.state.patrol_index = (self.state.patrol_index % #self.waypoints) + 1
        end
        return
    end

    -- Phase 2: below 50% HP, summon rats
    if self.hp / self.max_hp < 0.5 and not self.state.enraged then
        self.state.enraged = true
        ai.say("*The Rat King squeals an ear-piercing cry!*")
        ai.spawn_entities("giant_rat", 3, ai.get_location().id)
    end

    -- Combat logic
    local dist = ai.distance_to(target.id)
    if dist < 2 then
        ai.attack(target.id)
    elseif dist < 10 then
        ai.move_to_entity(target.id, 1.5)
    else
        -- Charge
        ai.move_to_entity(target.id, 1.5)
    end
end

function select_target(enemies)
    local best = nil
    local best_score = -999
    for _, e in ipairs(enemies) do
        if ai.is_hostile(e.id) then
            local score = e.hp / e.max_hp * -10  -- prioritize low HP
                + ai.distance_to(e.id) * -1       -- prefer closer
                + (e.attributes.str or 10)         -- prefer stronger
            if score > best_score then
                best_score = score
                best = e
            end
        end
    end
    return best
end
```

### Script State Persistence

The `self.state` Lua table is serialized to JSON at each save and stored in the entity's `ai_data` column:

```json
{
  "patrol_index": 3,
  "enraged": true,
  "hunt_target": "player_001"
}
```

On load, the JSON is deserialized back into a Lua table before the first tick. This keeps script state durable while the script logic stays on disk.

---

## AI Evaluation Order

Each tick, entities are processed in a deterministic order (by UUID, stable sort):

1. **Perception update** — refresh visible entities, items, sounds
2. **Goal evaluation** — archetype or Lua script decides next action
3. **Action queuing** — action appended to entity's action queue
4. **Action resolution** — movement processed, attacks queued to combat engine

```
Tick N
  ├── 1. Scheduler: fire due events
  ├── 2. World: weather, time, spawns
  ├── 3. For each entity (sorted by UUID):
  │     ├── a. AI: update perception → choose goal → queue action
  │     ├── b. Movement: resolve position changes
  │     ├── c. Status: apply DOTs, regen, environmental effects
  │     └── d. Quest: check active objective progress
  ├── 4. Combat: resolve all queued attacks
  ├── 5. Quest: check global failure conditions
  └── 6. Save: if save_interval elapsed
```

---

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/ai/archetypes` | List all built-in archetypes with descriptions |
| GET | `/ai/scripts` | List available Lua scripts |
| GET | `/ai/scripts/:id` | Get script source |
| POST | `/ai/scripts/:id/reload` | Hot-reload a single script from disk |
| POST | `/ai/scripts/reload` | Reload all scripts |
| GET | `/entities/:id/ai` | Get entity AI config |
| PUT | `/entities/:id/ai` | Update entity AI config (type, script_id, params) |
| POST | `/entities/:id/ai/debug/eval` | Run a Lua expression in the entity's VM and return result |

---

## Project Layout Additions

```
simuz/
├── internal/
│   ├── ai/                    # AI system
│   │   ├── archetypes.go      # Built-in behavior implementations
│   │   ├── scripted.go        # Lua VM pool, script loading, execution
│   │   ├── perception.go      # Vision, hearing, awareness
│   │   ├── sandbox.go         # Lua sandbox constraints
│   │   └── api.go             # Lua function bindings (ai.*)
│   └── ...
├── scripts/
│   ├── ai/                    # Lua AI scripts
│   │   ├── merchant_lev.lua
│   │   ├── rat_king.lua
│   │   └── lib/               # Shared Lua libraries
│   │       ├── pathfinding.lua
│   │       └── combat.lua
│   └── ...
└── ...
```
