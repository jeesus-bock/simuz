-- The Taken Courier
-- A Frosthold courier vanished on the road and was last seen near a kobold warren.

quest.define({
  id = "taken_courier",
  title = "The Taken Courier",
  type = "side",
  level = 3,
  description = "Captain Halvar asks you to track a missing courier who was ambushed near the Crystal Forest. Follow the trail into the kobold warrens, clear out the captors, and report back.",
  source = { type = "npc", npc_id = "frosthold_guard_captain" },
  stages = {
    {
      id = "find_warren",
      name = "Find the Warren",
      description = "Follow the trail into the kobold tunnels.",
      objectives = {
        { id = "visit_warren", type = "visit_location", description = "Reach the kobold warren", location_id = "kobold_warren" },
      },
    },
    {
      id = "clear_captors",
      name = "Drive Off the Captors",
      description = "Defeat the kobolds holding the courier.",
      requirements = { "find_warren" },
      objectives = {
        { id = "kill_kobolds", type = "kill_entities", description = "Kobolds defeated", count = 3, entity_template = "kobold" },
      },
    },
    {
      id = "report_back",
      name = "Report to Captain Halvar",
      description = "Return to Frosthold and tell the captain the road is clear.",
      requirements = { "clear_captors" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to the captain", npc_id = "frosthold_guard_captain" },
      },
    },
  },
  rewards = { experience = 120, gold = 25 },
})
