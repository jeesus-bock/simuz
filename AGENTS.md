# Simuz

## Meta

| Field | Value |
|-------|-------|
| LastUpdated | 2026-07-31 |
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
- Half-species (half-orc, half-elf, half-dwarf, half-goblin, half-hobgoblin, half-gnoll, half-kobold, half-fey) are crossbreed offspring of compatible parent species. Crossbreeding is rarer (0.05%/tick) than same-species reproduction (0.1%/tick). Half-species are fertile (CanReproduce=true) and form mate bonds.

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
- `CanReproduce` returns true for species that can breed naturally (human, orc, elf, goblin, fey, rat_king, kobold, vampire, hag, ogre, giant, bugbear, hobgoblin, lizardfolk, gnoll, troll, hydra, basilisk, cockatrice, manticore, skeleton, zombie, ghoul, half_orc, half_elf, half_dwarf, half_goblin, half_hobgoblin, half_gnoll, half_kobold, half_fey).
- `IsCavemanSpecies` returns true for species that reproduce without forming mate bonds (orc, ogre, giant, troll, cyclops, hydra, basilisk, cockatrice, manticore, skeleton, zombie, ghoul, medusa, floating_eye, beholder).
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

## Bestial Species & Generators

The `internal/species/registry.go` contains all registered species. Bestial/monster species include:

- **Bugbear** (`bugbear`): forests, highlands, waste — `IsCaveman: true`, nocturnal
- **Hobgoblin** (`hobgoblin`): highlands, plains, waste — militaristic, diurnal
- **Lizardfolk** (`lizardfolk`): swamps — amphibious, defensive
- **Gnoll** (`gnoll`): plains, waste — nomadic, nocturnal
- **Troll** (`troll`): forests, swamps — regenerative, no sleep cycle
- **Ogre** (`ogre`): highlands, waste — gluttonous, diurnal
- **Ettin** (`ettin`): highlands — two-headed, diurnal
- **Cyclops** (`cyclops`): mountains, waste — solitary, diurnal
- **Medusa** (`medusa`): caves, waste — immortal, petrifying gaze
- **Griffin** (`griffin`): highlands, mountains — diurnal, can fly
- **Wyvern** (`wyvern`): waste, highlands — nocturnal, venomous sting
- **Hydra** (`hydra`): swamps — multi-headed, regenerating, no sleep
- **Basilisk** (`basilisk`): waste, highlands — petrifying gaze, nocturnal
- **Cockatrice** (`cockatrice`): plains, forests — petrifying beak, nocturnal
- **Manticore** (`manticore`): forests, waste — venomous spikes, nocturnal
- **Skeleton** (`skeleton`): any raised location — undead, no CON
- **Zombie** (`zombie`): any raised location — undead, no CON
- **Ghoul** (`ghoul`): underground, wastes — undead, nocturnal
- **Lich** (`lich`): towers, dungeons — immortal undead, no sleep
- **Wraith** (`wraith`): shadow, underground — incorporeal undead
- **Mind Flayer** (`mind_flayer`): underdark — aberration, psionic
- **Beholder** (`beholder`): caves, dungeons — aberration, eye rays
- **Floating Eye** (`floating_eye`): swamps, caves — aberration, scouting

Dedicated generator functions in `internal/gen/world.go`:
- `generateHydras()` — swamps and lakes
- `generateBasilisks()` — deserts and rocky wastes, with petrifying lair locations
- `generateCockatrices()` — plains and forest farms, with nests
- `generateManticores()` — forests and wastes, with den locations
- `generateGriffins()` — highlands, with aeries
- `generateWyverns()` — cliffs and mountains, with perches
- `generateUndead()` — graveyard ruins, spawning skeletons/zombies/ghouls/lich/wraiths
- `generateAberrations()` — underdark lairs, spawning mind flayers/beholders/floating eyes
- `generateTrolls()` — forests and swamps
- `generateOgres()` — highlands
- `generateEttins()` — highlands
- `generateCyclopses()` — mountains and wastes
- `generateMedusas()` — caves and wastelands
