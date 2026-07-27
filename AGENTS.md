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
| `world.weather([id])` | `(table\|nil)` | Effective weather at loc (or self): type, temp, visibility, wind, humidity, harsh, stormy, vis_mod, travel_mod |
| `world.location_control([id])` | `(table\|nil)` | `{faction, strength}` controlling faction for location |
| `world.is_traveling()` | `(boolean)` | True if self is mid multi-tick cross-region travel |
| `world.parent_location(id)` | `(string, string\|nil)` | Get parent location ID for a location, or nil |
| `world.entities_at(id)` | `(string, table)` | List alive entity IDs at any location by ID |
| `world.move_to(id)` | `(boolean)` | Instant move within same region/city tree; multi-tick travel across regions |
| `world.nearby_entities()` | `(string, table)` | List alive entity IDs at `self.loc_id` (excludes self) |
| `world.entity_name(id)` | `(string, string\|nil)` | Get entity name for ID, or nil |
| `world.entity_items(id)` | `(string, table\|nil)` | List unequipped item DefIDs for entity, or nil |
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
| `world.drag_entity(target_id)` | `(boolean)` | Leash/drag target entity so it follows caller's movement |
| `world.undrag_entity(target_id)` | `(boolean)` | Release leash/drag on target entity |
| `world.is_leashed([target_id])` | `(boolean, string)` | Check if entity is leashed; returns (is_leashed, dragger_id) |
| `world.start_rescue(target_id)` | `(boolean)` | Mark target entity's rescue state as "in_progress" |
| `world.complete_rescue(target_id)` | `(boolean)` | Complete rescue of target entity and release leash |

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

## Divine Intervention System (`combat/simple.go` + `ai/scripted.go`)

Deities with `AI.Type == "scripted"` and `"deity"` in `AI.ScriptIDs` can perform divine interventions via `world.divine_intervention(deity_id, target_id, type, ...)`.

| Type | Signature | Description |
|------|-----------|-------------|
| `heal` | `divine_intervention(deityID, targetID, "heal")` | Heals target for `10 + deity.Level + (WIS/2)` HP |
| `bless` | `divine_intervention(deityID, targetID, "bless")` | Sets `blessed` flag on target for `30 + WIS` ticks; grants +2 attack skill and +1 DR while active |
| `smite` | `divine_intervention(deityID, targetID, "smite")` | Deals `5 + Level + STR/3` damage to ALL hostiles in target's location |
| `scare` | `divine_intervention(deityID, targetID, "scare")` | Deals `3 + Level/2` fear damage to all mortals in target's location |
| `quest` | `divine_intervention(deityID, targetID, "quest", questID)` | Grants quest `questID` to target entity |

## Faction Relations (`combat/simple.go`)

| Faction A | Faction B | Relation |
|-----------|-----------|----------|
| civilian | thief | Hostile |
| civilian | bandit | Hostile |
| civilian | vermin | Hostile |
| civilian | werewolf | Hostile |
| civilian | cultist | Hostile |
| guard | thief | Hostile |
| guard | bandit | Hostile |
| guard | vermin | Hostile |
| guard | werewolf | Hostile |
| guard | cultist | Hostile |
| merchant | thief | Hostile |
| merchant | bandit | Hostile |
| merchant | werewolf | Hostile |
| merchant | cultist | Hostile |
| orc | elf | Hostile |
| deity | orc | Hostile |
| deity | elf | Hostile |
| deity | beast | Hostile |
| deity | thief | Hostile |
| deity | bandit | Hostile |
| deity | vermin | Hostile |
| deity | goblin | Hostile |
| deity | kobold | Hostile |
| deity | undead | Hostile |
| deity | hag | Hostile |
| deity | werewolf | Hostile |
| deity | cultist | Hostile |
| deity | dragon | Hostile |
| kobold | civilian | Hostile |
| kobold | guard | Hostile |
| kobold | beast | Hostile |
| kobold | fey | Hostile |
| undead | civilian | Hostile |
| undead | guard | Hostile |
| undead | fey | Hostile |
| hag | civilian | Hostile |
| hag | guard | Hostile |
| hag | fey | Hostile |
| beast | ranger | Hostile |
| vermin | ranger | Hostile |
| werewolf | ranger | Hostile |
| cultist | ranger | Hostile |

