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

Relationship accessors (on `self` and `world`):
- `self.get_relationship(other_id)`, `self.get_relationships()`, `self.get_children()`, `self.get_parents()`, `self.get_partner()`
- `self.get_relationship_type(other_id)`, `self.get_relationship_since(other_id)`, `self.has_relationship(other_id)`, `self.has_relationship_type(other_id, type)`, `self.is_related(other_id)`
- `self.add_relationship(other_id, type, tick)`, `self.remove_relationship(other_id)`, `self.num_relationships()`
- `world.get_relationship(entity_id, other_id)`, `world.get_children(entity_id)`, `world.get_parents(entity_id)`, `world.get_partner(entity_id)`
- `world.get_relationship_type(entity_id, other_id)`, `world.get_relationship_since(entity_id, other_id)`, `world.has_relationship(entity_id, other_id)`, `world.has_relationship_type(entity_id, other_id, type)`, `world.is_related(entity_id, other_id)`
- `world.add_relationship(entity_id, other_id, type, tick)`, `world.remove_relationship(entity_id, other_id)`, `world.num_relationships(entity_id)`

If you need exact signatures or enum values, check the code instead of expanding this file.

## Systems Worth Preserving

- Travel can be multi-tick across regions and should preserve route data.
- Combat detail pages show active/downed grouping and event logs.
- Quest state is persisted and loaded from SQLite.
- Entity detail pages show travel routes, active effects, moods, and quests.
- The map view is SVG-based and should stay stable across refreshes.

## Entity System

- `Entity` has a `Gender` field (`"male"`, `"female"`, `"other"`) and a `Pregnant` flag.
- Gender constants are defined in `internal/entity/entity.go`: `GenderMale`, `GenderFemale`, `GenderOther`.
- `NewEntity` assigns a random default gender (`male` or `female`) so spawned entities can mate.
- `Entity` has a `LastReproductionTick` field for per-entity reproduction cooldown.
- `IsAdult` returns true when entity level >= 3 (reproductive age).
- `CanReproduce` returns true for species that can breed naturally (human, orc, elf, goblin, fey, rat_king, kobold, vampire, hag, ogre, giant).
- `IsCavemanSpecies` returns true for species that reproduce without forming mate bonds (orc, ogre, giant, troll, cyclops).
- `Profession` is a string field on `Entity` for occupational identity (e.g., "bandit", "merchant", "farmer").
- `Faction` on `Entity` is now narrow: it represents voluntary group membership only (cults, religions, political movements). Use `FactionCivilian`, `FactionCult`, `FactionDeity` constants. Species-based factions (orc, bandit, etc.) are stored in `Profession` or `FactionID` on the AI struct instead.
- Reproduction logic (`processReproduction`) lives in `internal/engine/tick.go`.
- Natural reproduction: adult male/female pairs of the same species at the same location have a 0.1% chance per tick to produce offspring.
- Caveman species (orc, ogre, giant, etc.) have a 0.3% chance per tick and do not form mate bonds.
- Reproduction is gated by:
  - Global entity cap of 5000
  - Per-location per-species population cap of 20
  - Per-entity cooldown of 1000 ticks (~16 minutes)
- `averageAttrs` computes child attributes as parent average with small random variance.
- `ProcessPregnancy` in `internal/engine/spawning.go` handles gestation completion.
- Pregnancy gestation periods are per-species via `SpeciesGestationTicks`; falls back to 200 ticks.
- Immortal/undead species (deity, vampire) cannot reproduce.
- All spawned entities (seeded, natural birth, respawn) receive randomized XP appropriate for their level via `randomXPForLevel`.

## Divine Realms

- `IsDivineRealm` in `internal/world/location.go` checks if a location or any ancestor has the `divine_realm` tag.
- Mortals cannot move TO divine realms, but can escape FROM them.
- Mortals cannot travel through divine realms mid-route; such travel is blocked.
- If a mortal somehow ends up in a divine realm, they can move to any mortal location.

## Editing Guidance

- Prefer `rg` for search and `apply_patch` for edits.
- Avoid destructive git commands.
- Do not add large static tables or full catalog dumps here; keep that data in code or generated docs.
- Remove outdated or duplicated guidance instead of layering more on top.
