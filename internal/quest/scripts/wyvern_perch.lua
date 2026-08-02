-- The Wyvern's Perch
-- A wyvern has been poisoning the water supply near Frosthold.

quest.define({
  id = "wyvern_perch",
  title = "The Wyvern's Perch",
  type = "side",
  level = 4,
  description = "A venomous wyvern has taken roost on the cliffs above Frosthold, poisoning the water supply. End its threat.",
  source = { type = "npc", npc_id = "frosthold_priest" },
  stages = {
    {
      id = "find_perch",
      name = "Find the Wyvern's Perch",
      description = "Locate the wyvern's roost on the cliffs.",
      objectives = {
        { id = "visit_cliffs", type = "visit_location", description = "Reach the cliffs", location_id = "northern_highlands" },
      },
    },
    {
      id = "slay_wyvern",
      name = "Slay the Wyvern",
      description = "Defeat the venomous wyvern.",
      requirements = { "find_perch" },
      objectives = {
        { id = "kill_wyvern", type = "kill_entities", description = "Wyvern slain", count = 1, entity_template = "wyvern" },
      },
    },
    {
      id = "purify_water",
      name = "Purify the Water",
      description = "Return to the priest and help purify the water supply.",
      requirements = { "slay_wyvern" },
      objectives = {
        { id = "report_priest", type = "talk_to_npc", description = "Report to priest", npc_id = "frosthold_priest" },
      },
    },
  },
  rewards = { experience = 240, gold = 55 },
})