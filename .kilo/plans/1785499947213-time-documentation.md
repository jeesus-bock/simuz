# Time System Documentation

This document describes the time system used in Simuz, clarifying the different time units, their relationships, and best practices for working with time in the simulation.

## Core Time Unit: Ticks

**Ticks** are the fundamental unit of time in Simuz. The game runs at a rate of **1 tick per real-world second** (1 Hz), as noted in the simulation requirements:

> "A 1 Hz tick loop for AI, combat, quests, weather, travel, and persistence."

Every game action, state update, and event occurs at tick boundaries. This makes ticks ideal for:

- AI decision-making and behavior
- Quest progression and objective tracking
- Combat timing and damage application
- Travel calculations
- Effect duration management
- Relationship tracking

## Higher-Level Time Units

The system provides human-readable time units for UI and reporting:

### GameTime Structure

The `GameTime` struct in `internal/world/time.go` contains:

- **Tick**: Current game time (1 tick = 1 real second)
- **Day**: Number of 24-hour days elapsed (starts at 1)
- **Hour**: Current hour within the day (0-23)
- **Minute**: Current minute within the hour (0-59)
- **Speed**: Game speed multiplier (default 1x)

### Time Conversions

| Unit | Relationship | Description |
|------|-------------|-------------|
| **Tick** | 1 tick = 1 second | Core game loop unit |
| **Minute** | 60 ticks | For displaying shorter durations |
| **Hour** | 3,600 ticks | For displaying session length |
| **Day** | 86,400 ticks | 24-hour periods |
| **Year** | 360 days | 360 days per year |

### Time Calculation Examples

```go
// 5 minutes = 300 ticks
gameTime.Minute = 5

// 2 hours = 7,200 ticks
gameTime.Hour = 2

// 1 day = 86,400 ticks
gameTime.Day = 1

// 6 days = 518,400 ticks
gameTime.Day = 6
```

## Time Phases and Seasons

### Day Phases

- **Dawn** (Hours 5-6): Early morning, reduced visibility
- **Day** (Hours 7-18): Full daylight, normal activity
- **Dusk** (Hours 19-20): Evening, reduced visibility
- **Night** (Hours 21-4): Full night, limited activity for many creatures

### Seasons

Seasons are based on a **360-day year**:

- **Spring** (Days 0-89): Growth, reproduction season
- **Summer** (Days 90-179): Peak activity, hunting season
- **Autumn** (Days 180-269): Harvest, preparation for winter
- **Winter** (Days 270-359): Reduced activity, hibernation

```go
func (gt *GameTime) Season() Season {
	dayInYear := gt.Day % 360
	switch {
	case dayInYear < 90:
		return Spring
	case dayInYear < 180:
		return Summer
	case dayInYear < 270:
		return Autumn
		default:
		return Winter
	}
}
```

## Time Speed

The game time speed multiplier affects how many real-world seconds pass per tick:

### Speed Values

- **1x (default)**: 1 tick per real second
- **2x**: 2 ticks per real second
- **0.5x**: 0.5 ticks per real second (slower game)

### Usage

Speed affects all time-based calculations:

```go
func (gt *GameTime) Advance() {
	gt.Tick++
	gameMinutes := gt.Speed  // Speed determines minutes per tick
	gt.Minute += gameMinutes
	// ... minute/hour/day rollover logic
}
```

## Time in Different Contexts

### AI and Behavior

Many AI behaviors are time-based:

- **Mood decay**: "Your mood will decay in X ticks"
- **Activity duration**: "Patrolling for X ticks"
- **Travel time**: "Traveling for X ticks"

```lua
-- Example: Set mood to last 30 minutes (1,800 ticks)
util.set_mood("furious", 1800)

-- Example: Patrol for 5 minutes (300 ticks)
if world.tick - patrol_start_tick > 300 then
    patrol()
end
```

### Quests and Objectives

Quest timing:

