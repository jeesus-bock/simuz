-- The Ettin's Two Heads
-- An ettin has been terrorizing the Northern Highlands.

quest.define({
  id = "ettin_two_heads",
  title = "The Ettin's Two Heads",
  type = "side",
  level = 5,
  description = "A two-headed ettin has been wreaking havoc in the Northern Highlands. Defeat both heads.",
  source = { type = "npc", npc_id = "frosthold_guard_captain" },
  stages = {
    {
      id = "find_ettin",
      name = "Find the Ettin",
      description = "Track the ettin to its lair in the highlands.",
      objectives = {
        { id = "visit_highlands", type = "visit_location", description = "Reach the highlands", location_id = "northern_highlands" },
      },
    },
    {
      id = "slay_ettin",
      name = "Slay the Ettin",
      description = "Defeat the two-headed ettin.",
      requirements = { "find_ettin" },
      objectives = {
        { id = "kill_ettin", type = "kill_entities", description = "Ettin slain", count = 1, entity_template = "ettin" },
      },
    },
    {
      id = "return_victor",
      name = "Return Victorious",
      description = "Return to Frosthold and report the ettin's defeat.",
      requirements = { "slay_ettin" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to captain", npc_id = "frosthold_guard_captain" },
      },
    },
  },
  rewards = { experience = 300, gold = 75 },
})