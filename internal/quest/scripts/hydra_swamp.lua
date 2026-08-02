-- The Hydra of the Swamp
-- A hydra has taken up residence in the swamp near the Crystal Forest.

quest.define({
  id = "hydra_swamp",
  title = "The Hydra of the Swamp",
  type = "side",
  level = 5,
  description = "A multi-headed hydra has claimed the swamp near the Crystal Forest. Slay the beast before it spreads terror.",
  source = { type = "npc", npc_id = "fairy_sparkle" },
  stages = {
    {
      id = "find_hydra",
      name = "Find the Hydra",
      description = "Locate the hydra's lair in the swamp.",
      objectives = {
        { id = "visit_swamp", type = "visit_location", description = "Reach the swamp", location_id = "crystal_forest" },
      },
    },
    {
      id = "slay_hydra",
      name = "Slay the Hydra",
      description = "Defeat the hydra. Beware — cut one head and two more may grow.",
      requirements = { "find_hydra" },
      objectives = {
        { id = "kill_hydra", type = "kill_entities", description = "Hydra slain", count = 1, entity_template = "hydra" },
      },
    },
    {
      id = "claim_reward",
      name = "Claim the Reward",
      description = "Return to the fairy and report the hydra's demise.",
      requirements = { "slay_hydra" },
      objectives = {
        { id = "report_fairy", type = "talk_to_npc", description = "Report to fairy", npc_id = "fairy_sparkle" },
      },
    },
  },
  rewards = { experience = 350, gold = 90 },
})