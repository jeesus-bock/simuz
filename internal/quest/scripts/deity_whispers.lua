-- Whispers from Olympus
-- Zeus seeks a mortal champion to investigate disturbances.

quest.define({
  id = "deity_whispers",
  title = "Whispers from Olympus",
  type = "main",
  level = 5,
  description = "Strange dreams plague the townsfolk. Zeus seeks a mortal champion to investigate disturbances in the mortal realm.",
  source = { type = "npc", npc_id = "zeus" },
  stages = {
    {
      id = "visit_temple",
      name = "Seek Guidance",
      description = "Travel to a temple and pray for guidance.",
      objectives = {
        { id = "visit_temple", type = "visit_location", description = "Visit a temple", location_id = "frosthold_temple" },
      },
    },
    {
      id = "investigate",
      name = "Investigate the Anomaly",
      description = "Head to the Ash Desert and investigate the magical anomaly.",
      requirements = { "visit_temple" },
      objectives = {
        { id = "reach_ash", type = "visit_location", description = "Reach the Ash Desert", location_id = "ash_desert" },
      },
    },
  },
  rewards = { experience = 500, gold = 100 },
})

return {
  util.event("quest_accepted", { quest_id = "deity_whispers", source = "zeus" }),
}
