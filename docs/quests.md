# Quest System

## Overview

Quests are defined in YAML files and loaded at runtime. The quest engine evaluates prerequisites, tracks objective progress, processes triggers, and awards rewards. This separates quest logic (declarative YAML) from engine code (Go), making quests content-editable without recompilation.

## Quest Lifecycle

```
INACTIVE ──→ ACTIVE ──→ STAGE_N ──→ ... ──→ COMPLETED
                 │                              │
                 │                              ├→ REPEATABLE (reset to ACTIVE)
                 │                              └→ LOCKED (permanent)
                 │
                 └──→ FAILED
                        │
                        ├→ RETRYABLE → ACTIVE
                        └→ PERMANENT (cannot retry)
```

| State | Meaning |
|-------|---------|
| INACTIVE | Not yet available (prerequisites unmet) |
| ACTIVE | Accepted, current stage in progress |
| STAGE_N | Active on stage N |
| COMPLETED | All stages done, rewards granted |
| FAILED | A failure condition was met |
| REPEATABLE | Completed, can accept again (daily/weekly) |

## YAML Quest Format

### Structure

```yaml
id: str                        # Unique quest identifier
title: str                     # Display name
type: main | side | faction | daily | repeatable
level: int                     # Recommended level (for UI display)

source:                        # How the quest is obtained
  type: npc | discovery | auto | item
  npc_id: str                  # If type == npc
  location_id: str             # Where the giver is found
  dialog:                      # If type == npc
    accept: str                # Dialog on first talk
    progress: str              # Dialog while active
    complete: str              # Dialog on turn-in

description: str               # Narrative text for journal

prerequisites:                 # Conditions to accept the quest
  quests_completed: [str]
  quests_active: [str]
  level_min: int
  level_max: int
  faction_reputation:          # {faction_id: min_reputation}
    faction_id: int
  flags: [condition]           # Arbitrary flag conditions
  and: [condition]             # All must pass
  or: [condition]              # Any must pass
  not: condition               # Must NOT pass

stages:
  - id: str                    # Stage identifier
    name: str                  # Stage display name
    description: str           # Journal text for this stage
    requirements: [str]        # Stage IDs that must be completed first

    objectives:
      - id: str                # Objective identifier
        type: str              # See Objective Types table
        description: str
        optional: bool         # true = bonus objective
        # Type-specific fields follow (see per-type docs)
        count: int
        entity_template: str
        location_id: str
        npc_id: str
        item_template: str
        recipe: str
        duration_hours: float

    on_enter: [trigger_action]       # Run when stage becomes active
    on_complete: [trigger_action]    # Run when all objectives done

rewards:
  experience: int
  gold: int
  items:
    - template: str
      count: int
  faction_reputation:
    faction_id: int
  unlocks:
    quests: [str]
    locations: [str]
    recipes: [str]
    abilities: [str]

failure_conditions:
  - type: time_limit | entity_dead | flag_set | objective_failed
    # Type-specific fields
    hours: int                 # For time_limit
    entity_id: str             # For entity_dead
    flag: str                  # For flag_set
    objective: str             # For objective_failed
    on_fail: [trigger_action]

triggers:                      # Side effects during the quest
  - on: stage_enter | stage_complete | objective_done | quest_complete | quest_failed
    stage: str                 # Filter: only fire for this stage
    objective: str             # Filter: only fire for this objective
    action: trigger_action     # What to do

branching:                     # Alternate outcomes based on conditions
  - condition: condition       # Evaluated on completion
    dialog_override: str       # Replace the complete dialog
    extra_rewards:             # Additional rewards
      gold: int
      items: [{template, count}]
```

### Condition Format

Conditions are used in `prerequisites`, `failure_conditions`, and `branching.condition`.

```yaml
# Simple condition
type: condition_type
# Type-specific params

# Compound conditions
and:
  - condition1
  - condition2

or:
  - condition1
  - condition2

not:
  condition
```

### Full Example

