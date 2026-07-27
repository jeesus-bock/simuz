# World Generation & Seed Data

## Overview

The simulation needs to be bootable from nothing. Generators produce coherent worlds — locations, entities, items, quests — from a seed, so every run starts with a living world. The system is deterministic (given the same seed, same output), scriptable via YAML templates, and extensible in Go.

---

## Pipeline

```
seed ──→ World Generator ──→ Region Generator ──→ Settlement Generator ──→ Dungeon Generator
          │                        │                        │                       │
          ▼                        ▼                        ▼                       ▼
       ┌──────┐                ┌────────┐              ┌───────────┐           ┌──────────┐
       │Biome │                │Climate │              │Districts  │           │Rooms     │
       │Map   │                │Regions │              │Buildings  │           │Monsters  │
       │      │                │        │              │NPCs       │           │Loot      │
       └──────┘                └────────┘              └───────────┘           └──────────┘
          │                        │                        │                       │
          └────────────────────────┴────────────────────────┴───────────────────────┘
                                               │
                                               ▼
                                    ┌──────────────────┐
                                    │   Entity Pop     │
                                    │   Item Dist      │
                                    │   Quest Gen      │
                                    └──────────────────┘
                                               │
                                               ▼
                                    ┌──────────────────┐
                                    │   world.yaml     │
                                    │   + entities/    │
                                    │   + quests/      │
                                    └──────────────────┘
```

---

## Architecture

### Package Layout

```
internal/
└── gen/
    ├── world.go              # Top-level orchestrator
    ├── biome.go              # Biome/climate generation
    ├── region.go             # Region subdivision
    ├── settlement.go         # Town/city layout
    ├── dungeon.go            # Dungeon room graph
    ├── populate.go           # Entity spawning
    ├── items.go              # Item placement
    ├── quests.go             # Quest generation
    ├── templates.go          # YAML template loader
    ├── rng.go                # Seeded PRNG wrapper
    └── output.go             # Serialize to world files

data/
├── templates/
│   ├── biomes.yaml           # Biome definitions
│   ├── regions.yaml          # Region type templates
│   ├── settlements.yaml      # Settlement templates (tavern, shop, temple, etc.)
│   ├── dungeons.yaml         # Dungeon room templates
│   ├── npc_roles.yaml        # NPC role → species, equipment, AI
│   └── quest_hooks.yaml      # Quest hook templates
├── worlds/                   # Generated world output (gitignored)
│   └── default/              # Default generated world
└── seed/                     # Hand-crafted seed data
    ├── locations.yaml         # Canonical locations (world → region → city)
    ├── entities.yaml          # Named NPCs and creature spawns
    └── items.yaml             # Static item placements
```

### Generator Interface

```go
type Generator struct {
    rng    *SeededRNG
    config GenConfig
}

type GenConfig struct {
    Seed        string                 // Deterministic seed (default: random)
    WorldSize   int                    // 1-100, affects region count
    DetailLevel int                    // 1-100, affects building/NPC density
    Templates   TemplateLoader         // YAML template source
    Output      string                 // Output directory
}

func (g *Generator) Generate() (*WorldSnapshot, error)

type WorldSnapshot struct {
    Locations []LocationDef    `yaml:"locations"`
    Entities  []EntityDef      `yaml:"entities"`
    Items     []ItemDef        `yaml:"items"`
    Quests    []QuestDef       `yaml:"quests"`
    Flags     map[string]any   `yaml:"flags"`
}
```

### Seeded RNG

All randomness flows from a single seed, making generation deterministic and reproducible.

```go
type SeededRNG struct {
    src *rand.Rand
}

func NewSeededRNG(seed string) *SeededRNG {
    h := sha256.Sum256([]byte(seed))
    return &SeededRNG{src: rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(h[:8]))))}
}

func (r *SeededRNG) Intn(n int) int
func (r *SeededRNG) Float64() float64
func (r *SeededRNG) Shuffle(slice any)
func (r *SeededRNG) Pick[T any](items []T) T
func (r *SeededRNG) WeightedPick(items []WeightedItem) int
func (r *SeededRNG) Derive(label string) *SeededRNG   // Branch RNG for deterministic sub-generation
```

`Derive` creates a child RNG from a label, so adding a new generator stage later doesn't change output of earlier stages. Sub-generators for each region, settlement, or entity get their own derived RNG.

---

