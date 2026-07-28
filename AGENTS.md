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

## Lua Scripting

Full API reference for Lua scripting is in `docs/lua-scripting.md`.

Key points:
- Scripts are loaded from `internal/ai/scripts/` and execute once per tick.
- Each script defines a `do_tick()` function as the main entry point.
- Scripts have access to `self` (the entity), `world` (game state & actions), and `util` (helpers).
- Scripts can return up to three values: `didAct`, log messages, and event tables.
- The `events` package (`internal/events/engine.go`) defines `SimEvent` for scripted AI actions.

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
- `Profession` is a string field on `Entity` for occupational identity (e.g., "bandit", "merchant", "farmer", "ranger", "bard", "fisherman"). It can be empty for entities with no specific profession.
- `Faction` on `Entity` is narrow: it represents voluntary group membership only (cults, religions, political movements). Use `FactionCivilian`, `FactionCult`, `FactionDeity` constants. Species-based identities and occupational roles are stored in `Profession` or `FactionID` on the AI struct instead.
- Reproduction logic (`processReproduction`) lives in `internal/engine/tick.go`.
- Natural reproduction: adult male/female pairs of the same species at the same location have a small chance to produce offspring, gated by cooldown and population caps.
- Caveman species (orc, ogre, giant, etc.) have a higher reproduction chance and do not form mate bonds.
- `ProcessPregnancy` in `internal/engine/spawning.go` handles gestation completion.
- Pregnancy gestation periods are per-species via `SpeciesGestationTicks`; falls back to 200 ticks.
- Immortal/undead species (deity, vampire) cannot reproduce.
- All spawned entities (seeded, natural birth, respawn) receive randomized XP appropriate for their level via `randomXPForLevel`.
- Profession is inherited from parents during natural reproduction and can be set on spawn rules for scripted entities.

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
