-- The Bard's Ballad
-- Help Lira find inspiration for a new song.

quest.define({
  id = "bard_ballad",
  title = "The Bard's Ballad",
  type = "side",
  level = 1,
  description = "A traveling bard seeks inspiration. Help them find a worthy story to sing about.",
  source = { type = "npc", npc_id = "bard_lira" },
  stages = {
    {
      id = "find_story",
      name = "Find a Story",
      description = "Travel to different towns and gather tales.",
      objectives = {
        { id = "visit_towns", type = "visit_location", description = "Visit Stillwater", location_id = "stillwater" },
      },
    },
    {
      id = "return_bard",
      name = "Return to the Bard",
      description = "Share your tales with the bard.",
      requirements = { "find_story" },
      objectives = {
        { id = "talk_bard", type = "talk_to_npc", description = "Speak with bard", npc_id = "bard_lira" },
      },
    },
  },
  rewards = { experience = 25, gold = 5 },
})
