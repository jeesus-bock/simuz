# World Design

## Hierarchical Location Model

The world is a tree of locations. Each location has a parent (except the root), children (sublocations), and neighbors with travel connections.

```
World (root)
├── Region: The Northern Reaches
│   ├── City: Frosthold
│   │   ├── District: Merchant Quarter
│   │   │   ├── Building: The Dwarven Anvil (smithy)
│   │   │   │   ├── Room: Forge
│   │   │   │   ├── Room: Sales Floor
│   │   │   │   └── Room: Storage
│   │   │   ├── Building: Temple of Moradin
│   │   │   └── Street: Silver Way
│   │   │       └── Stall: Herbalist Cart
│   │   ├── District: Noble Ward
│   │   │   ├── Building: Lord's Manor
│   │   │   └── Gardens
│   │   └── District: Docks
│   │       ├── Pier: North Pier
│   │       └── Warehouse District
│   ├── Village: Oakhaven
│   │   ├── Tavern: The Sleeping Fox
│   │   └── Mill
│   ├── Forest: Whisperwood
│   │   ├── Clearing: Old Stone Circle
│   │   └── Cave: Goblins' Den
│   └── Mountain: Mount Grimfang
│       ├── Peak
│       └── Mine: Abandoned Silver Mine
├── Region: Sunken Coast
└── Region: The Scar
```

### Data Model

```
type Location struct {
    ID          string
    Name        string
    Type        LocationType  // region, city, district, building, room, outdoors
    ParentID    *string       // nil for root
    Children    []string      // sublocation IDs
    Exits       []Exit        // travel connections to other locations
    Position    Position      // 2D coordinates within parent
    IsOutside   bool          // exposed to weather, day/night
    Terrain     TerrainType   // forest, desert, water, etc.
}
```

## Distances & 2D Map

Each location has a 2D position relative to its parent. The world can be visualized as nested maps:

- **World level** → regions laid out on a continent map
- **Region level** → cities and landmarks positioned within
- **City level** → districts arranged on a city grid
- **Building level** → rooms on a floorplan

```
type Position struct {
    X float64
    Y float64
}

func Distance(a, b Location) float64 {
    // Euclidean distance within shared parent
    dx := a.Position.X - b.Position.X
    dy := a.Position.Y - b.Position.Y
    return math.Sqrt(dx*dx + dy*dy)
}
```

Distances are only meaningful between siblings (same parent). To calculate distance between arbitrary locations, traverse up to the common ancestor and sum segment distances.

## Exits & Travel Connections

Exits are directed connections between locations at any level. They define how entities move through the world.

```
type Exit struct {
    FromID      string
    ToID        string
    TravelType  TravelMode  // walk, ride, sail, teleport, fly
    Distance    float64     // in kilometers
    TimeCost    Duration    // override: time to traverse
    Description string      // "a winding mountain path"
    Conditions  []Condition // blocked by weather, requires key, etc.
}
```

### Connection Examples

| From | To | Travel Type | Time | Conditions |
|------|-----|-------------|------|------------|
| Frosthold Gate | Oakhaven Road | walk | 4 hours | none |
| Frosthold Gate | Silver Way | walk | 5 min | none |
| Oakhaven | Whisperwood Edge | walk | 1 hour | none |
| Whisperwood Edge | Goblin's Den | walk | 20 min | hidden path |
| Frosthold Docks | Sunken Coast | sail | 3 days | clear weather |
| Abandoned Silver Mine | Frosthold Throne Room | teleport | instant | requires key |

### Sublocation Exits

Sublocations (rooms within a building, districts within a city) have exits that connect to other sublocations:

```
Room "Forge" → Exit through "Front Door" → "Sales Floor"
Room "Sales Floor" → Exit to "Silver Way" (street)
```

Exits can be nested — leaving a building puts you on the street of the parent district.

## Travel System

### Travel Modes

| Mode | Speed | Cost | Notes |
|------|-------|------|-------|
| walk | 5 km/h | free | Base travel for all entities |
| ride | 15 km/h | horse/mount cost | 3x walking speed |
| sail | 10 km/h | ship cost | Water routes only |
| fly | 30 km/h | flying mount or magic | Ignores terrain |
| teleport | instant | high (spell/scroll) | Requires magic or special exit |
| forced march | 6 km/h | fatigue cost | Walk but faster, costs stamina |