All other pairs default to Neutral. Same faction is Friendly.

### Dynamic Faction Relations
- Relations stored in mutable `factionRelations` map (combat/simple.go)
- `SetRelation(a, b, rel)` — set relation between two factions
- `ShiftRelation(a, b, delta)` — shift by delta step (Friendly→Neutral→Hostile)
- `RelationsJSON()` / `LoadRelationsJSON()` — persisted to SQLite in `world_state.faction_relations_json`
- Lua API: `world.set_relation("factionA", "factionB", "hostile"|"friendly"|"neutral")`
- Killing a target shifts relation hostile between the factions (tick.go + scripted.go)
- All existing `combat.Relation()` call sites use the map automatically

## Entity Archetypes (`internal/ai/archetypes.go` + `internal/engine/tick.go`)

| Archetype | Go Handler | Description |
|-----------|------------|-------------|
| passive | `processEntityAI` | Follows daily routine, roams randomly, returns home at dusk |
| aggressive | `processEntityAI` | Rate-limited per-faction (5 + hash%5 ticks), attacks hostiles in same location |
| defensive | `processEntityAI` | Attacks hostiles nearby every 3 ticks, returns home when threat passes |
| hunting | `processEntityAI` | Attacks hostiles, chases prey to adjacent locations, returns home at night |
| gathering | `processEntityAI` | Avoids combat, flees from hostiles, stays at outdoor locations |
| healing | `processEntityAI` | Heals injured nearby non-hostile every 5 ticks, stays near home |
| scouting | `processEntityAI` | Explores child locations every 2 ticks, flees from hostiles |
| guarding | `processEntityAI` | Stays at home, attacks any hostile that approaches |
| scripted | Lua VM | Runs all `AI.ScriptIDs` each tick sequentially |
| dormant | none | Inactive until awakened |

## Web UI (`internal/web/`)

Gin-based UI with HTMX 2.0.1 SSE live updates. All pages auto-refresh on each tick via SSE event `"tick"`.

| Page | Route | Fragment | Description |
|------|-------|----------|-------------|
| Dashboard | `/` | `/api/v1/ui/fragments/dashboard` | Tick, time, entity count, location count |
| Entities | `/entities` | `/api/v1/ui/fragments/entities` | Table of all entities with name/species/level/HP/location/AI |
| Entity Detail | `/entity/:id` | `/api/v1/ui/fragments/entity/:id` | Full entity card: stats, attributes, equipment, inventory, activity, AI, recent combat events |
| Locations | `/locations` | `/api/v1/ui/fragments/locations` | Tree + map; weather temp, controlling faction |
| Location Detail | `/location/:id` | `/api/v1/ui/fragments/location/:id` | Weather, control, exits, entities, combat, travelers |
| Combat | `/combat` | `/api/v1/ui/fragments/combat` | Active combat zones (locations with 2+ factions) |
| Combat Detail | `/combat/:location` | `/api/v1/ui/fragments/combat/:location` | Combatants + scrollable event log per zone |
| Quests | `/quests` | `/api/v1/ui/fragments/quests` | Active quests and their progress |
| AI | `/ai` | `/api/v1/ui/fragments/ai` | AI archetypes and loaded scripts |

## Economy (`internal/economy/trade.go`)

| Function | Description |
|----------|-------------|
| SellerMarkup | 1.5x base price |
| BuyerDiscount | 0.6x base price |
| Haggling | Threshold = (buyer.INT+WIS) - (seller.CHA+INT); uses `EffectiveAttrs()` so substances affect prices |
| Currency | cp (1), sp (10), gp (100), tp (1000), mp (10000), ep (100000) |

## Quest System (`internal/quest/`)

