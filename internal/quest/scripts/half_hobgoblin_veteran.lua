-- The Half-Hobgoblin Veteran
-- A half-hobgoblin veteran seeks to redeem her clan's honor.

quest.define({
  id = "half_hobgoblin_veteran",
  title = "The Half-Hobgoblin Veteran",
  type = "side",
  level = 4,
  description = "A half-hobgoblin veteran of the Frosthold guard seeks to redeem her clan's honor by eliminating a hobgoblin war band.",
  source = { type = "npc", npc_id = "half_hobgoblin_veteran" },
  stages = {
    {
      id = "find_war_band",
      name = "Find the War Band",
      description = "Track the hobgoblin war band to their camp.",
      objectives = {
        { id = "visit_camp", type = "visit_location", description = "Reach the hobgoblin camp", location_id = "northern_highlands" },
      },
    },
    {
      id = "defeat_war_band",
      name = "Defeat the War Band",
      description = "Defeat the hobgoblin war band and prove her valor.",
      requirements = { "find_war_band" },
      objectives = {
        { id = "kill_hobgoblins", type = "kill_entities", description = "Hobgoblins defeated", count = 4, entity_template = "hobgoblin" },
      },
    },
    {
      id = "earn_honor",
      name = "Earn Honor",
      description = "Return to the half-hobgoblin veteran and claim her honor.",
      requirements = { "defeat_war_band" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to veteran", npc_id = "half_hobgoblin_veteran" },
      },
    },
  },
  rewards = { experience = 200, gold = 50 },
})