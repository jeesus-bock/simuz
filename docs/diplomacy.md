# Diplomacy & Leaders

## Overview

The diplomacy system governs how political leaders (politicians) and diplomatic
agents (diplomats) interact across civilized species. It is built on three
pillars:

1. **Diplomats** — neutral envoys protected by diplomatic immunity.
2. **Politicians** — species-specific leaders who control immunity and conduct
   statecraft.
3. **Diplomatic immunity** — a Relation-based protection mechanism that prevents
   diplomats from being attacked unless a politician revokes it.

---

## Diplomat

**Profession:** `diplomat`
**Script:** `internal/ai/scripts/profession/diplomat.lua`

Diplomats are itinerant envoys who travel between locations seeking political
engagement. They carry diplomatic immunity — no entity attacks them by default.

### Behavior (every 20 ticks)

| Priority | Action | Detail |
|----------|--------|--------|
| 1 | Seek politicians | Scans nearby entities for `profession == "politician"`, then engages via `world.say_to()` |
| 2 | General statecraft | Finds strongest/weakest non-diplomat entities; flatters the strong, destabilizes the weak |
| 3 | Travel | If no targets, moves to a random exit |

### Diplomatic Immunity

Diplomats pass the Relation hostility test. The engine checks
`hasDiplomaticImmunity()` in `world.attack()`, `world.defend_self()`, and the
Go-side aggressive/hunting/defendPassiveSelf combat paths.

**Immunity holds** when:
- The attacker has no negative entity-level relation toward the diplomat
- No nearby same-faction politician has declared the diplomat hostile

**Immunity is bypassed** when:
- A politician calls `world.set_entity_relation(diplomat_id, "hostile")`
- The attacker or a nearby same-faction politician has a negative
  `EntityRelation` toward the diplomat

### Inventory (spawned)

Fine clothes, dagger, herb pouch, 10-30 cp / 5-15 sp / 1-4 gp.

---

## Politician

**Profession:** `politician`
**Script:** `internal/ai/scripts/profession/politician.lua`

Politicians are species-specific political leaders. They interact with diplomats,
manage faction relations, and control diplomatic immunity. Only civilized species
(human, orc, dwarf, elf) can be politicians.

### Species Variants

#### Human Councilor

| Trait | Value |
|-------|-------|
| Demeanor | Pragmatic, alliance-seeking, wealth-driven |
| Diplomat interaction | Holds court via `world.say_to()`, seeks shared language |
| Rival interaction | Eyes rival politicians with calculated suspicion |
| Foreigner interaction | Opens trade negotiations with non-faction entities |
| Mood | authoritative, scheming, diplomatic |

#### Orc Warlord

| Trait | Value |
|-------|-------|
| Demeanor | Strength-based, tribute-demanding, volatile |
| Diplomat interaction | 25% kill (revokes immunity + attacks), 35% tribute demand, 40% reluctant audience |
| Rival interaction | Chest-pounding territorial display |
| Guard interaction | Orders guards to patrol |
| Mood | furious, dominant, dismissive, aggressive |

The orc warlord is the only politician who can revoke diplomatic immunity. When
it decides to kill (25% roll), it calls:
```lua
world.set_entity_relation(dip.id, "hostile")  -- revokes immunity
world.attack(self.id, dip.id)                  -- attacks
```

#### Dwarf Thane

| Trait | Value |
|-------|-------|
| Demeanor | Trade-focused, conservative, clan-bound |
| Diplomat interaction | Grants formal audience if can communicate; waves off with "I dinnae understand yer jabber" otherwise |
| Foreigner interaction | "Got coin? Then we talk. No coin, no deal." |
| Mood | measured, gruff, transactional |

#### Elf Archon

| Trait | Value |
|-------|-------|
| Demeanor | Long-view strategist, nature-aligned, proud |
| Diplomat interaction | Disdainful of orc/goblin diplomats ("Your kind scars the land"); measured discourse with others |
| Rival interaction | Observes with ancient patience |
| Mood | aloof, serene, contemplative |

### Spawning

| Rule ID | Location | Species | Level | Notes |
|---------|----------|---------|-------|-------|
| `human_politician` | tavern | human | 3-5 | Initial spawn only |
| `orc_chief` | orc_camp | orc | 4-6 | FactionID: orc |
| `dwarf_thane` | dwarf_keep | dwarf | 4-6 | Fallback location created in world gen |
| `elf_archon` | fey_glade | elf | 5-7 | Initial spawn only |

### Behavior (every 15 ticks)

1. Dispatch to species-specific handler based on `self.species`
2. If no action taken, move to a random exit to inspect the domain