### Travel Process

1. Entity queues travel action: destination + mode
2. Tick loop calculates ETA based on distance + mode speed
3. Entity enters "traveling" state (cannot act during travel)
4. On arrival, entity is moved to destination and can act again
5. Travel can be interrupted (ambushes, weather events)

```
type TravelState struct {
    EntityID    string
    From        string
    To          string
    Mode        TravelMode
    StartedAt   uint64    // tick
    Duration    uint64    // ticks until arrival
    Remaining   uint64    // ticks left
}
```

## Day/Night Cycle

- 1 day = 86,400 real seconds, but the simulation runs on accelerated time
- Default: 1 real hour = 1 simulated day (~24x speed)
- Configurable: `TimeScale` parameter
- Each tick (1 real sec) advances simulated time by N seconds

```
type WorldTime struct {
    Tick         uint64    // total ticks elapsed
    SimulatedSec uint64    // total simulated seconds elapsed
    TimeScale    float64   // simulated seconds per real second
}
```

### Day/Night States

| Phase | Time of Day | Lighting | Effects |
|-------|------------|----------|---------|
| Dawn  | 05:00-06:00 | dim -> bright | Night creatures retreat |
| Day   | 06:00-20:00 | bright | Normal activity |
| Dusk  | 20:00-21:00 | bright -> dim | Day creatures seek shelter |
| Night  | 21:00-05:00 | dark (moonlight) | Night creatures active, -4 perception for diurnal |

Most creatures follow a diurnal/nocturnal schedule:
- **Diurnal**: awake day, sleep night (humans, elves, dwarves)
- **Nocturnal**: awake night, sleep day (owlbears, drow, bats, some undead)
- **Crepuscular**: active at dawn/dusk (some predators)

Sleep schedules are part of entity behavior:

```
type CircadianRhythm int
const (
    Diurnal CircadianRhythm = iota
    Nocturnal
    Crepuscular
)
```

## Weather System

Weather applies only to `IsOutside: true` locations. Inside locations have controlled conditions.

### Weather Properties

```
type Weather struct {
    LocationID  string
    Temperature TemperatureRange  // freezing, cold, mild, warm, hot
    Precipitation Precipitation  // none, drizzle, rain, storm, blizzard
    Wind        WindLevel        // calm, breeze, windy, gale, hurricane
    Visibility  int              // visibility penalty (0-10)
    CloudCover  int              // 0-100%
    Season      Season           // spring, summer, autumn, winter
}
```

### Weather Generation

Weather is generated per region and varies by:
- **Season**: baseline temperature + precipitation probability
- **Terrain**: deserts are hot/dry, forests are temperate, mountains are cold
- **Random variation**: daily weather rolls, storms develop and pass

```
type ClimateProfile struct {
    BaseTemp       TemperatureRange
    TempVariance   int              // +/- degrees
    RainChance     float64          // 0-1 per day
    StormChance    float64
    WindBaseline   WindLevel
}
```

Example climates:

| Terrain | Season | Temp | Rain | Wind |
|---------|--------|------|------|------|
| Temperate Forest | Spring | mild | 40% | breeze |
| Temperate Forest | Summer | warm | 30% | calm |
| Desert | Summer | hot | 5% | windy |
| Mountains | Winter | freezing | 60% (snow) | gale |
| Swamp | Autumn | mild | 70% | breeze |

### Weather Effects on Gameplay

| Condition | Effect |
|-----------|--------|
| Rain | -2 perception, muddy terrain slows travel 25% |
| Storm | -4 perception, ranged attacks -2, travel becomes dangerous |
| Blizzard | -6 perception, travel impossible without shelter, cold damage |
| Extreme Heat | Fatigue cost for physical activity |
| Fog | -4 visibility, easy to hide/ambush |
| High Wind | -2 ranged attacks, flying creatures must land |

## Time Model Summary

- **Tick**: 1 real second, the smallest simulation increment
- **Round**: 1 tick (combat operates per tick)
- **Turn**: variable — an entity's action opportunity
- **Day**: configurable (default 24 simulated hours = 1 real hour)
- **Scheduled events**: fire at specific tick/datetime (e.g., "every dawn", "every 15 min", "on day 3 at noon")
