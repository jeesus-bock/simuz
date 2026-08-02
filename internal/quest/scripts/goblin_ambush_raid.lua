-- The Goblin Ambush Raid
-- Goblin raiders have set up an ambush on the road to Stillwater.

quest.define({
  id = "goblin_ambush_raid",
  title = "The Goblin Ambush Raid",
  type = "side",
  level = 2,
  description = "Goblin raiders have ambushed travelers on the road to Stillwater. Clear them out and secure the route.",
  source = { type = "npc", npc_id = "stillwater_merchant" },
  stages = {
    {
      id = "find_ambush",
      name = "Find the Ambush Site",
      description = "Locate the goblin ambush point on the road.",
      objectives = {
        { id = "visit_road", type = "visit_location", description = "Reach the ambush site", location_id = "crystal_forest" },
      },
    },
    {
      id = "eliminate_raiders",
      name = "Eliminate the Raiders",
      description = "Defeat the goblin raiders.",
      requirements = { "find_ambush" },
      objectives = {
        { id = "kill_goblins", type = "kill_entities", description = "Goblins defeated", count = 4, entity_template = "goblin" },
      },
    },
    {
      id = "secure_route",
      name = "Secure the Route",
      description = "Return to the merchant and confirm the road is safe.",
      requirements = { "eliminate_raiders" },
      objectives = {
        { id = "report_merchant", type = "talk_to_npc", description = "Report to merchant", npc_id = "stillwater_merchant" },
      },
    },
  },
  rewards = { experience = 80, gold = 20 },
})