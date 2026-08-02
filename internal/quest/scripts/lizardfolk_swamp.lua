-- The Lizardfolk Swamp
-- Lizardfolk have poisoned the swamp waters near the Crystal Forest.

quest.define({
  id = "lizardfolk_swamp",
  title = "The Lizardfolk Swamp",
  type = "side",
  level = 3,
  description = "Lizardfolk have poisoned the swamp waters, threatening the Crystal Forest ecosystem. Cleanse the swamp.",
  source = { type = "npc", npc_id = "fairy_sparkle" },
  stages = {
    {
      id = "enter_swamp",
      name = "Enter the Swamp",
      description = "Brave the poisoned swamp and find the lizardfolk settlement.",
      objectives = {
        { id = "visit_swamp", type = "visit_location", description = "Reach the swamp", location_id = "crystal_forest" },
      },
    },
    {
      id = "purify_waters",
      name = "Purify the Waters",
      description = "Destroy the lizardfolk poison sacs contaminating the water.",
      requirements = { "enter_swamp" },
      objectives = {
        { id = "kill_lizardfolk", type = "kill_entities", description = "Lizardfolk defeated", count = 3, entity_template = "lizardfolk" },
      },
    },
    {
      id = "bless_forest",
      name = "Bless the Forest",
      description = "Return to the fairy and report the swamp is cleansed.",
      requirements = { "purify_waters" },
      objectives = {
        { id = "report_fairy", type = "talk_to_npc", description = "Report to fairy", npc_id = "fairy_sparkle" },
      },
    },
  },
  rewards = { experience = 130, gold = 30 },
})