# Simuz

## Meta

| Field | Value |
|-------|-------|
| LastUpdated | 2026-07-27 |
| Status | active |

## Scope

Simuz is a Go-based world simulation game with:
- A 1 Hz tick loop for AI, combat, quests, weather, travel, and persistence.
- Lua-driven scripted AI in `internal/ai/scripts/*.lua`, loaded at startup.
- A Gin + HTMX + SSE web UI that refreshes on each tick.

## Current Rules

- Keep entity lists deterministic: sort by alive, knocked out, dead, then name, then ID.
- Keep location, quest, inventory, effect, combat zone, and traveler lists deterministically ordered by stable keys.
- Quests are UI-sorted deterministically; active quest order follows `AcceptedTick`.
- Traveler routes are tracked in travel state and should be shown in the UI when relevant.
- The locations page keeps its tree/map view selection in browser storage so SSE swaps do not reset it.
- Use markers, badges, or explicit labels for active-state UI instead of letting ordering imply state.

## Lua / World API

The runtime exposes `self`, `world`, and `util` tables.

Notable runtime capabilities:
- Movement and travel: `world.move_to`, `world.is_traveling`, `world.travel_exits`, `world.parent_location`
- Entities and combat: `world.entity_info`, `world.entities_at`, `world.nearby_entities`, `world.attack`, `world.heal`
- Social/quest hooks: `world.talk_to`, `world.give_quest`, `world.quest_progress`, `world.quest_set`
- Rescue/travel leash support: `world.drag_entity`, `world.undrag_entity`, `world.is_leashed`, `world.start_rescue`, `world.complete_rescue`
- Items and economy: `world.add_item`, `world.use_item`, `world.try_buy`, `world.try_sell`, `world.craft`
- Utility: `util.log`, `util.mem_set`, `util.mem_get`, `util.json_encode`, `util.json_decode`, `util.set_mood`

If you need exact signatures or enum values, check the code instead of expanding this file.

## Systems Worth Preserving

- Travel can be multi-tick across regions and should preserve route data.
- Combat detail pages show active/downed grouping and event logs.
- Quest state is persisted and loaded from SQLite.
- Entity detail pages show travel routes, active effects, moods, and quests.
- The map view is SVG-based and should stay stable across refreshes.
- Gender and natural reproduction: adult male/female pairs of the same species at the same location can produce offspring each tick (2% chance per group per tick). Children are level 1, inherit averaged attributes with small random variation, and are assigned a random gender.

## Editing Guidance

- Prefer `rg` for search and `apply_patch` for edits.
- Avoid destructive git commands.
- Do not add large static tables or full catalog dumps here; keep that data in code or generated docs.
- Remove outdated or duplicated guidance instead of layering more on top.