```yaml
id: "rat_problem"
title: "The Rat Problem"
type: side
level: 2

source:
  type: npc
  npc_id: "innkeeper_greta"
  location_id: "frosthold_inn"
  dialog:
    accept: "Rats in my cellar! Clear them out and I'll reward you."
    progress: "Still hear them scratching down there..."
    complete: "Cellar's clean! Here's your reward."

description: >
  Innkeeper Greta at Frosthold has a rat infestation in her cellar.
  She's offering a reward to anyone brave enough to clear them out.
  Something about a giant rat king rumored to be down there...

prerequisites:
  quests_completed: ["frosthold_arrival"]
  level_min: 1

stages:
  - id: investigate
    name: "Investigate the Cellar"
    description: "Head down to the inn's cellar and see what you're dealing with."
    requirements: []

    objectives:
      - id: enter_cellar
        type: visit_location
        location_id: "frosthold_inn_cellar"
        description: "Enter the cellar"
        count: 1

    on_enter:
      - spawn_entities:
          template: "giant_rat"
          count: 6
          location_id: "frosthold_inn_cellar"

  - id: kill_rats
    name: "Cull the Vermin"
    description: "Kill 8 giant rats. Keep an eye out for the rat king."
    requirements: ["investigate"]

    objectives:
      - id: kill_rats_main
        type: kill_entities
        entity_template: "giant_rat"
        count: 8
        description: "Kill 8 giant rats"

      - id: kill_rat_king
        type: kill_entities
        entity_template: "rat_king"
        count: 1
        description: "Slay the Rat King (optional)"
        optional: true

    fail_conditions:
      - type: entity_dead
        entity_id: "innkeeper_greta"
        on_fail:
          - set_flag: "rat_problem_failed_greta_dead"

  - id: report
    name: "Report to Greta"
    description: "Return to the inn and tell Greta the cellar is safe."
    requirements: ["kill_rats"]

    objectives:
      - id: talk_greta
        type: talk_to_npc
        npc_id: "innkeeper_greta"
        dialog_node: "quest_complete_rats"
        description: "Talk to Greta about the cellar"

rewards:
  experience: 150
  gold: 50
  items:
    - template: "health_potion"
      count: 2
    - template: "rusty_sword"
      count: 1
  faction_reputation:
    frosthold: 15
  unlocks:
    quests: ["sewer_expedition"]
    locations: ["frosthold_sewer_entrance"]

failure_conditions:
  - type: time_limit
    hours: 72
    on_fail:
      - set_flag: "rat_problem_timed_out"
      - faction_reputation:
          frosthold: -10

triggers:
  - on: objective_done
    objective: "kill_rat_king"
    action: give_item "rat_king_fang" 1

  - on: quest_complete
    action: broadcast "Greta's cellar is safe again!"

branching:
  - condition:
      has_item: "rat_king_fang"
    dialog_override: "You got the Rat King too? Take this extra reward!"
    extra_rewards:
      gold: 100
      items:
        - template: "silver_ring"
          count: 1
```

---

## Objective Types

### kill_entities

Track kills of a specific entity template.

```yaml
type: kill_entities
entity_template: "giant_rat"     # Entity template ID
count: 8                         # Number to kill
```

**Progress:** Incremented each tick when an entity of matching template dies within visibility range of the quest holder.

---

### visit_location

Track entry into a specific location.

```yaml
type: visit_location
location_id: "frosthold_inn_cellar"
count: 1
```

**Progress:** Set to `count` when the entity enters the location.

---

### talk_to_npc

Track a conversation with an NPC.

```yaml
type: talk_to_npc
npc_id: "innkeeper_greta"
dialog_node: "quest_complete_rats"     # Required dialog node
```

**Progress:** Set to 1 when the entity completes the specified dialog node with the NPC.

---

### collect_items

Track collection of items into inventory.

```yaml
type: collect_items
item_template: "rat_king_fang"
count: 3
```

**Progress:** Incremented on item pickup. Decremented if item is dropped/sold.

---

### deliver_item

Track delivery of an item to an NPC.

```yaml
type: deliver_item
item_template: "rat_king_fang"
target_npc_id: "innkeeper_greta"
count: 1
```

**Progress:** Set to `count` when the entity transfers the item to the NPC via dialog or trade.

---

### travel_to

Track arrival at a distant location (via travel system).

```yaml
type: travel_to
destination_id: "frosthold_sewer_entrance"
```

