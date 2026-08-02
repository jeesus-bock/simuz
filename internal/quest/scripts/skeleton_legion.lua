-- The Skeleton Legion
-- Skeletons have risen from the graveyard ruins.

quest.define({
  id = "skeleton_legion",
  title = "The Skeleton Legion",
  type = "side",
  level = 3,
  description = "Skeletons have risen from the graveyard ruins and threaten the nearby settlements. Put them to rest.",
  source = { type = "npc", npc_id = "frosthold_priest" },
  stages = {
    {
      id = "find_graveyard",
      name = "Find the Graveyard",
      description = "Locate the graveyard ruins where the skeletons have risen.",
      objectives = {
        { id = "visit_graveyard", type = "visit_location", description = "Reach the graveyard", location_id = "rat_king_lair" },
      },
    },
    {
      id = "destroy_skeletons",
      name = "Destroy the Skeletons",
      description = "Defeat the skeleton legion.",
      requirements = { "find_graveyard" },
      objectives = {
        { id = "kill_skeletons", type = "kill_entities", description = "Skeletons destroyed", count = 6, entity_template = "skeleton" },
      },
    },
    {
      id = "bless_graveyard",
      name = "Bless the Graveyard",
      description = "Return to the priest and have the graveyard blessed.",
      requirements = { "destroy_skeletons" },
      objectives = {
        { id = "report_priest", type = "talk_to_npc", description = "Report to priest", npc_id = "frosthold_priest" },
      },
    },
  },
  rewards = { experience = 160, gold = 40 },
})