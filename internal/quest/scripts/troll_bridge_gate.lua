-- The Troll Bridge Gate
-- A troll blocks the bridge to the Northern Highlands.

quest.define({
  id = "troll_bridge_gate",
  title = "The Troll Bridge Gate",
  type = "side",
  level = 4,
  description = "A massive troll has taken up residence under the bridge to the Northern Highlands. Slay it to reopen the pass.",
  source = { type = "npc", npc_id = "frosthold_guard_captain" },
  stages = {
    {
      id = "reach_bridge",
      name = "Reach the Bridge",
      description = "Travel to the bridge gate in the Northern Highlands.",
      objectives = {
        { id = "visit_bridge", type = "visit_location", description = "Reach the bridge", location_id = "northern_highlands" },
      },
    },
    {
      id = "slay_troll",
      name = "Slay the Troll",
      description = "Defeat the troll blocking the bridge.",
      requirements = { "reach_bridge" },
      objectives = {
        { id = "kill_troll", type = "kill_entities", description = "Troll slain", count = 1, entity_template = "troll" },
      },
    },
    {
      id = "clear_pass",
      name = "Clear the Pass",
      description = "Return to Frosthold and confirm the pass is safe.",
      requirements = { "slay_troll" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to captain", npc_id = "frosthold_guard_captain" },
      },
    },
  },
  rewards = { experience = 200, gold = 50 },
})