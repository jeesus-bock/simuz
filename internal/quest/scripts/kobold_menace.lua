-- The Kobold Menace
-- Guard captain wants kobold raiders cleared from Crystal Forest.

quest.define({
  id = "kobold_menace",
  title = "The Kobold Menace",
  type = "side",
  level = 2,
  description = "Kobolds have been raiding caravans near the Crystal Forest. The guard captain wants them dealt with.",
  source = { type = "npc", npc_id = "frosthold_guard_captain" },
  stages = {
    {
      id = "investigate_raids",
      name = "Investigate the Raids",
      description = "Travel to the Crystal Forest and find evidence of kobold activity.",
      objectives = {
        { id = "find_kobolds", type = "visit_location", description = "Reach Crystal Forest", location_id = "crystal_forest" },
      },
    },
    {
      id = "eliminate_kobolds",
      name = "Eliminate the Threat",
      description = "Defeat the kobold raiding party.",
      requirements = { "investigate_raids" },
      objectives = {
        { id = "kill_kobolds", type = "kill_entities", description = "Kobolds defeated", count = 3, entity_template = "kobold" },
      },
    },
    {
      id = "report_victory",
      name = "Report Success",
      description = "Return to the guard captain with news of your victory.",
      requirements = { "eliminate_kobolds" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to captain", npc_id = "frosthold_guard_captain" },
      },
    },
  },
  rewards = { experience = 75, gold = 15 },
})
