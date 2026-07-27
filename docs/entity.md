# Entity Design

## Entity Types

```
type EntityType int
const (
    Creature EntityType = iota  // living being with AI
    NPC                          // non-player character (scripted)
    Resource                     // harvestable (trees, ore veins)
    Item                         // portable object
    Projectile                   // arrow, fireball in flight
    Corpse                       // dead creature remains
)
```

## Entity Data Model

```
type Entity struct {
    ID             string
    Name           string
    Species        string            // "human", "elf", "orc", etc.
    Type           EntityType
    LocationID     string            // current location
    TravelState    *TravelState      // nil if not traveling
    Attributes     Attributes        // STR, DEX, CON, INT, WIS, CHA
    Derived        DerivedStats      // computed from attributes
    Skills         map[string]int    // skill name → level
    HitPoints      HP
    FatiguePoints  FP
    StatusEffects  []StatusEffect
    Inventory      []string          // item IDs
    Equipped       Equipment         // worn/wielded items
    Behavior       *BehaviorState    // AI state (nil for items/resources)
    Circadian      CircadianRhythm   // diurnal / nocturnal / crepuscular
    Age            Age
    Faction        string
    Knowledge      map[string]int    // skills, lore, memory
}
```

## Attribute System (d20-Inspired)

Six core attributes, range 3–20 for mortals (higher for legendary/mythic beings).

| Attribute | Abbr | Measures | Primary Uses |
|-----------|------|----------|-------------|
| Strength | STR | Physical power | Melee damage, carry capacity, athletic feats |
| Dexterity | DEX | Agility, reflexes | Ranged attack, dodge, initiative, stealth |
| Constitution | CON | Endurance, health | Hit Points, fatigue, poison resistance |
| Intelligence | INT | Reasoning, memory | Knowledge, magic, skill points, languages |
| Wisdom | WIS | Perception, intuition | Will saves, perception, insight, healing |
| Charisma | CHA | Personality, presence | Social influence, leadership, intimidation |

### Attribute Modifier

```
modifier = (score - 10) / 2   // rounded down
```

| Score | Modifier | Example |
|-------|----------|---------|
| 3 | -4 | Crippling |
| 6-7 | -2 | Poor |
| 8-9 | -1 | Below average |
| 10-11 | 0 | Average human |
| 12-13 | +1 | Above average |
| 14-15 | +2 | Exceptional |
| 16-17 | +3 | Amazing |
| 18-19 | +4 | Legendary |
| 20 | +5 | Mythic |

```
type Attributes struct {
    STR int
    DEX int
    CON int
    INT int
    WIS int
    CHA int
}

func (a Attributes) Mod(score int) int {
    return (score - 10) / 2
}
```

## Derived Statistics

Computed from base attributes on creation and when attributes change.

```
type DerivedStats struct {
    MaxHP          int    // = CON * 2 + species_modifier
    MaxFP          int    // = CON + STR/2
    BaseSpeed      int    // meters per tick (walking)
    Initiative     int    // = DEX mod + species_mod
    CarryCapacity  int    // = STR * 5 (kg)
    BaseDodge      int    // = DEX/2 + any armor penalty
    NaturalDR      int    // natural armor (0 for humans)
    Size           SizeModifier
}
```

| Size | Modifier | Example | Base HP Mod |
|------|----------|---------|-------------|
| Fine | +8 | fly, ant | -4 |
| Diminutive | +4 | fairy, mouse | -2 |
| Tiny | +2 | cat, sprite | -1 |
| Small | +1 | halfling, goblin, gnome | 0 |
| Medium | 0 | human, elf, orc | 0 |
| Large | -1 | ogre, troll, horse | +2 |
| Huge | -2 | giant, dragon, elephant | +4 |
| Gargantuan | -4 | ancient dragon, kraken | +8 |
| Colossal | -8 | tarrasque | +16 |

## Skill System

