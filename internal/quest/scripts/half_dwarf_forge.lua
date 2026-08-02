-- The Half-Dwarf's Forge
-- A half-dwarf blacksmith needs rare ore from the mines.

quest.define({
  id = "half_dwarf_forge",
  title = "The Half-Dwarf's Forge",
  type = "side",
  level = 2,
  description = "A half-dwarf blacksmith needs rare ore from the dwarf mines to forge a legendary blade.",
  source = { type = "npc", npc_id = "half_dwarf_blacksmith" },
  stages = {
    {
      id = "mine_ore",
      name = "Mine the Rare Ore",
      description = "Descend into the dwarf mine and extract the rare ore.",
      objectives = {
        { id = "visit_mine", type = "visit_location", description = "Reach the mine", location_id = "dwarf_mine" },
      },
    },
    {
      id = "collect_ore",
      name = "Collect the Ore",
      description = "Gather the rare ore from the mine.",
      requirements = { "mine_ore" },
      objectives = {
        { id = "gather_ore", type = "collect_items", description = "Rare ore collected", count = 3, location_id = "dwarf_mine" },
      },
    },
    {
      id = "deliver_ore",
      name = "Deliver to the Blacksmith",
      description = "Bring the ore to the half-dwarf blacksmith.",
      requirements = { "collect_ore" },
      objectives = {
        { id = "deliver", type = "deliver_item", description = "Deliver ore", npc_id = "half_dwarf_blacksmith" },
      },
    },
  },
  rewards = { experience = 80, gold = 25 },
})