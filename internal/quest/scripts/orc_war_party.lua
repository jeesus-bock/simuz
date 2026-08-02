-- The Orc War Party
-- An orc war band threatens the northern trade routes.

quest.define({
  id = "orc_war_party",
  title = "The Orc War Party",
  type = "side",
  level = 3,
  description = "A war party of orcs has been raiding the northern trade routes. Eliminate their leaders to restore safe passage.",
  source = { type = "npc", npc_id = "frosthold_guard_captain" },
  stages = {
    {
      id = "find_war_party",
      name = "Locate the War Party",
      description = "Track the orc war party to their camp in the Northern Highlands.",
      objectives = {
        { id = "visit_camp", type = "visit_location", description = "Reach the orc camp", location_id = "northern_highlands" },
      },
    },
    {
      id = "slay_chieftains",
      name = "Slay the Chieftains",
      description = "Defeat the orc chieftains leading the raids.",
      requirements = { "find_war_party" },
      objectives = {
        { id = "kill_orcs", type = "kill_entities", description = "Orc chieftains slain", count = 3, entity_template = "orc" },
      },
    },
    {
      id = "return_guard",
      name = "Report to the Guard Captain",
      description = "Return to Frosthold and report the orc threat is neutralized.",
      requirements = { "slay_chieftains" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to captain", npc_id = "frosthold_guard_captain" },
      },
    },
  },
  rewards = { experience = 150, gold = 40 },
})