-- The Hobgoblin Siege
-- Hobgoblins are besieging a frosthold outpost.

quest.define({
  id = "hobgoblin_siege",
  title = "The Hobgoblin Siege",
  type = "side",
  level = 4,
  description = "A disciplined hobgoblin war band has besieged the Frosthold outpost. Break the siege and drive them back.",
  source = { type = "npc", npc_id = "frosthold_guard_captain" },
  stages = {
    {
      id = "reach_outpost",
      name = "Reach the Outpost",
      description = "Travel to the besieged Frosthold outpost.",
      objectives = {
        { id = "visit_outpost", type = "visit_location", description = "Reach the outpost", location_id = "frosthold" },
      },
    },
    {
      id = "break_siege",
      name = "Break the Siege",
      description = "Defeat the hobgoblin commanders and scatter their forces.",
      requirements = { "reach_outpost" },
      objectives = {
        { id = "kill_hobgoblins", type = "kill_entities", description = "Hobgoblins defeated", count = 5, entity_template = "hobgoblin" },
      },
    },
    {
      id = "pursue_retreat",
      name = "Pursue the Retreat",
      description = "Track the fleeing hobgoblins to their war camp and destroy it.",
      requirements = { "break_siege" },
      objectives = {
        { id = "visit_war_camp", type = "visit_location", description = "Reach the hobgoblin war camp", location_id = "northern_highlands" },
      },
    },
  },
  rewards = { experience = 250, gold = 60 },
})