- **Quest definitions are Lua scripts** in `internal/quest/scripts/*.lua`, loaded via `quest.LoadScripts()` / `gen.SeedQuests()`
- Each script calls `quest.define({ id, title, type, level, description, source, stages, rewards, ... })` once at load time
- Runtime engine remains Go (`Manager`): accept, progress, complete, fail, unlocks
- `QuestDef` objectives: `kill_entities`, `collect_items`, `visit_location`, `talk_to_npc`, `deliver_item`
- `ProgressObjective()` / `SetObjective()` auto-tracked
- Kill tracking via `questKilled()` in tick.go for active quests
- Visit tracking via `CheckVisitLocation()` in quest manager (called from Lua `move_to`)
- Item collection tracking via `CheckCollectItem()` in quest manager (called from Lua `add_item`)
- NPC interaction tracking via `CheckTalkToNPC()` and `CheckDeliverItem()` in quest manager
- Quest reward delivery: `OnQuestComplete` callback distributes XP, gold, items on completion
- Quest chain unlocking: `Rewards.Unlocks.Quests` unlocks new quests for the entity
- Time-limited quests: `FailCondition.Type == "time"` auto-fails quests after hours elapsed (checked every 120 ticks)
- Quest state persisted to SQLite (`entity_quests` table)
- Quest state loaded from SQLite on startup
- Entity detail view shows active quests with progress
- Entity AI Lua API (runtime hooks, not definitions): `world.give_quest`, `world.quest_progress`, `world.quest_set`, `world.talk_to`, `world.deliver_item`
- 12 quest scripts: rat_problem, deliver_sword, lost_heirlooms, deity_whispers, freya_favor, zeus_crazy_task, kobold_menace, vampire_hunt, hag_curse, fairy_escort, bard_ballad, taken_courier
- Recent addition: `taken_courier` — a rescue-style humanoid quest sourced from `frosthold_guard_captain`, tracking a missing courier into `kobold_warren` and ending with a report-back stage.

### Quest Lua definition API (`quest.define`)

```lua
quest.define({
  id = "my_quest",
  title = "My Quest",
  type = "side",           -- main | side | faction | daily | repeatable
  level = 1,
  description = "...",
  source = { type = "npc", npc_id = "frosthold_greta" },
  prerequisites = { level_min = 1, quests_completed = { "other_quest" } },
  stages = {
    {
      id = "stage1",
      name = "Do the thing",
      description = "...",
      requirements = {},   -- prior stage ids
      objectives = {
        { id = "kills", type = "kill_entities", description = "Rats", count = 5, entity_template = "rat" },
        { id = "talk", type = "talk_to_npc", description = "Speak", npc_id = "frosthold_greta" },
        { id = "go", type = "visit_location", description = "Go", location_id = "frosthold" },
        { id = "get", type = "collect_items", description = "Get", count = 1, item_template = "iron_sword" },
        { id = "give", type = "deliver_item", description = "Give", npc_id = "...", item_template = "..." },
      },
    },
  },
  rewards = {
    experience = 50,
    gold = 10,
    items = { { template = "bandage", count = 2 } },
    unlocks = { quests = { "next_quest" } },
  },
  failure_conditions = { { type = "time", hours = 24 } },
})
```

## Combat System (`internal/combat/simple.go`)

- D20-based: attack skill = `10 + level + (DEX-10)/2 + (STR-10)/4`, defense = `8 + level + (DEX-10)/2`
- Weapon damage from lookup table (meleeWeaponDef)
- Armor DR from equipped body/head/offhand
- Critical hit on natural 20
- Blessed entities: +2 attack skill (attacker), +1 DR (defender)
- Per-location Event tracking with `LocationEvents(locID, n)`
- `LootCorpse` transfers inventory + unequipped gear on kill

## Substance System (`internal/entity/effects.go` + `internal/items/substance.go`)

| Concept | Description |
|---------|-------------|
| `SubstanceEffect` | Item def field defining boost stats, crash stats, durations |
| `ActiveEffect` | Runtime state: boost_remaining, crash_remaining, boost/crash modifiers |
| `ApplySubstance()` | Applies boost to entity attributes for `duration` ticks, then crash for `crash_duration` ticks |
| `EffectiveAttrs()` | Returns base attributes + active effect modifiers (boost or crash phase) |
| `TickEffects()` | Decrements remaining ticks, transitions boost→crash, removes expired effects |
| HasEffect() | Check if entity has a specific active effect by name |

### Substances

