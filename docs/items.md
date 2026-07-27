# Items & Equipment

## Item Model

```
type Item struct {
    ID          string
    Name        string
    Type        ItemType
    Weight      float64    // kg
    Value       int        // base value in gold/silver
    Description string
    Properties  map[string]interface{}
}

type ItemType int
const (
    Weapon ItemType = iota
    Armor
    Shield
    Container
    Consumable
    Tool
    Material
    Treasure
    Quest
)
```

## Weapons

```
type Weapon struct {
    Item
    DamageDice      string      // "2d+1"
    DamageType      DamageType  // cr, cut, imp, pi-, pi, pi+, pi++
    Usage           WeaponUsage // Thrust, Swing
    StrengthMin     int         // minimum STR to wield effectively
    Reach           int         // 1 = close, 2 = medium, 3+ = long (meters)
    ParryBonus      int         // bonus to parry (usually 0 or +1)
    TwoHanded       bool
    Acc             int         // accuracy bonus for aiming (ranged)
    Range           [2]int      // half/max range (ranged)
    RoF             int         // rate of fire (shots per tick)
    Bulk            int         // penalty when moving and shooting
}

enum Usage {
    Thrust  // stab, poke — uses thrust damage from STR
    Swing   // slash, chop — uses swing damage from STR
}
```

### Melee Weapons Table

| Weapon | Damage | Type | ST Min | Reach | Parry | Weight | Notes |
|--------|--------|------|--------|-------|-------|--------|-------|
| Broadsword | sw+1 / thr+2 | cut/imp | 10 | 1,2 | +1 | 2 kg | Versatile |
| Shortsword | sw / thr+1 | cut/imp | 8 | 1 | 0 | 1.5 kg | |
| Greatsword | sw+3 / thr+3 | cut/imp | 12 | 2,3 | 0 | 5 kg | Two-handed |
| Dagger | sw-3 / thr-1 | cut/imp | 5 | C,1 | -1 | 0.5 kg | Concealable |
| Spear | thr+3 | imp | 9 | 1,2* | 0 | 2 kg | Can be thrown |
| Battleaxe | sw+2 | cut | 11 | 1 | 0 | 3 kg | Unbalanced |
| Warhammer | sw+3 | cr | 12 | 1 | 0 | 3 kg | Unbalanced |
| Mace | sw+2 | cr | 10 | 1 | 0 | 2.5 kg | |
| Club | sw+1 | cr | 7 | 1 | 0 | 1 kg | Improvised |
| Quarterstaff | sw+2 / thr+2 | cr/cr | 7 | 2,3 | +2 | 2 kg | Two-handed |
| Handaxe | sw | cut | 8 | 1 | 0 | 1 kg | Thrown |
| Flail | sw+2 | cr | 11 | 1 | -2 | 3 kg | Ignores shield DB |

### Ranged Weapons Table

| Weapon | Damage | Type | ST Min | Acc | Range | RoF | Bulk | Weight | Notes |
|--------|--------|------|--------|-----|-------|-----|------|--------|-------|
| Shortbow | thr+1 | imp | 7 | 1 | 50/100 | 1 | -4 | 1.5 kg | |
| Longbow | thr+2 | imp | 11 | 2 | 100/150 | 1 | -6 | 2 kg | |
| Crossbow | thr+3 | imp | 7 | 3 | 100/120 | 1/3 | -6 | 4 kg | 1 shot/3 ticks |
| Sling | sw | cr | 5 | 0 | 30/60 | 1 | -3 | 0 kg | Ammo = stones |
| Javelin | thr+1 | imp | 7 | 1 | 30/50 | 1 | -4 | 1 kg | Thrown |
| Throwing Knife | thr-1 | imp | 5 | 0 | 10/20 | 1 | -2 | 0.5 kg | |

## Armor

```
type Armor struct {
    Item
    Location   ArmorLocation  // which body part(s) it covers
    DR         int            // damage resistance
    Weight     float64
    Penalty    ArmorPenalty   // DX penalty, dodge penalty, move penalty
    Flexible   bool           // true = leather, chain; false = plate
}
```

