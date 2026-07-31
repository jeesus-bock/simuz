# Time System

Simuz has two distinct time concepts that must not be confused.

## Tick (simulation tick)

The fundamental unit. The engine runs at **1 Hz** — one tick per real-world second.

Every timestamp, cooldown, and state counter in the simulation uses ticks:
`KnockedOutTick`, `LastReproductionTick`, `TimeOfDeath`, `SinceTick`, `DecayAtTick`,
`StartTick`, `LastMealTick`, effect durations, mood durations, and quest accepted ticks.

## GameTime (calendar time)

`GameTime` in `internal/world/time.go` tracks the in-world calendar. Each tick
advances the calendar by `Speed` game-minutes (default speed = 24).

| Unit | Game-minutes | Ticks at speed 24 |
|------|-------------|-------------------|
| 1 game-hour | 60 | 2.5 |
| 1 game-day | 1 440 | 60 |
| 1 game-year (360 days) | 518 400 | 21 600 |

Day phases (dawn / day / dusk / night) and seasons (spring / summer / autumn /
winter) derive from `GameTime.Hour` and `GameTime.Day`.

### Adjustable speed

Speed (game-minutes per tick) can be changed at runtime:

- **API:** `PUT /api/v1/world/speed` with body `{"speed": N}` (1..1440)
- **Web UI:** Dashboard dropdown with preset values

Higher speed = faster calendar. At speed 1440, one tick = one game-day.

### Conversion helpers

```go
// internal/world/time.go
world.TicksPerGameDay(speed)       // e.g. TicksPerGameDay(24) → 60
world.GameDaysToTicks(days, speed) // e.g. GameDaysToTicks(80, 24) → 4800
world.GameYearsToTicks(years, speed) // e.g. GameYearsToTicks(80, 24) → 1728000

// On GameTime:
gt.Year()            // current game-year (1-indexed)
gt.DayInYear()       // day within current year (1..360)
gt.TicksPerGameDayFor() // ticks per game-day at current speed
```

## Species lifecycle fields

Species fields in `internal/species/species.go` use **game-years** for age
thresholds and **ticks** for everything else:

| Field | Unit | Example (human) | Meaning |
|-------|------|-----------------|---------|
| `MaxAge` | game-years | 80 | Entity dies after 80 game-years |
| `AdultAge` | game-years | 16 | Entity can reproduce after 16 game-years |
| `GestationTicks` | ticks | 1 200 | Pregnancy lasts 1 200 ticks |
| `StarvationThreshold` | ticks | 259 200 | Starvation begins after 259 200 ticks ≈ 3 real-days of wall time |

### How the conversion works

`processAging` in `internal/engine/aging.go` compares `GameTime.Day` against
`MaxAge * 360` (game-days in a year):

```go
maxAgeDays := ent.MaxAge * world.GameYearDays
if sim.Time.Day >= maxAgeDays { ... }
```

`IsAdult(currentDay)` on `Entity` does the same for `AdultAge`:

```go
adultDay := sp.AdultAge * 360
return currentDay >= adultDay || e.Level >= 3
```

## Entity timestamp fields

All timestamp fields on `Entity` (`internal/entity/entity.go`) are in
simulation ticks:

| Field | Description |
|-------|-------------|
| `Age` | Ticks alive; increments by 1 per tick |
| `LastMealTick` | Tick of last meal |
| `LastReproductionTick` | Tick of last reproduction |
| `KnockedOutTick` | Tick when knocked out |
| `TimeOfDeath` | Tick when entity died |

## Effects and moods

`ActiveEffect.StartTick`, `ActiveEffect.Duration`, `MoodModifier.DecayAtTick`
are all in ticks.

## Travel

`TravelTime` in `internal/world/travel.go` returns ticks. Multi-tick cross-region
travel stores departure tick and ETA tick.

## Quests

`EntityQuestState.AcceptedTick` is a simulation tick. Quest failure conditions
with type `"time"` use game-hours (converted from ticks via speed):

```go
totalHours := (sim.Tick * uint64(sim.Time.Speed)) / 60
```

## Lua scripting

In Lua scripts, `world.tick` is the current simulation tick. Mood durations
passed to `util.set_mood(name, ticks)` are in ticks.

## Rules of thumb

1. **Store and compare in ticks.** Convert to game-days/hours/years only for display.
2. Species `MaxAge` and `AdultAge` are the exception — they are authored in
   game-years and compared against `GameTime.Day` (converted via `* 360`).
3. When adding a new time-based field, always document its unit in a comment.
4. Speed is adjustable at runtime via `PUT /api/v1/world/speed` or the dashboard UI.
