-- The Rat Problem
-- Greta asks the player to clear rats from the inn cellar.

quest.define({
  id = "rat_problem",
  title = "The Rat Problem",
  type = "side",
  level = 1,
  description = "The innkeeper Greta in Frosthold has reported a rat infestation in the cellar. Clear out the vermin.",
  source = { type = "npc", npc_id = "frosthold_greta" },
  stages = {
    {
      id = "clear_rats",
      name = "Clear the Cellar",
      description = "Kill the rats in the Sleeping Dragon's cellar.",
      objectives = {
        { id = "kill_rats", type = "kill_entities", description = "Rats slain", count = 5, entity_template = "rat" },
      },
    },
    {
      id = "report",
      name = "Report to Greta",
      description = "Return to Greta and claim your reward.",
      requirements = { "clear_rats" },
      objectives = {
        { id = "talk_greta", type = "talk_to_npc", description = "Speak with Greta", npc_id = "frosthold_greta" },
      },
    },
  },
  rewards = { experience = 50, gold = 10 },
})
