# Combat System

## Design Philosophy

GURPS-inspired: statistics-heavy, tactical, simulationist. Combat is resolved per tick (1 second rounds) with opposed skill checks, hit locations, armor penetration, and detailed wound effects.

## Core Resolution Mechanic

All checks use **3d6 roll-under**: roll 3d6, succeed if the total is ≤ target number (skill level or attribute).

```
Natural 3-4: Critical success (maximize damage, extra effects)
Natural 17  : Critical failure (drop weapon, stumble)
Natural 18  : Automatic failure (and roll on critical miss table)
```

## Combat Round (1 Tick)

1. **Initiative** — entities act in DEX order (highest first)
2. **Action declaration** — choose Maneuver (see below)
3. **Resolution** — attack → defense → damage → effects
4. **End of tick** — status effect ticks, recovery

## Maneuvers

Each entity chooses one maneuver per tick:

| Maneuver | Description | Effects |
|----------|-------------|---------|
| Attack | Make one attack | Roll skill vs target's defense |
| All-Out Attack | Aggressive, no defense | +4 to hit OR +2 damage, no active defense |
| All-Out Defense | Full defensive | +2 to all active defenses, no attack |
| Move | Move up to full speed | Can also take a free action |
| Move & Attack | Move + attack at penalty | -4 to hit, skill capped at 9 |
| Aim | Prepare ranged attack | +Acc bonus per consecutive aimed tick |
| Ready | Draw item, reload, etc. | No attack possible |
| Feint | Fake-out to lower defense | Opposed skill contest, winner reduces foe's defense |
| Wait | Hold action for trigger | Interrupt when trigger condition met |
| Concentrate | Focus on spell/ability | Distraction possible if damaged |

## Attack Resolution

```
func ResolveAttack(attacker, defender Entity) AttackResult {
    // 1. Attacker rolls weapon skill
    attackRoll := Roll3d6()
    if attackRoll > attacker.Skill(weapon) && attackRoll != 3 {
        return Miss{}
    }

    // 2. Defender chooses active defense
    defense := defender.ChooseDefense(attackType)
    defenseRoll := Roll3d6()
    hit := defenseRoll > defense.Skill  // defender fails

    if !hit {
        return Blocked{defenseType}
    }

    // 3. Hit location
    location := RollHitLocation(aimed)

    // 4. Damage
    rawDamage := RollDamage(attacker, weapon)
    dr := defender.Armor.At(location)
    penetrating := max(0, rawDamage - dr)
    finalDamage := ApplyWoundingModifier(penetrating, weapon.DamageType)
    woundEffects := ApplyDamage(defender, location, finalDamage)

    return Hit{location, finalDamage, woundEffects}
}
```

## Hit Locations

```
func RollHitLocation(aimed TargetLocation) HitLocation {
    if aimed != None {
        // Called shot: roll against -penalty
        // If successful, hit aimed location
        return aimed
    }
    // Random location
    roll := Roll3d6()
    switch {
    case roll <= 4:   return Brain     // -7 to hit
    case roll <= 5:   return Head      // -5
    case roll == 6:   return RightArm  // -2 or -4 (far arm)
    case roll == 7:   return RightHand // -4
    case roll <= 9:   return Torso
    case roll == 10:  return Groin     // -3
    case roll <= 12:  return Torso
    case roll == 13:  return LeftArm   // -2 or -4 (far arm)
    case roll <= 14:  return LeftLeg   // -2
    case roll <= 15:  return LeftHand  // -4
    case roll >= 16:  return Vitals    // -3
    }
}
```

| Location | Penalty to Hit | Damage Modifier | Special |
|----------|---------------|-----------------|---------|
| Torso | 0 | ×1.0 | Default hit location |
| Head | -5 | ×2.0 (imp/pi) / ×1.5 (cut) | Knockdown, stun, possible unconsciousness |
| Brain | -7 | ×4.0 | Instant KO/death on penetration |
| Vitals | -3 | ×3.0 | Heart/lungs — shock, bleeding |
| Arm | -2 (near) / -4 (far) | ×1.0 | Crippled if > HP/2 damage |
| Hand | -4 | ×1.0 | Crippled easily, may drop item |
| Leg | -2 | ×1.0 | Crippled if > HP/2, knockdown |
| Foot | -4 | ×1.0 | Crippled, knockdown |
| Groin | -3 | ×2.0 (imp) | Shock, knockdown |
| Face | -5 | ×1.5 | Blindness, knockout |

## Active Defenses

Defender chooses one defense per attack (only if aware of the attack):

| Defense | Base Skill | Modifiers |
|---------|-----------|-----------|
| Dodge | DEX/2 + 3 | -encumbrance penalty |
| Parry | (Weapon Skill)/2 + 3 | -2 per parry after first, unbalanced weapon unready |
| Block | (Shield Skill)/2 + 3 | -2 per block after first |

```
func (e *Entity) ParrySkill() int {
    return e.Skills["Melee"]/2 + 3
}

func (e *Entity) BlockSkill() int {
    return e.Skills["Shield"]/2 + 3
}

func (e *Entity) DodgeSkill() int {
    return e.Attributes.Mod(e.Attributes.DEX)/2 + 3 + e.Equipped.DodgeBonus
}
```

