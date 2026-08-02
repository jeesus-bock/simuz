-- The Cyclops of the Mountain
-- A cyclops has been blocking the mountain pass.

quest.define({
  id = "cyclops_mountain",
  title = "The Cyclops of the Mountain",
  type = "side",
  level = 5,
  description = "A solitary cyclops has taken over the mountain pass, blocking all travel. Slay the beast to reopen the route.",
  source = { type = "npc", npc_id = "frosthold_guard_captain" },
  stages = {
    {
      id = "climb_mountain",
      name = "Climb the Mountain",
      description = "Ascend the mountain to reach the cyclops's lair.",
      objectives = {
        { id = "visit_peak", type = "visit_location", description = "Reach the mountain peak", location_id = "northern_highlands" },
      },
    },
    {
      id = "slay_cyclops",
      name = "Slay the Cyclops",
      description = "Defeat the solitary cyclops.",
      requirements = { "climb_mountain" },
      objectives = {
        { id = "kill_cyclops", type = "kill_entities", description = "Cyclops slain", count = 1, entity_template = "cyclops" },
      },
    },
    {
      id = "reopen_pass",
      name = "Reopen the Pass",
      description = "Return to Frosthold and confirm the mountain pass is safe.",
      requirements = { "slay_cyclops" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to captain", npc_id = "frosthold_guard_captain" },
      },
    },
  },
  rewards = { experience = 320, gold = 80 },
})