- **Acceptance tick**: When a quest was accepted
- **Objective progress**: Track when objectives are completed
- **Stage timing**: Quest stages can have duration requirements

```go
func (qm *QuestManager) Accept(entityID, questID string, entityLevel int, currentTick uint64) bool {
	// Store acceptance tick for progression calculations
	state := &EntityQuestState{
		AcceptedTick: currentTick,
	}
}
```

### Combat and Effects

Combat timing:

- **Attack cooldowns**: "Can attack again in X ticks"
- **Effect durations**: "Poison effect lasts X ticks"
- **Travel times**: "Travel from location A to B takes X ticks"

```go
// Effect duration example
effect := ActiveEffect{
	Name: "Poison",
	StartTick: currentTick,
	Duration: 300, // 5 minutes
}
```

### Relationships

Relationship tracking:

- **Since tick**: When a relationship was established
- **Activity tracking**: When relationship events occurred

```go
func (e *Entity) AddRelationship(otherID string, relType RelationshipType, tick uint64) {
	rel := Relationship{
		OtherID:   otherID,
		Type:       relType,
		SinceTick:  tick,
	}
	// Store relationship
}
```

## Best Practices

### Choosing Time Units

Use appropriate units for your context:

1. **For AI decisions**: Use ticks (most precise)
2. **For UI display**: Use days/hours/minutes
3. **For progress bars**: Use percentages of total duration
4. **For event logging**: Use ticks for consistency

### Time Calculations

```go
// Convert ticks to human-readable format
func formatTicks(ticks uint64) string {
	days := ticks / 86400
	hours := (ticks % 86400) / 3600
	minutes := (ticks % 3600) / 60
	return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
}
```

### Time Comparisons

```go
// Compare times using ticks (most reliable)
if currentTick - lastAttackTick > 300 {  // 5 minutes
	canAttack = true
}

// For display, convert to hours/minutes
func displayDuration(ticks uint64) string {
	if ticks < 60 {
		return "seconds"
	} else if ticks < 3600 {
		return fmt.Sprintf("%d minutes", ticks/60)
	} else {
		return fmt.Sprintf("%d hours", ticks/3600)
	}
}
```

## Common Time-Related Issues

### Pitfall 1: Mixing Time Units

**BAD:** Mixing different time units in the same calculation
```go
// Wrong: mixing days and ticks
if entity.day > 10 {  // day is not defined, should be Day
	startTravel(entity, "forest")
}
```

**GOOD:** Consistent use of ticks
```go
// Right: everything in ticks, convert for UI only
ticksToTravel := world.TravelTime(entity.location, "forest")
if currentTick - lastTravelTick > ticksToTravel {
	arriveAtDestination()
}
```

### Pitfall 2: Time Speed Assumptions

**BAD:** Assuming constant 1x speed
```go
// Wrong: assumes real-time passing
wait(time.Since(startTime))
```

**GOOD:** Using game ticks
```go
// Right: uses game ticks, accounts for speed
if gameTime.Tick - startTick > requiredTicks {
	completeTask()
}
```

### Pitfall 3: Season/Phase Logic

**BAD:** Checking system time instead of game time
```go
// Wrong: uses real time, not game time
if time.Now().Weekday() == time.Saturday {
	// Weekend logic
}
```

**GOOD:** Using game time phases
```go
// Right: uses game time phases
if gameTime.Phase() == Day {
	// Daytime activities
} else if gameTime.Phase() == Night {
	// Nighttime activities
}
```

## Summary

The Simuz time system uses:

1. **Ticks** as the core unit (1 tick = 1 real second)
2. **GameTime** struct for higher-level units (day, hour, minute)
3. **Day phases** (dawn, day, dusk, night) for behavior changes
4. **Seasons** based on a 360-day year
5. **Time speed** multiplier for controlling game pace

When working with time, prefer ticks for internal logic and calculations, and convert to days/hours/minutes only for UI display. Always account for game speed when dealing with real-world time expectations.

This documentation should help developers understand and work with the time system consistently throughout the codebase.
