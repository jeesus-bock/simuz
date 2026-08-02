-- The Wraith in the Shadow
-- A wraith haunts the shadow lands near the Crystal Forest.

quest.define({
  id = "wraith_shadow",
  title = "The Wraith in the Shadow",
  type = "side",
  level = 4,
  description = "A wraith has been haunting the shadow lands, draining the life from travelers. Put the spirit to rest.",
  source = { type = "npc", npc_id = "frosthold_guard_captain" },
  stages = {
    {
      id = "enter_shadow",
      name = "Enter the Shadow Lands",
      description = "Brave the shadow lands and find the wraith.",
      objectives = {
        { id = "visit_shadow", type = "visit_location", description = "Reach the shadow lands", location_id = "crystal_forest" },
      },
    },
    {
      id = "destroy_wraith",
      name = "Destroy the Wraith",
      description = "Defeat the incorporeal wraith.",
      requirements = { "enter_shadow" },
      objectives = {
        { id = "kill_wraith", type = "kill_entities", description = "Wraith destroyed", count = 1, entity_template = "wraith" },
      },
    },
    {
      id = "restore_light",
      name = "Restore Light to the Forest",
      description = "Return to the guard captain and report the wraith's destruction.",
      requirements = { "destroy_wraith" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to captain", npc_id = "frosthold_guard_captain" },
      },
    },
  },
  rewards = { experience = 260, gold = 65 },
})