---

## Bandit Chief Interaction

**Script:** `internal/ai/scripts/profession/bandit_chief.lua`

Bandit chiefs target merchants, couriers, and travelers as primary targets. They
also track diplomats as secondary targets with a 30% attack chance per tick.

Because bandit chiefs call `world.attack()` directly, the diplomatic immunity
check in the engine blocks the attack unless a politician has revoked immunity.
This creates a dynamic where bandit attacks on diplomats fail by default, but
succeed after an orc warlord (or another politician) declares the diplomat
hostile.

---

## Lua API

### `world.set_entity_relation(target_id, relation)`

Sets entity-level hostility on the caller's Relation struct.

| Parameter | Type | Values |
|-----------|------|--------|
| `target_id` | string | Entity ID |
| `relation` | string | `"hostile"` (-10), `"friendly"` (10), `"neutral"` (0) |

Used by politicians to revoke diplomatic immunity:
```lua
world.set_entity_relation(diplomat_id, "hostile")
```

### `world.get_entity_relation(target_id)`

Returns the combined Relation score between the caller and target as a string.

| Returns | Condition |
|---------|-----------|
| `"hostile"` | combined < 0 |
| `"friendly"` | combined > 0 |
| `"neutral"` | combined == 0 |

### `world.say_to(target_id, message)`

Sends a speech message to a target entity. Used for diplomatic discourse.

### `world.can_communicate(target_id)`

Returns true if the caller and target share at least one language at proficiency
>= 1.

### `world.best_shared_language(target_id)`

Returns the language ID with the highest combined proficiency between the caller
and target.

---

## Relation System

The Relation system (`internal/relation/relation.go`) calculates hostility as a
combined score from four sources:

| Source | Weight | Example |
|--------|--------|---------|
| EntityRelation | Per-entity | Diplomat immunity override |
| FactionRelation | Per-faction | Orc faction vs civilian |
| SpeciesRelation | Per-species | Orc vs elf (-100 default) |
| ProfessionRelation | Per-profession | Bandit vs guard (-50) |

A combined score < 0 means hostile (attack allowed). Diplomats are protected
because their `ProfessionRelation` is neutral (0) for all professions, and no
entity/faction/species relation targets them specifically.

---

## Go-Side Protection

Diplomatic immunity is enforced in three combat paths in `internal/engine/tick.go`:

1. **Aggressive archetype** (line ~460): skips diplomats where
   `ent.Relation.GetEntityRelation(other.ID) >= 0`
2. **Hunting archetype** (line ~496): same check
3. **defendPassiveSelf** (line ~1053): skips diplomats where
   `ent.Relation.GetEntityRelation(other.ID) >= 0`

And in `internal/ai/scripted.go`:

4. **world.attack()** (line ~1522): calls `hasDiplomaticImmunity()` before
   proceeding with combat
5. **world.defend_self()** (line ~1487): calls `hasDiplomaticImmunity()` before
   adding to hostiles list

---

## Enhancement Opportunities

The current system is functional but minimal. Areas for expansion:

### Faction-level diplomacy
- Politicians could negotiate treaties between factions using
  `world.set_relation(factionA, factionB, "friendly")`
- Track treaty state in faction `CurrentState` / `PrimaryObjective` fields
- Alliances, trade agreements, non-aggression pacts

### Diplomatic events
- Formal summits where multiple politicians and diplomats gather
- Treaty signing ceremonies that emit `SimEvent`s
- Embassy establishment: diplomats could claim a location as their embassy

### Reputation system
- Track diplomatic success/failure per entity
- Reputation affects CHA-based checks and willingness to negotiate
- Diplomats who survive orc warlord encounters gain prestige

### Espionage
- Diplomats could gather intelligence on faction troop counts, wealth, plans
- Politicians could order diplomats to spy on rival factions
- Discovery of espionage could trigger immunity revocation

### War declarations
- Politicians could formally declare war between factions
- War state changes all faction members' Relation toward the enemy
- Ceasefire negotiations through diplomats

### Species-specific diplomatic protocols
- Dwarf thanes require written contracts (item-based treaties)
- Elf archons negotiate in decades-long timelines
- Orc warlords respect only strength demonstrations (combat trials)
- Human councils use vote-based decision making

### Economic diplomacy
- Trade route establishment between settlements
- Tariff negotiation affecting item prices
- Resource sharing agreements (food, weapons, materials)

### Succession and coups
- Politician death triggers succession events
- Rival politicians can attempt coups
- Diplomatic fallout from leadership changes
