# Profession AI Scripts

Scripts in this directory define profession-specific behaviors for NPCs. These scripts govern the day-to-day activities of NPCs based on their occupation — trading, crafting, serving, farming, and more. Each script is intended for entities that have a specific profession role.

---

## blacksmith.lua

**Purpose:** Works at the forge during the day, crafts arms and armour from raw materials, sells finished goods, and buys ore/coal from miners.

**Behavior:**
- Crafts items using available recipes (smelting iron, crafting bandages).
- Buys raw materials (iron ore, coal, cloth, leather) from nearby merchants or miners.
- Sells finished wares (swords, armour, helmets, boots, shields, bandages) to civilians and guards.
- Restocks materials periodically if supply runs low.
- Returns to the smithy at night.

**Intended for:** Blacksmith NPCs who run a forge or workshop.

---

## bard.lua

**Purpose:** Traveling musician who performs in taverns for tips, inspires allies, and calms hostiles during performances.

**Behavior:**
- Moves between tavern locations on a set route.
- Performs during the day for tips (copper, silver, drinks).
- Performance quality tiers (lousy, mediocre, great) affect earnings and inspiration chance.
- Can inspire nearby civilians and calm hostile entities during performances.
- Occasionally drinks for inspiration (more likely at higher quality tiers).
- Returns to a tavern at night.

**Intended for:** Bard NPCs who travel between towns performing for coin.

---

## bard_patron.lua

**Purpose:** A regular tavern-goer who listens to bards, buys drinks, and occasionally tips performers.

**Behavior:**
- Sits in the common room during the evening.
- Buys drinks from the innkeeper.
- Occasionally tips the bard (copper pieces).
- Sips drinks for enjoyment.
- Returns home at night.

**Intended for:** Tavern patron NPCs who are audience members for bards.

---

## courier.lua

**Purpose:** A civilian messenger who roams town buildings, flees danger, and cooperates with rescue/leash behavior.

**Behavior:**
- Wanders between town buildings during the day.
- Flees immediately when hostiles are nearby.
- Returns home at dusk/night.
- Can be leashed or rescued by guards.
- Avoids combat at all costs.

**Intended for:** Courier NPCs who deliver messages between towns and buildings.

---

## dog.lua

**Purpose:** A farm or town companion that stays near home, follows leashes cleanly, and wanders a little during the day.

**Behavior:**
- Stays near home location.
- Follows leash commands from owners (farmers).
- Wanders nearby during the day.
- Returns home at night.
- Never engages in combat.

**Intended for:** Dog NPCs that serve as farm companions or town pets.

---

## farmer.lua

**Purpose:** Manages livestock, tends crops, sells produce at market, and returns home at night.

**Behavior:**
- Tends animals during the day (feeding, slaughtering old livestock).
- Slaughters animals that have reached their age threshold for meat, leather, wool, feathers.
- Leashes a farm dog for market trips.
- Sells animal products (meat, eggs, milk, wool, leather, grain) at the market.
- Buys bait or supplies if needed.
- Returns home at night.

**Intended for:** Farmer NPCs who manage livestock and sell produce.

---

## fisherman.lua

**Purpose:** Lives by the pond, fishes during the day, sells catch at the market, buys bait, and sleeps at night.

**Behavior:**
- Fishes at the pond during the day (trout, salmon, catfish).
- Sells fish at the market to merchants and civilians.
- Buys bait when inventory is low.
- Weather affects fishing (storms block fishing).
- Returns to the pond at night/dusk.

**Intended for:** Fisherman NPCs who live by a pond or river.

---

## herbalist.lua

**Purpose:** Gathers herbs in the wild or garden during the day, crafts poultices and salves, sells remedies, and returns home at night.

**Behavior:**
- Gathers herbs periodically from the environment.
- Crafts remedies (herbal poultices, healing salves, bandages) at a campfire or cauldron.
- Sells remedies to wounded civilians, priests, and merchants.
- Returns home at night/dusk.

**Intended for:** Herbalist NPCs who gather and craft healing items.

---

## innkeeper.lua

**Purpose:** Tends the common room, serves drinks to patrons, buys tankards from salesmen, and keeps a stock of ale, wine, and liquor.

**Behavior:**
- Serves drinks to patrons (beer, ale, wine, liquor, mead, brandy).
- Buys tankards from traveling salesmen.
- Restocks drinks when inventory runs low.
- Cleans the bar periodically.
- Closes the inn at night/dusk and returns home.

**Intended for:** Innkeeper NPCs who run a tavern or inn.

---

## miner.lua

**Purpose:** Digs ore and coal at the mine during the day, sells raw materials to blacksmiths, and returns to town at night.

**Behavior:**
- Mines iron ore and coal at the mine during the day.
- Sells raw materials to blacksmiths and merchants.
- Drinks beer/ale after work in the evening.
- Returns home at night.

**Intended for:** Miner NPCs who work at a mine.

---

## priest.lua

**Purpose:** Heals wounded civilians, returns to home temple at dusk, and wanders to nearby buildings during the day.

**Behavior:**
- Heals wounded civilians using divine intervention.
- Wanders to nearby buildings during the day.
- Returns to home temple at dusk/night.
- Occasionally drinks wine or mead for divine communion.
- Mood shifts based on time of day (prayerful at night, serene during day).

**Intended for:** Priest NPCs who serve at a temple and heal the wounded.

---

## ranger.lua

**Purpose:** A wilderness hunter who tracks and kills hostile beasts, gathers pelts and meat, sells them, and camps at night.

**Behavior:**
- Hunts hostile beasts (beasts, vermin) in the wild.
- Gathers pelts, meat, leather, feathers from slain prey.
- Sells goods to merchants and civilians.
- Makes camp at home at night.
- Stays in forests and fields during the day.

**Intended for:** Ranger NPCs who hunt wildlife and gather resources in the wilderness.

---

## ranger_profession.lua

**Purpose:** Profession script for rangers — hunts prey, tracks targets via relationships, gathers resources, and avoids unnecessary combat.

**Behavior:**
- Tracks prey and hostiles via relationship system (rival bonds).
- Hunts prey and gathers resources when no threats are present.
- Defends against hostile entities (bandits, aggressive factions).
- Avoids unnecessary combat.
- Demonstrates profession-based behavior and relationship tracking.

**Intended for:** Ranger NPCs who need profession-specific logic separate from the species-based ranger script.

---

## thief.lua

**Purpose:** Hides in the shadows, pickpockets civilians and merchants, sells stolen goods, and flees when threatened or when guards are near.

**Behavior:**
- Pickpockets valuables from civilians, merchants, and guards.
- Sells stolen goods to fences (merchants, bandits, civilians).
- Flees when HP is low or when guards are nearby.
- Hides in shadows and avoids direct confrontation.
- Moves between locations to avoid detection.

**Intended for:** Thief NPCs who engage in stealthy criminal activities.

---

## traveling_salesman.lua

**Purpose:** Wanders between towns, trades goods, haggles with creatures, sleeps at inns at night, and trades during the day.

**Behavior:**
- Follows a route between towns and buildings.
- Sells wares (clothes, weapons, armour, shields) to civilians and merchants.
- Buys goods (tankards, holy symbols) from merchants.
- Trades during the day, sleeps at inns at night.
- Haggles with creatures for better deals.

**Intended for:** Traveling salesman NPCs who move between towns trading goods.