## Stage 1: World Generator

Produces the top-level biome map and region layout.

### Input: Seed, WorldSize

### Output: Location tree root + children

```yaml
# Example output
locations:
  - id: "aetheria"
    name: "Aetheria"
    type: world
    climate: temperate
    area_sq_km: 1000000
    children:
      - id: "northern_highlands"
        name: "Northern Highlands"
        type: region
        climate: alpine
        biome: mountain_forest
        pos: {x: 100, y: 200}
        children: [...]
      - id: "sunken_marches"
        name: "Sunken Marches"
        type: region
        climate: subtropical
        biome: swamp
        pos: {x: 400, y: 600}
        children: [...]
      ...
```

### Biome Template (data/templates/biomes.yaml)

```yaml
biomes:
  - id: temperate_forest
    name: "Temperate Forest"
    climate: temperate
    frequency: 0.25                        # Probability weight
    settlement_density: 0.6                # 0-1, how many settlements
    spawn_table:
      deer: {weight: 0.4, min_count: 3, max_count: 8}
      boar: {weight: 0.2, min_count: 1, max_count: 3}
      wolf: {weight: 0.1, min_count: 2, max_count: 5}
    resources:
      - timber
      - herbs

  - id: desert
    name: "Desert"
    climate: arid
    frequency: 0.1
    settlement_density: 0.2
    spawn_table:
      giant_scorpion: {weight: 0.3, min_count: 1, max_count: 2}
      sand_worm: {weight: 0.1, min_count: 1, max_count: 1}
    resources:
      - salt
      - obsidian
```

---

## Stage 2: Region Generator

Subdivides the world into regions, assigning biomes based on world position and climate bands.

### Algorithm

1. Divide world into a grid (size based on WorldSize)
2. For each cell, assign biome using noise-based sampling + climate band (latitude)
3. Merge adjacent cells with same biome into named regions
4. Generate region names from Markov chain trained on a name corpus
5. Determine natural resources and terrain features

---

## Stage 3: Settlement Generator

Populates regions with towns, cities, villages, and standalone locations.

### Settlement Template (data/templates/settlements.yaml)

```yaml
settlement_templates:
  - id: village
    name: "{name} Village"
    min_population: 50
    max_population: 500
    buildings:
      - template: "house"               # Residential
        count_range: [10, 50]
        weight: 0.5
      - template: "tavern"
        count_range: [1, 2]
        weight: 0.3
      - template: "general_store"
        count_range: [0, 1]
        weight: 0.15
      - template: "temple"
        count_range: [0, 1]
        weight: 0.05
    npc_roles:
      - role: "innkeeper"
        count: 1
        weight: 0.4
      - role: "shopkeeper"
        count: 1
        weight: 0.3
      - role: "farmer"
        count_range: [5, 20]
        weight: 0.8
      - role: "guard"
        count_range: [2, 5]
        weight: 0.2
    ai_archetype_default: passive
    quest_hooks:
      - template: "rat_infestation"
        weight: 0.3
      - template: "bandit_threat"
        weight: 0.2
      - template: "missing_person"
        weight: 0.1

  - id: city
    name: "{name}"
    min_population: 5000
    max_population: 50000
    districts:
      - name: "Market District"
        weight: 0.3
        buildings:
          - template: "market_stall" {count_range: [10, 50]}
          - template: "tavern" {count_range: [3, 8]}
          ...
      - name: "Noble Quarter"
        weight: 0.1
        buildings: [...]
      - name: "Slums"
        weight: 0.2
        buildings: [...]
      - name: "Industrial"
        weight: 0.15
        buildings: [...]
      - name: "Temple District"
        weight: 0.1
        buildings: [...]
```

### Building Templates

```yaml
building_templates:
  - id: tavern
    name: "{adj} {noun}"              # e.g., "The Rusty Nail", "The Drunken Dragon"
    tags: [tavern, vendor_food, vendor_drink, lodgings]
    interior_rooms:
      - name: "Common Room"
        type: room
        exits: [street]
      - name: "Cellar"
        type: room
        exits: [common_room]
      - name: "Guest Rooms"
        type: room
        count_range: [2, 6]
        exits: [common_room]
    npcs:
      - role: "innkeeper"
        count: 1
        spawn_location: "Common Room"
      - role: "barmaid"
        count_range: [1, 3]
      - role: "patron"
        count_range: [3, 12]
    loot_table:
      - item: "ale"
        count_range: [20, 100]
      - item: "wine"
        count_range: [5, 20]

  - id: dungeon_entrance
    name: "{adj_foreboding} {noun_cave}"
    tags: [dungeon, dangerous]
    type: cave
    dungeon_floor_count_range: [1, 5]
```

