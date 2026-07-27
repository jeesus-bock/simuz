# Simuz — World Simulation

## Meta

| Field | Value |
|-------|-------|
| LastUpdated | 2026-07-27 |
| Status | active |
| Scripts | deity, fisherman, guard, priest, rat_king, traveling_salesman, defensive, hunting, gathering, healing, scouting, aggressive, bard, farmer, kobold, fairy, vampire, hag, goblin_ambush, blacksmith, innkeeper, miner, herbalist, child, bard_patron, thief, ranger, cultist, werewolf, bandit_chief, necromancer, dragon |
| Go Archetypes | passive, guarding, dormant (all hostile/factional entities now use Lua aggressive.lua) |

## Project Overview

Simuz is a Go-based world simulation game. The core loop advances ticks at 1 Hz. Each tick processes scheduled events, entity AI, combat, quests, weather, and optionally persists state to SQLite via `Storage.Save()`.

The Lua AI scripting system uses `gopher-lua` (`github.com/yuin/gopher-lua`). Scripts in `internal/ai/scripts/*.lua` are auto-loaded by `InitScripts()` into a `ScriptManager` keyed by script name. Entities with `AI.Type == "scripted"` run all scripts in `AI.ScriptIDs` each tick (sequentially, same entity scope).

The web UI uses Go + Gin + HTMX 2.0.1 + SSE extension. Dashboard, entities, locations, combat, quests, and AI pages all live-update via SSE fragment endpoints on every tick.

## Lua API Reference

### `self` (entity globals)

| Field | Type | Description |
|-------|------|-------------|
| `self.name` | string | Entity display name |
| `self.id` | string | Entity unique ID |
| `self.species` | string | Species tag |
| `self.faction` | string | Entity faction |
| `self.loc_id` | string | Current location ID |
| `self.hp` | number | Current HP |
| `self.max_hp` | number | Max HP |
| `self.fp` | number | Current FP |
| `self.max_fp` | number | Max FP |
| `self.level` | number | Entity level |
| `self.xp` | number | Entity XP |
| `self.age` | number | Entity age in ticks |
| `self.hunger` | number | Hunger ratio 0.0–1.0 (0=full, 1=starving); always 0 for auto-forage species |
| `self.inventory` | table | List of item DefIDs (unequipped, non-currency) |
| `self.home` | string | Home location ID (nil if none) |
| `self.state` | table | Activity state: `{activity=string, since_tick=number, until_tick=number, location_id=string}` |
| `self.mood` | string | Current mood: "neutral", "happy", "stressed", "tired", "relaxed", "drunk", "inspired", etc. |

### `world` table

