-- The Beholder's Eye Rays
-- A beholder has been terrorizing the caves with its deadly eye rays.

quest.define({
  id = "beholder_eye_rays",
  title = "The Beholder's Eye Rays",
  type = "side",
  level = 5,
  description = "A beholder with deadly eye rays has taken up residence in the caves. Destroy the aberration before it destroys all who enter.",
  source = { type = "npc", npc_id = "frosthold_priest" },
  stages = {
    {
      id = "find_beholder",
      name = "Find the Beholder",
      description = "Locate the beholder's lair in the caves.",
      objectives = {
        { id = "visit_caves", type = "visit_location", description = "Reach the caves", location_id = "rat_king_lair" },
      },
    },
    {
      id = "destroy_beholder",
      name = "Destroy the Beholder",
      description = "Defeat the beholder and end its eye ray terror.",
      requirements = { "find_beholder" },
      objectives = {
        { id = "kill_beholder", type = "kill_entities", description = "Beholder destroyed", count = 1, entity_template = "beholder" },
      },
    },
    {
      id = "purify_caves",
      name = "Purify the Caves",
      description = "Return to the priest and report the beholder's defeat.",
      requirements = { "destroy_beholder" },
      objectives = {
        { id = "report_priest", type = "talk_to_npc", description = "Report to priest", npc_id = "frosthold_priest" },
      },
    },
  },
  rewards = { experience = 380, gold = 95 },
})