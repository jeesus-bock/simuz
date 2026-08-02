-- The Mind Flayer's Lair
-- A mind flayer has been psionically controlling creatures in the underdark.

quest.define({
  id = "mind_flayer_underdark",
  title = "The Mind Flayer's Lair",
  type = "side",
  level = 5,
  description = "A mind flayer has been psionically controlling creatures in the underdark lairs. Destroy the aberration.",
  source = { type = "npc", npc_id = "frosthold_guard_captain" },
  stages = {
    {
      id = "find_underdark",
      name = "Find the Underdark Lair",
      description = "Locate the mind flayer's lair in the underdark.",
      objectives = {
        { id = "visit_underdark", type = "visit_location", description = "Reach the underdark lair", location_id = "rat_king_lair" },
      },
    },
    {
      id = "destroy_flayer",
      name = "Destroy the Mind Flayer",
      description = "Defeat the mind flayer and free its thralls.",
      requirements = { "find_underdark" },
      objectives = {
        { id = "kill_flayer", type = "kill_entities", description = "Mind flayer destroyed", count = 1, entity_template = "mind_flayer" },
      },
    },
    {
      id = "free_thralls",
      name = "Free the Thralls",
      description = "Return to the guard captain and report the mind flayer's defeat.",
      requirements = { "destroy_flayer" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to captain", npc_id = "frosthold_guard_captain" },
      },
    },
  },
  rewards = { experience = 350, gold = 85 },
})