---

## Stage 4: Dungeon Generator

Procedurally generates dungeon floor plans as connected room graphs.

### Algorithm

1. Pick dungeon template (cave, crypt, mine, tower, sewer)
2. Generate room density based on DifficultyLevel
3. Place rooms using BSP (binary space partition) or random walk
4. Connect rooms with corridors
5. Assign room types (entrance, treasure, monster, boss, exit, trap, empty)
6. Populate monsters and loot per room type
7. Place exits (stairs down, teleporters, secret doors)

```yaml
dungeon_templates:
  - id: goblin_cave
    name: "{color} Caves"
    floor_count_range: [1, 3]
    rooms_per_floor: [5, 15]
    room_types:
      entrance: {weight: 0.05, monsters: none}
      monster:  {weight: 0.40, spawn_table: {goblin: 0.6, goblin_shaman: 0.2, wolf: 0.2}}
      treasure: {weight: 0.10, loot_table: {gold: [10, 50], rusty_sword: 0.3}}
      boss:     {weight: 0.05, spawn: {goblin_chief: 1}, loot_table: {gold: [50, 200]}}
      trap:     {weight: 0.10, trap_types: [pit, arrow, tripwire]}
      empty:    {weight: 0.30}
    exit_type: stairs_down
```

---

## Stage 5: Entity Population

Fills locations with NPCs and creatures based on species tables, settlement templates, and biome spawn tables.

### NPC Role Template (data/templates/npc_roles.yaml)

```yaml
npc_roles:
  - id: innkeeper
    species_weights:
      human: 0.7
      halfling: 0.2
      dwarf: 0.1
    equipment:
      - {item: "apron", slot: "body", weight: 1.0}
      - {item: "dagger", slot: "weapon", weight: 0.6}
      - {item: "club", slot: "weapon", weight: 0.4}
    ai: guarded                              # AI archetype
    ai_script: "innkeeper"                   # Overridden if scripted
    behavior:                                # Default personality params
      work_hours: [6, 22]
      home_location: tavern
      dialog_tags: [innkeeping, gossip, rumors]
    loot:                                    # What they carry
      gold_range: [5, 50]

  - id: town_guard
    species_weights:
      human: 0.8
      half_orc: 0.15
      dwarf: 0.05
    equipment:
      - {item: "leather_armor", slot: "body", weight: 1.0}
      - {item: "spear", slot: "weapon", weight: 0.7}
      - {item: "shortsword", slot: "weapon", weight: 0.3}
      - {item: "shield", slot: "offhand", weight: 0.6}
    ai: patrol
    ai_script: "patrol_guard"
    behavior:
      waypoint_range: 200                    # Patrol radius (meters)
      shift_hours: [8, 8]                    # 8 on, 8 off
      dialog_tags: [guard_duty, gossip, warnings]

  - id: merchant
    species_weights:
      human: 0.6
      dwarf: 0.2
      halfling: 0.15
      gnome: 0.05
    equipment:
      - {item: "fine_clothes", slot: "body", weight: 1.0}
      - {item: "dagger", slot: "weapon", weight: 0.8}
    ai: greedy
    behavior:
      work_hours: [8, 20]
      haggle_factor: [0.8, 1.2]              # Price multiplier range
      dialog_tags: [trade, gossip, rumors]
    inventory:
      stocks_from_region: true               # Inventory reflects local resources
      restock_interval_hours: 24
      categories:
        - weapons
        - armor
        - potions
        - food

  - id: generic_creature
    # Used for biome spawns — no role, just species + AI
    ai: aggressive                           # Default wild creature AI
```

### Name Generation

Names are generated using Markov chains, syllable tables, or pre-built lists:

```yaml
name_generators:
  - id: fantasy_human
    type: markov
    order: 2
    corpus_file: "data/names/human.txt"
    formats:
      - "{first} {last}"
      - "{title} {first} of {location}"
  - id: fantasy_tavern
    type: template
    patterns:
      - "The {adj} {noun}"
      - "The {noun} and {noun}"
      - "{adj_old} {noun}"
    word_sets:
      adj: [Drunken, Rusty, Golden, Silver, Broken, Green, Red, Crowned, Wandering, Jolly]
      noun: [Dragon, Nail, Tankard, Horse, Fox, Bear, Lion, Cask, Lantern, Star]
```