| Function | Signature | Description |
|----------|-----------|-------------|
| `world.tick` | number | Current simulation tick |
| `world.time` | string | Formatted time string |
| `world.phase` | string | Current day phase ("dawn","day","dusk","night") |
| `world.day` | number | Current day number |
| `world.day_of_week` | number | Day of week (1-7, Monday-Sunday) |
| `world.day_name` | string | Day name ("Monday", "Tuesday", etc.) |
| `world.location_name(id)` | `(string, string)` | Get location display name for an ID |
| `world.exits_from(id)` | `(string, table)` | List child location IDs reachable from a parent |
| `world.travel_exits(id)` | `(table)` | Graph exits `{target_id, direction, distance}` for location and its region |
| `world.weather([id])` | `(table|nil)` | Effective weather at loc (or self): type, temp, visibility, wind, humidity, harsh, stormy, vis_mod, travel_mod |
| `world.location_control([id])` | `(table|nil)` | `{faction, strength}` controlling faction for location |
| `world.is_traveling()` | `(boolean)` | True if self is mid multi-tick cross-region travel |
| `world.parent_location(id)` | `(string, string|nil)` | Get parent location ID for a location, or nil |
| `world.entities_at(id)` | `(string, table)` | List alive entity IDs at any location by ID |
| `world.move_to(id)` | `(boolean)` | Instant move within same region/city tree; multi-tick travel across regions |
| `world.nearby_entities()` | `(string, table)` | List alive entity IDs at `self.loc_id` (excludes self) |
| `world.entity_name(id)` | `(string, string|nil)` | Get entity name for ID, or nil |
| `world.entity_items(id)` | `(string, table|nil)` | List unequipped item DefIDs for entity, or nil |
| `world.entity_info(id)` | `(table)` | Get detailed entity info table (name, species, faction, hp, max_hp, level, xp, age, alive, location_id, hunger) |
| `world.feed(entity_id)` | `(boolean)` | Reset entity's LastMealTick to current tick (returns true if target exists and alive) |
| `world.is_hostile(factionA, factionB)` | `(boolean)` | Check if two factions are hostile |
| `world.attack(attacker_id, target_id)` | `(boolean)` | Perform combat attack between two entity IDs |
| `world.add_item(item_def_id)` | `(boolean)` | Add an item instance to self's inventory by def ID |
| `world.try_buy(seller_id, item_def_id)` | `(table)` | Buy item from seller; `{done=true/false, price=number}` |
| `world.try_sell(buyer_id, item_def_id)` | `(table)` | Sell item to buyer; `{done=true/false, price=number}` |
| `world.divine_intervention(deity_id, target_id, type, ...)` | `(table)` | Perform a divine intervention; see types below |
| `world.use_item(def_id)` | `(boolean)` | Consume a substance item from self's inventory, apply its effect (boost then crash), remove it |
| `world.steal(target_id, item_def_id)` | `(table)` | Remove item from target's inventory and add to self; `{done=true/false}` |
| `world.damage_location(attacker_id, amount)` | `(table)` | Deal `amount` damage to all non-friendly entities at attacker's location; `{targets=number}` |
| `world.heal(target_id, amount)` | `(boolean)` | Heal target entity for `amount` HP (returns true if target exists and alive) |

### `util` table

| Function | Signature | Description |
|----------|-----------|-------------|
| `util.rand_int(n)` | `(number)` | Random int in `[0, n)` |
| `util.log(msg)` | `()` | Log message to server log |
| `util.mem_set(key, val)` | `()` | Store a value in entity memory flags (persisted to SQLite) |
| `util.mem_get(key)` | `(any)` | Retrieve a value from entity memory flags |
| `util.json_encode(val)` | `(string)` | JSON-encode a value |
| `util.json_decode(s)` | `(any)` | JSON-decode a string |
| `util.set_mood(mood)` | `()` | Set entity mood (affects behavior and interactions) |

## New Engine Hooks (Go)
- `onDragStart(draggerID, targetID)` – fires when an entity begins dragging another (e.g., leashed dog).
- `onDragEnd(draggerID, targetID)` – fires when dragging ends.
- `onRescueStart(rescuerID, targetID)` – fires when a rescue mission begins.
- `onRescueComplete(rescuerID, targetID)` – fires when rescue successfully completes.

## New API Endpoints (HTTP JSON)
- `POST /api/v1/drag` – body `{ "dragger_id": "E1", "target_id": "E2" }`
- `POST /api/v1/rescue/start` – body `{ "rescuer_id": "E1", "target_id": "E2" }`
- `POST /api/v1/rescue/complete` – body `{ "rescuer_id": "E1", "target_id": "E2" }`

These endpoints update the entity fields `leashed_by` and `rescue_state` and emit the corresponding engine events.

## Entity Layout

| Species | Count | Faction(s) | Notable | AI Archetype |
|---------|-------|------------|---------|-------------|
| human | ~40 | civilian, thief, bandit, merchant, ranger, cultist, werewolf, undead | innkeepers, guards, priests, thieves, bards, merchants, fisherman, farmers, miners, herbalists, children, patrons, rangers, cultists, werewolf, necromancer, bandit chief | passive, scripted |
| ... (rest of original table unchanged) |

### Added Fields to Entity JSON
- `leashed_by` – ID of the entity dragging this one (empty if none).
- `rescue_state` – "none", "in_progress", or "completed".

## Web API (`internal/api/entities.go`)
- `entityToDetail` now includes the new `leashed_by` and `rescue_state` fields.

## Web UI Adjustments
- Entity detail page can now display drag/rescue status and provide action buttons (future work).

*All documentation reflects the renamed file name `aGENTS.md`.*
