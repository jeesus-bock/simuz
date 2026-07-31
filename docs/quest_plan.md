# Quest System Enhancement Plan

## Overview

This document outlines proposed enhancements, bug fixes, and known issues for the quest system based on the current codebase (`internal/quest/`).

---

## Known Issues

### 1. Prerequisites Are Not Fully Validated
**File:** `internal/quest/quest.go` — `canAccept()` / `CanAccept()`

The `Prerequisites` struct supports `QuestsCompleted`, `QuestsActive`, `FactionRep`, and `Flags`, but `canAccept()` only checks `LevelMin` and `LevelMax`. The remaining prerequisite fields are parsed from Lua but never evaluated. This means quests with faction or flag requirements can be accepted without meeting those conditions.

**Fix:** Implement checks for all prerequisite fields in `canAccept()` and `CanAccept()`.

### 2. Fail Conditions Are Parsed but Never Checked
**File:** `internal/quest/loader.go` — `tableToQuestDef()`

Fail conditions (`failure_conditions`) are parsed into `QuestDef.FailConditions` but there is no logic anywhere in `quest.go` or `loader.go` that evaluates them. Time-based (`Hours`), entity-based (`EntityID`), and flag-based (`Flag`) fail conditions are all inert.

**Fix:** Add a `CheckFailConditions(entityID, questID)` method in `Manager` and call it periodically (e.g., on each tick or when relevant state changes).

### 3. Optional Objectives Are Not Distinguished
**File:** `internal/quest/quest.go` — `checkStageCompletion()`

`ObjectiveDef` has an `Optional` field, but `checkStageCompletion()` treats all objectives the same — it requires `progress >= count` for every objective. Optional objectives should not block stage completion.

**Fix:** Skip optional objectives when evaluating whether all objectives in a stage are satisfied.

### 4. Quest Dialog Strings Are Unused
**File:** `internal/quest/definition.go` — `QuestDialog` struct

`QuestDialog` defines `Accept`, `Progress`, and `Complete` message templates, but no code in the loader or manager references these fields. NPCs and the UI have no way to display context-appropriate dialog when a quest is offered, progressed, or completed.

**Fix:** Wire `QuestDialog` into the quest acceptance, progression, and completion flows so callers can retrieve the appropriate message.

### 5. `luaInt` Does Not Handle String Numbers
**File:** `internal/quest/loader.go` — `luaInt()`

`luaInt()` only handles `lua.LNumber`. If a Lua script passes a numeric string (e.g., `count = "5"`), it silently returns 0. This is unlikely with the current scripts but is fragile.

**Fix:** Add a `luaStringToFloat` or `luaNumber` helper that coerces `lua.LString` values containing numbers.

### 6. No Cycle Detection in Stage Requirements
**File:** `internal/quest/loader.go` — `tableToQuestDef()`

Stage `Requirements` reference other stage IDs, but there is no validation that the dependency graph is acyclic. A circular dependency would cause `checkStageCompletion()` to never complete the stage.

**Fix:** Add a topological sort or DFS cycle check when loading quest definitions.

### 7. `CheckCollectItem` Does Not Validate Entity Ownership
**File:** `internal/quest/quest.go` — `CheckCollectItem()`

`CheckCollectItem` iterates over all active quests for the entity and increments matching objectives. However, it does not verify that the item being collected actually belongs to the entity or was obtained legitimately. If the item system calls this on any item interaction, it could lead to unintended quest progression.

**Fix:** Ensure `CheckCollectItem` is called only when the entity legitimately acquires the item, or add ownership validation.

---

## Proposed Enhancements

### 1. Quest Difficulty Scaling
**Target:** `internal/quest/definition.go` — `QuestDef`

Add a `ScaleWithLevel bool` field and a `LevelScale` multiplier so that rewards and objective counts adjust based on the entity's level relative to the quest level. This would make side quests more rewarding for over-leveled players and more accessible for under-leveled ones.

### 2. Quest Groups and Categories
**Target:** `internal/quest/definition.go` — `QuestDef`

Add a `Group` or `Category` field (e.g., "frosthold", "crystal_forest", "faction_guard") to enable UI grouping, filtered quest boards, and region-specific quest tracking.

### 3. Time-Limited Quests via Fail Conditions
**Target:** `internal/quest/quest.go` — `Manager`

Implement the `Hours` fail condition so that quests expire if not completed within a set timeframe. This would enable daily quests and timed challenges. The `Manager` would need to track `AcceptedTick` (already present) and compare against the current tick on each check.