Skills represent training in specific areas. Skill level range: 0 (untrained) to 20+ (legendary).

Each skill is tied to a governing attribute:

### Combat Skills

| Skill | Attribute | Use |
|-------|-----------|-----|
| Melee (Broadsword) | DEX | Sword fighting |
| Melee (Axe) | DEX | Axe fighting |
| Melee (Spear) | DEX | Spear/polearm fighting |
| Melee (Unarmed) | DEX | Fist-fighting, grappling |
| Ranged (Bow) | DEX | Archery |
| Ranged (Crossbow) | DEX | Crossbow |
| Ranged (Thrown) | DEX | Throwing weapons |
| Shield | DEX | Blocking with shield |
| Dodge | DEX | Evading attacks |

### Physical Skills

| Skill | Attribute | Use |
|-------|-----------|-----|
| Athletics | STR | Climbing, swimming, jumping |
| Stealth | DEX | Moving silently, hiding |
| Acrobatics | DEX | Balancing, tumbling |
| Riding | DEX | Horse/mount riding |

### Knowledge Skills

| Skill | Attribute | Use |
|-------|-----------|-----|
| Lore (Nature) | INT | Plants, animals, weather |
| Lore (History) | INT | Past events, kingdoms |
| Lore (Arcana) | INT | Magic, enchanted items |
| Lore (Religion) | INT | Gods, rituals |
| Lore (Local) | INT | Current events, personalities |

### Perception Skills

| Skill | Attribute | Use |
|-------|-----------|-----|
| Perception | WIS | General awareness |
| Tracking | WIS | Following trails |
| Insight | WIS | Reading intentions |
| Medicine | WIS | Healing, diagnosis |

### Social Skills

| Skill | Attribute | Use |
|-------|-----------|-----|
| Persuasion | CHA | Convincing others |
| Intimidation | CHA | Coercion |
| Deception | CHA | Lying, disguises |
| Performance | CHA | Music, acting |

### Crafting Skills

| Skill | Attribute | Use |
|-------|-----------|-----|
| Smithing | INT | Metalworking |
| Woodworking | INT | Carpentry, bows |
| Alchemy | INT | Potions, poisons |
| Cooking | INT | Food preparation |
| Tailoring | INT | Clothing, leather |

### Skill Resolution (3d6 Roll-Under)

```
func SkillCheck(skill int, modifiers int) bool {
    // Roll 3d6, succeed if total <= skill + modifiers
    roll := roll3d6()
    return roll <= (skill + modifiers)
}

// Critical success: roll <= 3 (natural 3-4)
// Critical failure: roll >= 17 (natural 17-18)
// Auto fail on 18 regardless of skill
```

## Saving Throws

| Save | Attribute | Used For |
|------|-----------|----------|
| Fortitude | CON | Poison, disease, physical endurance |
| Reflex | DEX | Area effects, traps, dodging |
| Will | WIS | Mental control, fear, illusions |

```
func FortitudeSave(entity Entity, dc int) bool {
    return SkillCheck(entity.Attributes.CON, dc)
}
```

## Hit Points & Vitality

```
type HP struct {
    Current int
    Max     int
}

type FatiguePoints struct {
    Current int
    Max     int
}
```

- **MaxHP** = CON × 2 + species bonus + size bonus
- **MaxFP** = CON + STR/2
- At 0 HP: unconscious
- Below 0 HP: dying (lose 1 HP/tick until saved or dead at -MaxHP)
- At 0 FP: exhaustion penalties (-2 to all actions)
- Below 0 FP: unconscious

## Entity Behavior & AI

```
type BehaviorState struct {
    CurrentGoal    Goal
    GoalQueue      []Goal
    Personality    PersonalityTraits
    Memory         map[string]interface{}
    LastDecision   uint64  // tick
}

type Goal struct {
    Type    GoalType  // travel, hunt, rest, socialize, flee, eat, work
    Target  string    // entity ID, location ID, item ID
    Priority int      // higher = more urgent
    Deadline uint64   // tick by which this should be completed
}

type PersonalityTraits struct {
    Aggression   int  // 0-100
    Curiosity    int  // 0-100
    Sociability  int  // 0-100
    Discipline   int  // 0-100
    Cowardice    int  // 0-100
}
```