| Item | Boost Effect | Crash Effect | Duration | Source |
|------|-------------|--------------|----------|--------|
| Beer | +1 STR, +1 CON, -2 INT/WIS/CHA | -1 STR, -1 CON | 20t boost, 10t crash | Innkeepers |
| Wine | +2 STR, +2 CON, -3 INT/WIS/CHA | -1 STR/CON/INT/WIS | 25t boost, 12t crash | Innkeepers, merchants |
| Liquor | +3 STR, +3 CON, -4 INT/WIS/CHA | -2 STR/CON, -1 INT/WIS | 30t boost, 15t crash | Innkeepers, merchants |
| Night Bloom | +4 DEX, -1 CON | -3 DEX | 30t boost, 15t crash | Merchants |
| Sage Leaf | +4 INT, +4 WIS | -3 INT, -3 WIS | 30t boost, 15t crash | Merchants |
| Dreamer's Cap | +2 to all, +6 INT/WIS | -3 all, -4 INT/WIS | 40t boost, 20t crash | Merchants (rare) |
| Trout | Heal 5 HP | — | instant | Fisherman |
| Salmon | Heal 10 HP | — | instant | Fisherman |
| Catfish | Heal 10 HP + 5 FP | — | instant | Fisherman |
| Raw Chicken | Heal 8 HP | — | instant | Farmer (slaughter) |
| Raw Pork | Heal 12 HP | — | instant | Farmer (slaughter) |
| Raw Beef | Heal 20 HP | — | instant | Farmer (slaughter) |
| Raw Mutton | Heal 14 HP | — | instant | Farmer (slaughter) |
| Raw Goat Meat | Heal 14 HP | — | instant | Farmer (slaughter) |
| Egg | Heal 2 HP | — | instant | Chickens |
| Milk | Heal 4 HP | — | instant | Cows |
| Bandage | 2 HP/tick for 10 ticks | — | 10t HOT | Crafted |
| Herbal Poultice | 3 HP/tick for 15 ticks + instant 2 HP | — | 15t HOT | Crafted |
| Healing Salve | 5 HP/tick for 10 ticks + instant 5 HP | — | 10t HOT | Crafted |

### Integration
- `world.use_item(defID)` Lua binding removes item from inventory, applies its `SubstanceEffect` to the consuming entity
- Combat (`SimpleAttack`) uses `EffectiveAttrs()` for STR/DEX calculations
- Economy haggling uses `EffectiveAttrs()` for INT/WIS/CHA
- Divine intervention uses `EffectiveAttrs()` for WIS/STR
- Entity detail UI shows Active Effects card (green for boost, red for crash) with remaining ticks; attributes card shows effective values in (parentheses) when they differ from base
- Effects are persisted to SQLite via `effects_json` column
- Instant consumables (HealHP/HealFP, zero duration) apply immediately with no ActiveEffect entry
- `SubstanceEffect` / `ActiveEffect` have `HealPerTick` / `FPPerTick` fields for heal-over-time items
- `TickEffects()` applies per-tick HOT before stat effect decay

## Aging System (`internal/entity/species.go` + `internal/engine/aging.go`)

| Concept | Description |
|---------|-------------|
| `Age` | Entity int field, incremented every tick |
| `MaxAge` | Species-dependent: chicken=2400, pig=7200, cow=14400, sheep/goat=10800, human=30000, elf=60000, orc=22000, kobold=6000, vampire=0 (immortal), hag=90000, deity=0 (immortal) |
| `LastMealTick` | Tick of last meal; set at entity creation |
| `StarvationThreshold` | MaxAge/3 ticks without food before starvation damage |
| `StarvationDamage` | 1 HP per 10 ticks past threshold |
| `AutoForage` | All species except chicken/pig/cow/sheep/goat auto-reset LastMealTick each tick (never starve) |
| `old_age()` | Kills entity when Age >= MaxAge |
| `starveToDeath()` | Deals damage each tick past threshold until dead |

### Farm Animal Products

| Item | HealHP | Source |
|------|--------|--------|
| Raw Chicken | 8 | Chickens (slaughter at age > 900) |
| Raw Pork | 12 | Pigs (slaughter at age > 2400) |
| Raw Beef | 20 | Cows (slaughter at age > 4800) |
| Raw Mutton | 14 | Sheep (slaughter at age > 3600) |
| Raw Goat Meat | 14 | Goats (slaughter at age > 3600) |
| Egg | 2 | Chickens |
| Milk | 4 | Cows |