**Progress:** Set to 1 when entity's travel state resolves and they arrive at the destination.

---

### craft_items

Track crafting of specific items.

```yaml
type: craft_items
recipe: "iron_sword"
count: 1
```

**Progress:** Incremented on successful craft completion.

---

### survive_time

Track survival for a duration.

```yaml
type: survive_time
duration_hours: 24          # In-game hours
```

**Progress:** Ticks up while the entity is alive and the quest is active. Set to `count` when duration elapses.

---

### escort_npc

Track escort of an NPC to a destination.

```yaml
type: escort_npc
npc_id: "merchant_lev"
destination_id: "frosthold_gate"
```

**Progress:** Set to 1 when the escorted NPC arrives at the destination. Fails if NPC dies.

---

### use_item_at

Track use of a specific item at a specific location.

```yaml
type: use_item_at
item_template: "cellar_key"
location_id: "frosthold_cellar_door"
```

**Progress:** Set to 1 on successful use.

---

### custom

Arbitrary progress determined by a Go function.

```yaml
type: custom
script: "check_sewer_clearance"    # Maps to a registered Go function
params:
  min_clear_radius: 10
```

**Progress:** Evaluated by calling the registered `func(entity, params) bool` each tick.

---

## Condition Types

| Condition | Fields | Evaluates |
|-----------|--------|-----------|
| `level_ge` | `value: int` | Entity level >= value |
| `level_le` | `value: int` | Entity level <= value |
| `quest_completed` | `quest_id: str` | Quest is in COMPLETED state |
| `quest_active` | `quest_id: str` | Quest is in ACTIVE state |
| `quest_not_started` | `quest_id: str` | Quest is INACTIVE |
| `has_item` | `template: str`, `count: int` | Entity has N of item |
| `entity_alive` | `entity_id: str` | Entity is alive |
| `entity_dead` | `entity_id: str` | Entity is dead |
| `flag_set` | `flag: str` | Flag exists and is truthy |
| `flag_unset` | `flag: str` | Flag does not exist or is falsy |
| `location_visited` | `location_id: str` | Entity has visited location |
| `faction_reputation` | `faction_id: str`, `value: int` | Reputation >= value |
| `time_of_day` | `phase: dawn/day/dusk/night` | World time matches |
| `season` | `season: spring/summer/autumn/winter` | World season matches |
| `weather` | `type: str` | Current weather type |
| `skill_level` | `skill: str`, `value: int` | Skill level >= value |
| `has_flag` | `flag: str` | Entity has a specific flag |
| `objective_done` | `quest_id: str`, `objective_id: str` | Objective completed |
| `stage_active` | `quest_id: str`, `stage_id: str` | Entity on this stage |
| `and` | `conditions: [condition]` | All sub-conditions true |
| `or` | `conditions: [condition]` | Any sub-condition true |
| `not` | `condition: condition` | Sub-condition false |

---

## Trigger Actions

Actions run in response to triggers (`on_enter`, `on_complete`, `triggers`, `failure_conditions.on_fail`).

| Action | Params | Effect |
|--------|--------|--------|
| `spawn_entities` | `template, count, location_id` | Spawn entity instances |
| `give_item` | `template, count` | Add item to entity inventory |
| `remove_item` | `template, count` | Remove item from entity inventory |
| `set_flag` | `flag` | Set a world/entity flag |
| `clear_flag` | `flag` | Remove a flag |
| `faction_reputation` | `faction_id, value` | Modify faction standing |
| `broadcast` | `message` | Send a message to nearby entities |
| `dialog` | `npc_id, node` | Force a dialog interaction |
| `teleport` | `location_id` | Move entity to location |
| `advance_stage` | `quest_id, stage_id` | Force stage advancement |
| `fail_quest` | `quest_id` | Fail the quest immediately |
| `complete_objective` | `objective_id` | Force objective completion |
| `run_script` | `script, params` | Call a registered Go function |

---

## Runtime Engine

```
internal/quest/
├── definition.go           # YAML data structures (Quest, Stage, Objective, etc.)
├── engine.go               # QuestManager — central controller
├── conditions.go           # Condition evaluator (condition type → func)
├── objectives.go           # Objective progress tracker (objective type → func)
├── triggers.go             # Trigger dispatcher
├── store.go                # Quest state persistence
└── quest.example.yaml      # Reference example
```

