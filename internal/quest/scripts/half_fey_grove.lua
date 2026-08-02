-- The Half-Fey's Grove
-- A half-fey guardian needs help protecting a sacred grove.

quest.define({
  id = "half_fey_grove",
  title = "The Half-Fey's Grove",
  type = "side",
  level = 3,
  description = "A half-fey guardian seeks to protect a sacred grove from dark fey creatures.",
  source = { type = "npc", npc_id = "half_fey_guardian" },
  stages = {
    {
      id = "protect_grove",
      name = "Protect the Grove",
      description = "Defend the sacred grove from dark fey creatures.",
      objectives = {
        { id = "visit_grove", type = "visit_location", description = "Reach the grove", location_id = "crystal_forest" },
      },
    },
    {
      id = "banish_fey",
      name = "Banish the Dark Fey",
      description = "Defeat the dark fey creatures corrupting the grove.",
      requirements = { "protect_grove" },
      objectives = {
        { id = "kill_fey", type = "kill_entities", description = "Dark fey banished", count = 4, entity_template = "fairy" },
      },
    },
    {
      id = "bless_grove",
      name = "Bless the Grove",
      description = "Return to the half-fey guardian and report the grove is safe.",
      requirements = { "banish_fey" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to guardian", npc_id = "half_fey_guardian" },
      },
    },
  },
  rewards = { experience = 120, gold = 30 },
})