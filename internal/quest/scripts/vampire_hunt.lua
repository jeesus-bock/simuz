-- The Vampire's Grasp
-- Seek Ravenmoor Manor and end Count Valerius.

quest.define({
  id = "vampire_hunt",
  title = "The Vampire's Grasp",
  type = "side",
  level = 4,
  description = "Whispers speak of a vampire lord in the northern lands. Seek out Ravenmoor Manor and end the undead threat.",
  source = { type = "npc", npc_id = "frosthold_priest" },
  stages = {
    {
      id = "find_manor",
      name = "Locate Ravenmoor Manor",
      description = "Find the vampire's lair in the Northern Highlands.",
      objectives = {
        { id = "visit_manor", type = "visit_location", description = "Find the manor", location_id = "ravenmoor_manor" },
      },
    },
    {
      id = "slay_vampire",
      name = "Slay the Vampire",
      description = "Defeat Count Valerius in his coffin chamber.",
      requirements = { "find_manor" },
      objectives = {
        { id = "kill_vampire", type = "kill_entities", description = "Vampire slain", count = 1, entity_template = "vampire" },
      },
    },
    {
      id = "bless_manor",
      name = "Cleanse the Manor",
      description = "Return to the priest and report the vampire's demise.",
      requirements = { "slay_vampire" },
      objectives = {
        { id = "report_priest", type = "talk_to_npc", description = "Report to priest", npc_id = "frosthold_priest" },
      },
    },
  },
  rewards = { experience = 300, gold = 75 },
})

return {
  util.event("quest_accepted", { quest_id = "vampire_hunt", source = "frosthold_priest" }),
}