### QuestManager API

```
type QuestManager struct {}

func NewQuestManager(store QuestStore) *QuestManager

// Load quest definitions from YAML files in a directory
func (qm *QuestManager) LoadQuests(dir string) error

// Check if an entity can accept a quest
func (qm *QuestManager) CanAccept(entity *Entity, questID string) bool

// Accept a quest, setting it to ACTIVE on stage 0
func (qm *QuestManager) AcceptQuest(entity *Entity, questID string) error

// Get all quest states for an entity
func (qm *QuestManager) GetStates(entityID string) map[string]QuestState

// Process an event that might advance objectives (called by tick loop)
func (qm *QuestManager) ProcessEvent(event QuestEvent) error

// Evaluate conditions against entity + world state
func (qm *QuestManager) EvaluateConditions(entity *Entity, cond Condition) bool
```

### Event-Driven Progress

The tick loop calls `ProcessEvent` for events that may affect quests:

```go
type QuestEvent struct {
    Type   QuestEventType
    Source string    // entity or location ID
    Data   map[string]interface{}
}

// Event types
const (
    EventEntityKilled       // Data: {killer_id, victim_template, victim_id}
    EventEntityTalked       // Data: {entity_id, npc_id, dialog_node}
    EventLocationEntered    // Data: {entity_id, location_id}
    EventItemCollected      // Data: {entity_id, item_template, count}
    EventItemDelivered      // Data: {entity_id, item_template, target_npc_id}
    EventItemUsed           // Data: {entity_id, item_template, location_id}
    EventCraftCompleted     // Data: {entity_id, recipe}
    EventTravelCompleted    // Data: {entity_id, destination_id}
    EventTick               // Data: {tick_number}
    EventTimePassed         // Data: {hours}
    EventCustom             // Data: {script, ...}
)
```

### Tick Loop Integration

```go
func (tl *TickLoop) Tick() {
    tl.tick++

    // Process scheduled events
    tl.scheduler.ProcessDue(tl.tick)

    // Update world
    tl.world.Tick()

    // Update entities
    for _, entity := range tl.entities.All() {
        entity.Tick()

        // Check active quest failure conditions
        tl.questManager.CheckFailures(entity)
    }

    // Process quest time limits (global check)
    tl.questManager.ProcessTick(tl.tick)

    // Combat
    tl.combat.Tick()

    // Persistence
    if tl.tick%saveInterval == 0 {
        tl.storage.Save()
    }
}
```

### State Persistence

```yaml
# Stored per entity in the database
entity_quests:
  "rat_problem":
    state: ACTIVE                        # INACTIVE / ACTIVE / COMPLETED / FAILED
    current_stage: "kill_rats"
    completed_stages: ["investigate"]
    objectives:
      "kill_rats_main": 5                # 5/8 giant rats killed
      "kill_rat_king": 0
    variables:                            # Quest-scoped variables
      rats_spawned: true
    accepted_tick: 14235
    started_at: "day 3, 14:30"
```

---

## Quest File Loading

YAML quest files live in a `quests/` directory. On startup:

```
quests/
├── frosthold_arrival.yaml
├── rat_problem.yaml
├── sewer_expedition.yaml
├── guild_quests/
│   ├── hunters_league_01.yaml
│   └── merchants_guild_01.yaml
└── faction_quests/
    ├── frosthold_militia.yaml
    └── underdark_trade.yaml
```

Loading:

```go
// Load all quests from directory tree
err := questManager.LoadQuests("quests/")
```

Quests are indexed by ID. Duplicate IDs cause a startup error.

---

## Validation

On load, each quest YAML is validated:

- Required fields: `id`, `title`, `type`, `stages`
- Each stage has at least one objective
- `requirements` references only existing stage IDs in the same quest
- Objective type matches a registered handler
- Condition types match registered evaluators
- Trigger action types match registered actions
- Source NPC exists in the entity templates
- Items/rewards reference existing item templates

Validation errors are fatal at startup to catch issues before runtime.
