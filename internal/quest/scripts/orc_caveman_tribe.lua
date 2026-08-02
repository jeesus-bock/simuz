-- The Orc Caveman Tribe
-- An orc caveman tribe needs help defending their territory.

quest.define({
  id = "orc_caveman_tribe",
  title = "The Orc Caveman Tribe",
  type = "side",
  level = 3,
  description = "An orc caveman tribe is under attack by rival species. Help them defend their cave.",
  source = { type = "npc", npc_id = "orc_caveman_chief" },
  stages = {
    {
      id = "reach_cave",
      name = "Reach the Cave",
      description = "Travel to the orc caveman tribe's lair.",
      objectives = {
        { id = "visit_cave", type = "visit_location", description = "Reach the cave", location_id = "northern_highlands" },
      },
    },
    {
      id = "defend_cave",
      name = "Defend the Cave",
      description = "Help the orc cavemen fight off the invaders.",
      requirements = { "reach_cave" },
      objectives = {
        { id = "kill_invaders", type = "kill_entities", description = "Invaders defeated", count = 5, entity_template = "goblin" },
      },
    },
    {
      id = "celebrate_victory",
      name = "Celebrate Victory",
      description = "Return to the orc chieftain and celebrate the victory.",
      requirements = { "defend_cave" },
      objectives = {
        { id = "report_chief", type = "talk_to_npc", description = "Report to chieftain", npc_id = "orc_caveman_chief" },
      },
    },
  },
  rewards = { experience = 130, gold = 30 },
})