### Behavior Priority

On each tick, entities evaluate their state and choose actions:

1. **Survival** — if HP < 25%, seek healing; if FP < 25%, rest; if starving, seek food
2. **Threat response** — if hostile entity in range, fight or flee
3. **Scheduled needs** — follow circadian rhythm (sleep, be active)
4. **Current goal** — progress toward active goal
5. **Idle** — wander, socialize, explore

```
func (e *Entity) Decide() Action {
    if e.HitPoints.Current < e.HitPoints.Max*0.25 {
        return FleeOrSeekHealing()
    }
    if e.FatiguePoints.Current < e.FatiguePoints.Max*0.25 && e.ShouldRest() {
        return Rest()
    }
    if threat := e.NearestThreat(); threat != nil {
        return e.HandleThreat(threat)
    }
    if e.Behavior.CurrentGoal != nil {
        return e.ProgressGoal()
    }
    return Idle()
}
```

### Goal Types

| Goal | Description |
|------|-------------|
| Travel | Move to a location |
| Hunt | Find and kill prey |
| Forage | Gather resources |
| Rest | Sleep or idle to recover FP |
| Socialize | Interact with same-species entities |
| Flee | Escape from threat |
| Patrol | Guard or wander territory |
| Work | Perform profession (mine, farm, craft) |
| Train | Improve skill |
| Explore | Wander into unknown locations |

## Entity Lifecycle

### Spawn

```
type SpawnConfig struct {
    Species     string
    LocationID  string
    Attributes  Attributes   // optional, random if nil
    Age         Age
    Faction     string
}
```

- Attributes generated based on species templates with variance
- Starting skills based on age and species
- Equipment based on profession/faction

### Growth & Aging

```
type Age struct {
    Years       int
    Mature      int   // age of adulthood
    MaxLifespan int
}
```

- Creatures gain skill XP through use
- Some species advance through life stages (larva → pupa → adult)
- Aging affects attributes (STR/CON decline, INT/WIS may increase)

### Death

- HP reaches -MaxHP: permanent death
- Special cases: undead resurrection, reincarnation
- Corpse remains as entity, can be looted

## Perception

Entities perceive their surroundings based on their senses.

```
type PerceptionResult struct {
    VisibleEntities []EntitySummary
    AudibleEvents   []Event
    LocationState   LocationSummary
}

func (e *Entity) Perceive(w *World) PerceptionResult {
    result := PerceptionResult{}
    location := w.GetLocation(e.LocationID)

    for _, other := range location.Entities {
        if CanSee(e, other) {
            result.VisibleEntities = append(result.VisibleEntities,
                other.Summary(e))
        }
    }
    return result
}
```

### Factors Affecting Perception

| Factor | Effect |
|--------|--------|
| Distance | -1 per 10 meters beyond base range |
| Lighting | dim = -2, dark = -4, blinded = impossible |
| Weather | rain = -2, fog = -4, blizzard = -6 |
| Cover | target behind cover = -2 to -6 |
| Stealth | opposed check: Perception vs Stealth |
| Species | elves +2 night vision, dwarves darkvision 60ft |

### Vision Types

| Type | Range | Notes |
|------|-------|-------|
| Normal | 60m (day), 15m (night) | Humans, halflings |
| Low-Light Vision | 60m (any light) | Elves, ignores dim light penalty |
| Darkvision | 30-60m (total dark) | Dwarves, orcs, goblins — see in black and white |
| Blindsight | 10-30m | Bats, blind creatures — sense without sight |
| Tremorsense | 30m | Underground creatures — sense vibrations |