---

## Stage 6: Quest Generator

Creates quests based on world state, faction needs, settlement context, and entity personalities.

### Quest Hook Templates (data/templates/quest_hooks.yaml)

```yaml
quest_hooks:
  - id: rat_infestation
    title: "The {location_name} Pest Problem"
    type: side
    level_range: [1, 3]
    prerequisites:
      - type: none                          # Available to anyone
    giver_role: innkeeper
    stages:
      - id: clear_pests
        objectives:
          - type: kill_entities
            entity_template: giant_rat
            count_range: [5, 10]
        rewards:
          gold_range: [20, 80]
          items_chance:
            health_potion: 0.5
    failure_conditions:
      - type: time_limit
        hours_range: [24, 72]

  - id: bandit_threat
    title: "Bandits on the {road_name}"
    type: side
    level_range: [3, 6]
    prerequisites:
      - type: none
    giver_role: town_guard
    stages:
      - id: investigate
        objectives:
          - type: visit_location
            location_type: bandit_camp
      - id: clear_bandits
        objectives:
          - type: kill_entities
            entity_template: bandit
            count_range: [4, 8]
    rewards:
      gold_range: [100, 500]
      faction: {town: 20}

  - id: missing_person
    title: "Find {npc_name}"
    type: side
    level_range: [2, 5]
    giver_role: any
    stages:
      - id: investigate
        objectives:
          - type: talk_to_npc
            npc_role: any
            count: {npc_count_range: [2, 5]}
            hint: "Ask around town for clues"
      - id: rescue
        objectives:
          - type: kill_entities
            entity_template: kidnapper
          - type: escort_npc
            npc_id: "{target_npc}"
            destination: "{origin_location}"
    rewards:
      gold_range: [50, 200]
```

The quest generator:
1. Scans settlements for available quest hooks (based on giver NPC presence)
2. Selects hooks that match the world's context (location names, NPC names)
3. Fills in template variables — `{location_name}`, `{npc_name}`, `{road_name}` — from generated world data
4. Assigns appropriate levels, counts, and rewards based on area difficulty
5. Chains related hooks into mini-questlines when possible

---

## Stage 7: Item Distribution

Places items in the world based on location context and templates.

```yaml
item_placement:
  - role: merchant
    inventory:
      - item: health_potion
        count_range: [3, 10]
        price_mult: [0.8, 1.2]
      - item: iron_sword
        count_range: [1, 3]
      - item: leather_armor
        count_range: [1, 3]
      # Items derived from local resources
      - from_region_resources: true
        item_count_range: [5, 15]

  - role: innkeeper
    inventory:
      - item: ale
        count_range: [20, 100]
      - item: bread
        count_range: [10, 50]
      - item: cheese
        count_range: [5, 20]

  - dungeon_loot:
      common:
        gold_range: [5, 50]
        items:
          - health_potion: 0.3
          - torch: 0.8
          - rusty_dagger: 0.2
      rare:
        gold_range: [50, 200]
        items:
          - silver_ring: 0.1
          - magic_scroll: 0.05
      boss:
        gold_range: [100, 500]
        items:
          - weapon_or_armor: 0.8
          - unique_item: 0.2
```

---

## CLI Integration

```go
// cmd/simuz/main.go

var rootCmd = &cobra.Command{
    Use: "simuz",
    // ...
}

var generateCmd = &cobra.Command{
    Use:   "generate",
    Short: "Generate a new world",
    Run: func(cmd *cobra.Command, args []string) {
        gen := gen.Generator{
            Config: gen.GenConfig{
                Seed:        cmd.Flag("seed").Value.String(),
                WorldSize:   viper.GetInt("world_size"),
                DetailLevel: viper.GetInt("detail_level"),
                Output:      viper.GetString("world_dir"),
            },
        }
        world, err := gen.Generate()
        // Write world.yaml + entities/* + quests/* to output dir
    },
}

func init() {
    rootCmd.AddCommand(generateCmd)
    generateCmd.Flags().String("seed", "", "Generation seed (default: random)")
    generateCmd.Flags().Int("world-size", 50, "World size 1-100")
    generateCmd.Flags().Int("detail", 50, "Detail level 1-100")
}
```

