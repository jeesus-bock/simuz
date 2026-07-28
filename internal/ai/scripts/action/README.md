# Action AI Scripts

Scripts in this directory define general-purpose behavioral patterns for entities based on their role or disposition rather than faction, profession, or species. These scripts are shared across multiple entity types and govern core combat, movement, and survival behaviors.

---

## aggressive.lua

**Purpose:** Attacks hostile entities in the same location on sight and roams between adjacent locations when no target is available.

**Behavior:**
- Finds the nearest hostile entity and attacks immediately.
- Rate-limited per-faction so that different factions attack at different cadences.
- When no hostile is present, wanders between adjacent locations (camps, dens, sibling buildings).
- Prefers real exits over falling back to children of the parent location.

**Intended for:** Aggressive monsters, wild beasts, and any entity that should engage hostiles proactively without needing a specific profession or faction trigger.

---

## gathering.lua

**Purpose:** Collects resources from the environment while avoiding combat.

**Behavior:**
- Checks for hostile entities nearby before gathering.
- If a hostile is detected and HP is low, flees toward home.
- Periodically attempts to gather resources with a chance-based roll.
- Returns home periodically if away from the home location.
- Prefers outdoor, resource-rich locations.

**Intended for:** Peaceful NPCs and creatures that collect herbs, ore, or other materials — e.g., gatherers, herbalists, and non-combatant animals.

---

## hunting.lua

**Purpose:** Actively pursues and attacks prey when in range, tracking hostile entities across nearby locations.

**Behavior:**
- Scans for hostile entities at the current location and attacks if found.
- If no prey is at the current location, hunts across adjacent exits.
- Wanders toward random adjacent locations when no prey is found.
- Returns home at night/dusk.

**Intended for:** Predatory creatures and hunters that actively chase down targets across multiple locations — e.g., wolves, rangers, and predatory monsters.

---

## healing.lua

**Purpose:** Heals injured allies and creatures nearby, staying close to a sanctuary or home location.

**Behavior:**
- Finds the most injured non-hostile entity nearby (lowest HP percentage).
- Tends to the wounded entity (logs the action; actual healing is handled by `divine_intervention` for deity-type entities).
- Returns home at night/dusk.
- Periodically moves back toward home if away.

**Intended for:** Healer NPCs, priests, druids, and any support entity that should stay near allies and provide healing.

---

## scouting.lua

**Purpose:** Explores territory, reports findings, and flees when threatened. Avoids direct combat at all costs.

**Behavior:**
- Periodically explores adjacent exits to map territory.
- If a hostile is spotted, flees immediately if HP is low; otherwise stays alert and avoids contact.
- Returns home periodically.
- Prioritizes self-preservation over engagement.

**Intended for:** Scouting units, spies, and any entity that should gather intelligence rather than fight — e.g., scouts, spies, and timid creatures.

---

## goblin_ambush.lua

**Purpose:** A stealth-focused combat script for goblins and similar creatures that ambush prey from hiding.

**Behavior:**
- Hides before combat for a first-strike advantage.
- Gains bonus damage when attacking from a hidden state.
- Steals loot from killed targets.
- Flees when outnumbered or at low HP.
- Periodically moves between exits when no target is present.

**Intended for:** Goblinoids, bandit ambushers, and any creature that relies on stealth and surprise rather than open confrontation.