### Defense Modifiers

| Situation | Modifier |
|-----------|----------|
| Retreat (step back) | +3 to all defenses |
| Feverish Defense | +2 (costs 1 FP) |
| All-Out Defense | +2 |
| Stunned | -4 |
| Prone | -3 |
| Encumbered | -1 per level |
| Multiple attacks in same tick | -2 per defense after first |

## Damage

### Damage Roll

Weapon damage is expressed as dice plus modifier: e.g., `2d+1` for a broadsword.

```
func RollDamage(strength int, weapon Weapon) int {
    // Base damage from strength (thrust vs swing)
    thrust, swing := StrengthDamageTable(strength)
    var base int
    if weapon.Usage == Swing {
        base = swing
    } else {
        base = thrust
    }
    // Add weapon modifier
    return RollDice(base + weapon.DiceMod)
}
```

### Strength Damage Table

| STR | Thrust | Swing | STR | Thrust | Swing |
|-----|--------|-------|-----|--------|-------|
| 1-2 | 1d-6 | 1d-5 | 13 | 1d | 2d-1 |
| 3-4 | 1d-5 | 1d-4 | 14 | 1d | 2d |
| 5-6 | 1d-4 | 1d-3 | 15 | 1d+1 | 2d+1 |
| 7-8 | 1d-3 | 1d-2 | 16 | 1d+1 | 2d+2 |
| 9 | 1d-2 | 1d-1 | 17 | 1d+2 | 3d-1 |
| 10 | 1d-2 | 1d | 18 | 1d+2 | 3d |
| 11 | 1d-1 | 1d+1 | 19 | 2d-1 | 3d+1 |
| 12 | 1d-1 | 1d+2 | 20 | 2d-1 | 3d+2 |

### Damage Types & Wounding Modifiers

After subtracting DR, the remaining penetrating damage is multiplied by the wounding modifier:

| Type | Abbr | Wound Mod | Examples |
|------|------|-----------|---------|
| Crushing | cr | ×1.0 | Club, fist, mace, falling |
| Cutting | cut | ×1.5 | Sword, axe, claws |
| Impaling | imp | ×2.0 | Spear, arrow, dagger stab |
| Small Piercing | pi- | ×0.5 | Small arrows, bolts |
| Piercing | pi | ×1.0 | Most firearms |
| Large Piercing | pi+ | ×1.5 | Rifles, heavy bows |
| Huge Piercing | pi++ | ×2.0 | Anti-material rifles |
| Burning | burn | ×1.0 | Fire, acid, magic |

### Armor & Damage Resistance

```
type ArmorSet struct {
    Head    int     // DR
    Torso   int
    Arms    int
    Legs    int
    Hands   int
    Feet    int
    Shield  int     // extra DR when blocking
}

func (a ArmorSet) At(location HitLocation) int {
    // Returns DR for the specified hit location
}
```

Armor reduces damage before wounding modifiers:
```
penetrating = max(1, rawDamage - dr)  // minimum 1 if DR exceeded

// If damage <= DR and damage type is crushing:
bluntTrauma = max(0, rawDamage - dr/2) / 2  // bruising through armor
```

### Blunt Trauma

If a crushing attack fails to penetrate DR, or a flexible armor stops a cutting/impaling blow, the target still takes blunt trauma:
- Flexible armor (leather, chain): 1 HP per 5 points of rolled damage
- Rigid armor (plate): no blunt trauma unless crushing attack

## Wound Effects

### Shock

On any hit that deals damage:
- Next tick only: -damage dealt to DX and IQ (minimum -0, maximum -4)
- Does not affect active defenses

### Major Wound

A single injury > 1/2 of max HP:
- Knockdown: roll HT or fall prone
- Stun: roll HT or be stunned (no actions, -4 defenses)
- Crippling: if hitting a limb, it's crippled (cannot use)

### Crippling

- Limb injury > HP/2: crippled
- Hand/foot injury > HP/3: crippled
- Crippled limb: unusable (drop weapon, fall, etc.)
- Roll HT: success = temporary (heal when HP restored), failure = lasting (1d months)

### Bleeding

- Any cutting, impaling, or piercing wound deals 1 HP/tick bleed damage
- Multiple wounds stack
- First Aid or medical treatment stops bleeding (Medicine skill check)

### Unconsciousness & Death

| State | Condition |
|-------|-----------|
| Conscious | HP > 0 |
| Disabled | HP = 0, can take one action then pass out |
| Dying | HP < 0, lose 1 HP/tick, roll HT each tick |
| Dead | HP < -MaxHP (or HP ≤ -5×CON, whichever is more) |

## Healing

| Method | Rate |
|--------|------|
| Natural rest | 1 HP per 8 hours of rest |
| Natural (active) | 1 HP per 24 hours |
| First aid | +1d HP once per wound (successful Medicine roll) |
| Magical | Varies by spell |
| Crippled limb (temporary) | Heals when HP restored above cripple threshold |
| Crippled limb (lasting) | 1d months |
| Fatigue recovery | All FP after 10 minutes rest, 1 FP per 2 minutes |
