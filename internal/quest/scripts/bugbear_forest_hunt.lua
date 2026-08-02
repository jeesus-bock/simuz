-- The Bugbear Forest Hunt
-- Bugbears have been stalking travelers through the forest.

quest.define({
  id = "bugbear_forest_hunt",
  title = "The Bugbear Forest Hunt",
  type = "side",
  level = 3,
  description = "Bugbears have been ambushing travelers in the forest. Hunt them down and end their reign of terror.",
  source = { type = "npc", npc_id = "frosthold_guard_captain" },
  stages = {
    {
      id = "track_bugbears",
      name = "Track the Bugbears",
      description = "Follow the bugbear trail deep into the forest.",
      objectives = {
        { id = "visit_forest", type = "visit_location", description = "Enter the forest", location_id = "crystal_forest" },
      },
    },
    {
      id = "hunt_bugbears",
      name = "Hunt the Bugbears",
      description = "Kill the bugbear hunters stalking the forest.",
      requirements = { "track_bugbears" },
      objectives = {
        { id = "kill_bugbears", type = "kill_entities", description = "Bugbears slain", count = 3, entity_template = "bugbear" },
      },
    },
    {
      id = "return_proof",
      name = "Return with Proof",
      description = "Bring back bugbear trophies to the guard captain.",
      requirements = { "hunt_bugbears" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to captain", npc_id = "frosthold_guard_captain" },
      },
    },
  },
  rewards = { experience = 120, gold = 35 },
})