### Farm Entities (3 farms, 1 per town)
- Per farm: 1 farmer (scripted/farmer.lua), 4 chickens, 2 pigs, 1 cow, 2 sheep, 1 goat
- All animals are passive Go archetype, faction civilian
- Farmers feed hungry animals and slaughter age-ready ones; sell products at market

## World Building

### Weather (`internal/world/weather.go`)
- Regenerates every 240 ticks on outdoor locations with weather
- Regional climate bias: highlands cold, marches wet, desert arid, forest foggy
- `VisibilityModifier` applied to outdoor combat hit chance
- `TravelSpeedModifier` slows multi-tick cross-region travel
- Harsh outdoor weather: slower idle regen, civilian stress mood
- Lua: `world.weather()`; fisherman skips storms; farmers stay home; guards note fog

### Travel (`internal/world/travel.go` + `internal/engine/travel.go`)
- Bidirectional region exits between all 5 regions
- Instant `move_to` within same region/city tree
- Multi-tick `TravelState` for cross-region hops (weather-scaled)
- Traveling entities skip AI until arrival
- Lua: `world.travel_exits()`, `world.is_traveling()`, `move_to` returns boolean

### Territory (`internal/engine/territory.go`)
- `Location.ControllingFaction` + `ControlStrength` (0–100)
- Updated every 30 ticks from living presence (level-weighted)
- Kills nudge control toward killer faction
- Spawns can require matching controller (`RequireFaction`)
- Lua: `world.location_control(id)`
- Persisted via SQLite location columns

### Wild sites (generated)
| ID | Region | Occupants |
|----|--------|-----------|
| orc_camp | crystal_forest | orcs |
| wolf_den | crystal_forest | wolves |
| spider_grove | crystal_forest | spiders |
| kobold_warren | crystal_forest | kobolds |
| fey_glade | crystal_forest | elves, Willow, Sparkle |
| bandit_camp | golden_plains | bandits |
| bear_den | northern_highlands | bears, Brutus |
| goblin_hollow | northern_highlands | goblin gatherers |
| boar_wallow | sunken_marches | boars |
| ash_ruins | ash_desert | Uruk, bat |
| scorpion_dunes | ash_desert | spider spawns |
| cultist_camp | ash_desert | Keth, Vorg, Zara (cultists) |
| werewolf_cottage | northern_highlands | Cursed Traveler (werewolf) |
| golden_plains_graveyard | golden_plains | Morth the Pale (necromancer) |
| dragon_lair | ash_desert | Ashscale the Ancient (dragon boss, 250 HP) |

Building tags: `inn`, `market`, `temple`, `guardhouse`, `blacksmith`, `forge`, `cauldron`, `workbench`, `farm`, `campfire`, `dungeon`

### Town extras (generated per town)
- `{town}_mine` — Abandoned Mine (miner workplace)
- `{town}_herbalist_hut` — Herbalist Hut (campfire + cauldron tags)

## Entity Spawning (`internal/engine/spawning.go`)

| Concept | Description |
|---------|-------------|
| `SpawnRule` | species, faction, location, count, interval; optional `RequireFaction` |
| `SpawnManager` | `ProcessSpawns(world, entities, tick, rng)` |
| Unique IDs | `{species}_spawn_{name}_{loc}_{tick}_{idx}` |

### Spawn Rules

| Rule | Location | Species | Faction | Count | Interval |
|------|----------|---------|---------|-------|---------|
| orc_patrol | orc_camp | orc | orc | 4 | 120 |
| wolf_pack | wolf_den | wolf | beast | 4 | 90 |
| bandit_camp | bandit_camp | human | bandit | 4 | 150 |
| bear_den | bear_den | bear | beast | 2 | 200 |
| boar_herd | boar_wallow | boar | beast | 2 | 180 |
| rat_infest | rat_king_lair_entrance | rat | vermin | 3 | 60 |
| rat_corridor | rat_king_lair_corridor | rat | vermin | 3 | 60 |
| spider_nest | spider_grove | spider | beast | 2 | 120 |
| goblin_gatherers | goblin_hollow | goblin | goblin | 2 | 180 |
| kobold_warren | kobold_warren | kobold | kobold | 4 | 150 |
| ash_scorpions | scorpion_dunes | spider | beast | 2 | 160 |
| ash_orcs | ash_ruins | orc | orc | 2 | 180 |

