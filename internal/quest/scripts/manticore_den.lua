-- The Manticore's Den
-- A manticore has been stalking the forest roads.

quest.define({
  id = "manticore_den",
  title = "The Manticore's Den",
  type = "side",
  level = 4,
  description = "A manticore with venomous spikes has been preying on travelers in the forest. Find and slay the beast.",
  source = { type = "npc", npc_id = "frosthold_guard_captain" },
  stages = {
    {
      id = "track_manticore",
      name = "Track the Manticore",
      description = "Follow the manticore's trail through the forest.",
      objectives = {
        { id = "visit_forest", type = "visit_location", description = "Enter the forest", location_id = "crystal_forest" },
      },
    },
    {
      id = "slay_manticore",
      name = "Slay the Manticore",
      description = "Defeat the venomous manticore.",
      requirements = { "track_manticore" },
      objectives = {
        { id = "kill_manticore", type = "kill_entities", description = "Manticore slain", count = 1, entity_template = "manticore" },
      },
    },
    {
      id = "claim_reward",
      name = "Claim the Reward",
      description = "Return to the guard captain with proof of the manticore's demise.",
      requirements = { "slay_manticore" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to captain", npc_id = "frosthold_guard_captain" },
      },
    },
  },
  rewards = { experience = 250, gold = 60 },
})