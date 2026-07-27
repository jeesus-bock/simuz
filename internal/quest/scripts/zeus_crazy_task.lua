-- Zeus's Divine Challenge
-- Climb, slay a guardian, and return as a hero.

quest.define({
  id = "zeus_crazy_task",
  title = "Zeus's Divine Challenge",
  type = "main",
  level = 1,
  description = "Zeus has chosen you for a divine challenge! Prove your worth by completing an impossible task.",
  source = { type = "npc", npc_id = "zeus" },
  stages = {
    {
      id = "visit_mountain",
      name = "Ascend the Mountain",
      description = "Climb to the highest peak in the Northern Highlands.",
      objectives = {
        { id = "climb", type = "visit_location", description = "Reach the mountain peak", location_id = "northern_highlands" },
      },
    },
    {
      id = "defeat_beast",
      name = "Slay the Guardian",
      description = "Defeat the beast that guards the mountain.",
      requirements = { "visit_mountain" },
      objectives = {
        { id = "kill_beast", type = "kill_entities", description = "Guardian slain", count = 1, entity_template = "bear" },
      },
    },
    {
      id = "return_hero",
      name = "Return as a Hero",
      description = "Return to civilization and tell of your exploits.",
      requirements = { "defeat_beast" },
      objectives = {
        { id = "visit_town", type = "visit_location", description = "Return to town", location_id = "frosthold" },
      },
    },
  },
  rewards = { experience = 150, gold = 30 },
})