Usage:

```bash
simuz generate --seed=myworld --world-size=70 --detail=80
# Creates data/worlds/myworld/ with full world snapshot
```

---

## Integration with Simulation

The generated data is loaded at startup by the existing loaders:

```go
func main() {
    worldDir := viper.GetString("world_dir")  // "data/worlds/default"

    // If the directory doesn't exist, auto-generate
    if _, err := os.Stat(worldDir); os.IsNotExist(err) {
        gen := gen.Generator{...}
        gen.Generate()
    }

    // Load world data (same path as hand-crafted seed data)
    loader := data.NewLoader(worldDir)
    locations := loader.LoadLocations()
    entities := loader.LoadEntities()
    items := loader.LoadItems()

    // Start simulation with loaded data
    sim := engine.NewSimulation(locations, entities, items)
    sim.Run()
}
```

Alternatively, the generator can produce data incrementally:

```go
// Generate on demand when a player moves to an unexplored region
region := sim.World.GetRegion("unknown_lands")
if region == nil {
    region = gen.GenerateRegion("unknown_lands", seed.Derive("region_unknown"))
    sim.World.AddRegion(region)
}
```

---

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/admin/generate` | Trigger world generation with given seed/params |
| GET | `/admin/generate/status` | Generation progress |
| POST | `/admin/generate/region/:id` | Generate a single region on demand |
| POST | `/admin/generate/dungeon` | Generate a dungeon at a location |

---

## Ship-with Seed Lua Scripts

Pre-coded AI scripts shipped in `scripts/ai/`:

| Script | Archetype Equivalent | Role |
|--------|---------------------|------|
| `merchant.lua` | greedy | Haggling, restocking, closing shop at night |
| `innkeeper.lua` | guarded | Serving drinks, renting rooms, gossip, calling guards on trouble |
| `patrol_guard.lua` | patrol | Waypoint walking, calling for backup, shift change |
| `town_guard.lua` | guarded | Standing post, warning, attacking |
| `farmer.lua` | passive | Daily routine: wake → eat → work fields → eat → sleep |
| `blacksmith.lua` | passive | Crafting during work hours, selling, sleeping |
| `beggar.lua` | passive | Panhandling, sleeping in doorways, fleeing guards |
| `wild_animal.lua` | aggressive | Hunting, fleeing fire, territorial marking |
| `dragon.lua` | scripted | Multi-phase boss: fly → land → breath attack → flee when HP low |
| `bandit.lua` | greedy | Ambush, loot, flee when outnumbered |

Each script is ~30-100 lines and lives in `scripts/ai/` alongside user-created scripts.

---

## Project Layout Additions

```
simuz/
├── internal/
│   ├── gen/                   # World generation
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
│   └── ...
├── data/
│   ├── templates/             # YAML generation templates
│   │   ├── biomes.yaml
│   │   ├── settlements.yaml
│   │   ├── dungeons.yaml
│   │   ├── npc_roles.yaml
│   │   └── quest_hooks.yaml
│   ├── worlds/                # Generated world output (gitignored)
│   │   └── default/
│   │       ├── world.yaml
│   │       ├── entities.yaml
│   │       ├── items.yaml
│   │       └── quests.yaml
│   └── seed/                  # Hand-crafted seed data overrides
│       ├── locations.yaml
│       ├── entities.yaml
│       └── items.yaml
├── scripts/
│   ├── ai/
│   │   ├── merchant.lua       # Pre-coded AI scripts
│   │   ├── innkeeper.lua
│   │   ├── patrol_guard.lua
│   │   ├── town_guard.lua
│   │   ├── farmer.lua
│   │   ├── blacksmith.lua
│   │   ├── beggar.lua
│   │   ├── wild_animal.lua
│   │   ├── dragon.lua
│   │   ├── bandit.lua
│   │   └── lib/
│   │       ├── pathfinding.lua
│   │       ├── dialog.lua
│   │       └── combat.lua
│   └── ...
└── ...
```

---

## Determinism Guarantees

- Same seed → identical world every time
- `Derive(label)` branching ensures generator stage changes don't cascade
- YAML templates are read in fixed order (sorted by filename)
- All randomness uses the seeded PRNG, never `math/rand` global or `crypto/rand`
- Adding a new generator stage mid-pipeline requires a new `Derive` label, not insertion