### Armor Table

| Armor | DR | Locations | Weight | Penalty | Flexible | Notes |
|-------|----|-----------|--------|---------|----------|-------|
| Leather Jack | 1 | Torso, Arms | 4 kg | -1 dodge | Yes | |
| Leather Pants | 1 | Legs | 2 kg | - | Yes | |
| Studded Leather | 2 | Torso, Arms | 6 kg | -1 dodge | Yes | |
| Chainmail | 4 | Torso | 10 kg | -1 dodge, -1 move | Yes | |
| Chain Coif | 4 | Head | 2 kg | - | Yes | |
| Chain Leggings | 4 | Legs | 5 kg | -1 move | Yes | |
| Scale Mail | 4 | Torso, Arms | 12 kg | -2 dodge, -1 move | No | |
| Plate Breastplate | 5 | Torso | 8 kg | -1 dodge | No | |
| Plate Helm | 5 | Head | 3 kg | - | No | -1 perception |
| Plate Arms | 5 | Arms | 4 kg | -1 dodge, -1 DX | No | |
| Plate Legs | 5 | Legs | 6 kg | -1 dodge, -1 move | No | |
| Full Plate | 5 | All | 20 kg | -2 dodge, -1 move, -1 DX | No | |
| Shield (Small) | 2 | Block DR | 3 kg | - | Yes | +2 block skill |
| Shield (Large) | 3 | Block DR | 6 kg | -1 dodge | Yes | +3 block skill, cover |

### Encumbrance

| Level | Weight Carried | Dodge Penalty | Move Penalty | Fatigue Cost |
|-------|---------------|---------------|--------------|-------------|
| None | ≤ STR×5 | 0 | 0 | 0 |
| Light | STR×5 to STR×10 | -1 | -1 | ×1.5 |
| Medium | STR×10 to STR×15 | -2 | -2 | ×2 |
| Heavy | STR×15 to STR×20 | -3 | -3 | ×3 |
| Overload | > STR×20 | Cannot move | - | - |

## Consumables

```
type Consumable struct {
    Item
    EffectType   EffectType  // heal, buff, cure_poison, etc.
    Potency      int
    Uses         int         // charges before consumed
}
```

| Item | Effect | Weight | Value |
|------|--------|--------|-------|
| Healing Herb | +1d HP over 1 hour | 0.1 kg | 5 silver |
| Healing Potion | +2d HP instantly | 0.5 kg | 50 gold |
| Antidote | Cure poison | 0.5 kg | 30 gold |
| Stamina Draft | +1d FP instantly | 0.5 kg | 20 gold |
| Bandages | +1d HP (first aid) | 0.2 kg | 2 silver |

## Tools

| Tool | Skill | Use | Weight |
|------|-------|-----|--------|
| Smith's Hammer | Smithing | Craft metal items | 3 kg |
| Woodworking Tools | Woodworking | Craft wood items | 2 kg |
| Alchemy Kit | Alchemy | Brew potions | 5 kg |
| Lockpicks | Stealth | Open locks | 0.2 kg |
| Rope (10m) | Athletics | Climbing, binding | 1 kg |

## Crafting

```
type Recipe struct {
    Name        string
    Skill       string
    Difficulty  int        // target number for skill check
    Materials   []Material // required inputs
    Output      string     // item ID to produce
    Time        int        // ticks to craft
    Tools       []string   // required tools
}

type Material struct {
    ItemID string
    Count  int
}
```

### Example Recipes

| Recipe | Skill | Materials | Time |
|--------|-------|-----------|------|
| Forge Broadsword | Smithing 12 | 2 × Iron Ingot, 1 × Leather Grip | 3600 ticks (1h) |
| Craft Leather Armor | Smithing 10 | 3 × Cured Hide, Thread | 1800 ticks (30m) |
| Brew Healing Potion | Alchemy 14 | 2 × Healing Herb, 1 × Glass Vial, Water | 600 ticks (10m) |
| Fletch Arrows (20) | Woodworking 8 | 1 × Wood, Feathers, 20 × Arrowhead | 300 ticks (5m) |