## Crafting System (`internal/items/crafting.go`)

| Concept | Description |
|---------|-------------|
| `Recipe` | Struct with ID, Name, Inputs ([]RecipeInput), Output (RecipeOutput), Station string |
| `RecipeInput` | DefID string + Count int |
| `RecipeOutput` | DefID string + Count int |
| `HasMaterials(inv, inputs)` | Check if entity has enough inputs |
| `RemoveInputs(inv, inputs)` | Consume inputs from inventory |
| `world.craft(recipeID)` | Lua: crafts item if at correct station and has materials |
| `world.has_material(defID)` | Lua: checks inventory for a material |
| `world.recipe_info(recipeID)` | Lua: returns recipe details table |

### Stations

| Tag | Locations |
|-----|-----------|
| `forge` | Town blacksmiths |
| `cauldron` | Town temples, hag cottage |
| `campfire` | Farm buildings |
| `workbench` | Guardhouses |

### Recipes

| ID | Inputs | Output | Station |
|----|--------|--------|---------|
| smelt_iron | iron_ore×2, coal×1 | iron_ingot×1 | forge |
| craft_bandage | cloth×2, leather_strips×1 | bandage×2 | workbench |
| cook_poultice | herb×2 | herbal_poultice×1 | campfire |
| refine_salve | herb×4, coal×1 | healing_salve×1 | cauldron |

### Materials (new items)
- `iron_ore`, `coal`, `cloth`, `leather_strips`, `herb`, `iron_ingot`, `copper_ingot`

## Mood System (`internal/entity/mood.go`)

| Concept | Description |
|---------|-------------|
| `MoodModifier` | Struct with Source, Mood string, DecayAtTick uint64 |
| `AddMoodModifier(source, mood, duration)` | Adds a timed mood modifier to entity |
| `EffectiveMood()` | Returns the most frequent mood among active modifiers (or "neutral") |
| `MoodStatMods()` | Returns combined stat modifiers from all active moods |
| `TickMoods(tick)` | Removes expired modifiers, updates `entity.Mood` |
| `util.set_mood(mood, duration)` | Lua: adds mood modifier with optional duration (default 30 ticks) |
| Mood stat table | happy(+1 CHA), angry(+2 STR, -1 WIS), fearful(-2 STR, +2 DEX), stressed(-1 CON, -1 WIS), relaxed(+1 CON, +1 WIS), inspired(+2 INT, +2 WIS), tired(-1 STR/DEX/CON) |

### Integration
- Combat kills add "happy" mood for 30 ticks
- Combat hits add "angry" (attacker) and "fearful" (defender) for 10 ticks
- Mood stat mods are incorporated into `EffectiveAttrs()`
- Entity detail UI shows current mood and active mood modifiers with remaining ticks
- Mood modifiers persisted via `mood_modifiers` field on Entity (SQLite via `effects_json`-style column? — currently in-memory, but entity fully serialized with JSON tags)

## Natural Healing (`internal/entity/effects.go` + `internal/engine/tick.go`)

| Concept | Description |
|---------|-------------|
| Passive regen | Sleeping: +1 HP/tick, +1 FP/tick. Idle/meditating: +1 HP/3ticks, +1 FP/2ticks |
| Combat lockout | No regen while hostiles of your faction are in the same location |
| Heal-over-time | Items with `HealPerTick` / `FPPerTick` field on `SubstanceEffect` apply per-tick healing during `TickEffects()` |
| Items | Bandage (2HP/t × 10t), Herbal Poultice (3HP/t × 15t + instant 2HP), Healing Salve (5HP/t × 10t + instant 5HP) |

### Integration
- `SubstanceEffect` got `HealPerTick` / `FPPerTick` fields (substance.go)
- `ActiveEffect` got matching fields; `TickEffects()` applies per-tick heal before stat effect decay
- `ApplySubstance()` signature extended to pass HOT values
