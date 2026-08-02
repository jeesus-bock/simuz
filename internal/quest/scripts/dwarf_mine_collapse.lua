-- The Dwarf Mine Collapse
-- Dwarven miners are trapped underground after a collapse.

quest.define({
  id = "dwarf_mine_collapse",
  title = "The Dwarf Mine Collapse",
  type = "side",
  level = 2,
  description = "A mine collapse has trapped dwarven miners deep underground. Rescue them before the tunnels flood.",
  source = { type = "npc", npc_id = "dwarf_miner" },
  stages = {
    {
      id = "enter_mine",
      name = "Enter the Mine",
      description = "Descend into the collapsed mine and locate the trapped miners.",
      objectives = {
        { id = "visit_mine", type = "visit_location", description = "Reach the mine", location_id = "dwarf_mine" },
      },
    },
    {
      id = "rescue_miners",
      name = "Rescue the Miners",
      description = "Free the trapped dwarven miners from the rubble.",
      requirements = { "enter_mine" },
      objectives = {
        { id = "clear_rubble", type = "kill_entities", description = "Rubble cleared", count = 3, entity_template = "golem" },
      },
    },
    {
      id = "escort_miners",
      name = "Escort Miners to Safety",
      description = "Lead the rescued miners back to the surface.",
      requirements = { "rescue_miners" },
      objectives = {
        { id = "visit_surface", type = "visit_location", description = "Return to surface", location_id = "frosthold" },
      },
    },
  },
  rewards = { experience = 100, gold = 30 },
})