### 4. Quest Reward Preview
**Target:** `internal/quest/quest.go` — `Manager`

Add a `PreviewRewards(entityID, questID)` method that returns the expected rewards without accepting the quest. This is useful for UI quest boards where the player can hover or inspect a quest before committing.

### 5. Shared / Global Objectives
**Target:** `internal/quest/quest.go` — `Manager`

Support a new objective type `global_kill` or `global_collect` that counts toward the objective for all active quests that reference it. For example, killing rats could progress both "The Rat Problem" and any other rat-related quest simultaneously.

### 6. Quest Persistence and Save/Load
**Target:** `internal/quest/quest.go` — `Manager`

Add `SaveState(entityID)` and `LoadState(entityID)` methods that serialize/deserialize the full `EntityQuestState` map to/from the `Storage` interface (already available on `Simulation`). Currently `LoadState` exists but only loads a single quest; it should support bulk loading.

### 7. Event-Driven Quest Hooks
**Target:** `internal/quest/quest.go` — `Manager`

Expand the objective types to include `on_kill`, `on_collect`, `on_level_up`, `on_faction_rep`, and `on_flag` so that quests can react to game events beyond the current `visit_location`, `talk_to_npc`, `deliver_item`, `kill_entities`, and `collect_items`.

### 8. Quest Journal Export
**Target:** `internal/quest/quest.go` — `Manager`

Add a `JournalData(entityID)` method that returns a structured summary of all quests (active, completed, available) suitable for rendering in a quest journal UI. Include stage names, objective descriptions, progress, and rewards.

---

## Quest Script Audit

The following Lua scripts are present and should be reviewed for consistency:

| Script | Quest ID | Type | Level | Stages | Notes |
|---|---|---|---|---|---|
| `rat_problem.lua` | `rat_problem` | side | 1 | 2 | Baseline working quest |
| `deliver_sword.lua` | `deliver_sword` | side | 1 | 2 | Tests deliver_item objective |
| `freya_favor.lua` | `freya_favor` | side | 3 | 2 | Uses collect_items with location_id |
| `deity_whispers.lua` | `deity_whispers` | main | 5 | 2 | Main quest example |
| `taken_courier.lua` | `taken_courier` | side | 3 | 3 | Multi-stage with kill + talk |
| `vampire_hunt.lua` | `vampire_hunt` | side | 4 | 3 | Three-stage chain |
| `hag_curse.lua` | `hag_curse` | side | 3 | 3 | Three-stage chain |
| `bard_ballad.lua` | `bard_ballad` | side | 1 | 2 | Simple visit + talk |
| `zeus_crazy_task.lua` | `zeus_crazy_task` | main | 1 | 3 | Main quest, low level |
| `fairy_escort.lua` | `fairy_escort` | side | 2 | 2 | Collect + deliver |
| `kobold_menace.lua` | `kobold_menace` | side | 2 | 3 | Three-stage chain |
| `lost_heirlooms.lua` | `lost_heirlooms` | side | 2 | 2 | Collect in specific location |

**Total:** 12 quests (matches the test expectation in `loader_test.go`).

### Observations from Scripts
- All quests use `quest.define()` correctly.
- `rat_problem` and `lost_heirlooms` use `collect_items` with a `location_id`, which is the only quests that combine item collection with a location constraint.
- No quests use `failure_conditions`, `prerequisites`, or `unlocks` — these features are untested by the Lua scripts.
- `zeus_crazy_task` is a main quest at level 1, which may be too easy for a main quest.
- No repeatable or daily quests exist yet despite the `QuestType` constants supporting them.

---

## Test Coverage Gaps

- `loader_test.go` validates structure but does not test quest progression (accept → stage completion → reward).
- No tests for fail conditions, prerequisites, or optional objectives.
- No tests for `RecordActivity` or `RecentActivity`.
- No tests for quest unlock chaining (`triggerUnlocks`).

---

## Priority Recommendations

1. **High:** Implement prerequisite validation (issue #1) — currently broken.
2. **High:** Implement fail condition checking (issue #2) — dead code.
3. **Medium:** Handle optional objectives (issue #3) — incorrect stage completion logic.
4. **Medium:** Wire quest dialogs (issue #4) — unused feature.
5. **Low:** Add cycle detection (issue #6) — defensive improvement.
6. **Low:** Add quest journal export (enhancement #8) — UI-facing feature.
