-- The Half-Gnoll Pack
-- A half-gnoll outcast seeks to prove her worth to the pack.

quest.define({
  id = "half_gnoll_pack",
  title = "The Half-Gnoll Pack",
  type = "side",
  level = 3,
  description = "A half-gnoll outcast seeks to prove her worth by eliminating a rival gnoll pack.",
  source = { type = "npc", npc_id = "half_gnoll_outcast" },
  stages = {
    {
      id = "find_rival_pack",
      name = "Find the Rival Pack",
      description = "Track the rival gnoll pack through the waste lands.",
      objectives = {
        { id = "visit_waste", type = "visit_location", description = "Reach the waste lands", location_id = "ash_desert" },
      },
    },
    {
      id = "defeat_pack",
      name = "Defeat the Rival Pack",
      description = "Eliminate the rival gnoll pack.",
      requirements = { "find_rival_pack" },
      objectives = {
        { id = "kill_gnolls", type = "kill_entities", description = "Gnolls defeated", count = 4, entity_template = "gnoll" },
      },
    },
    {
      id = "earn_loyalty",
      name = "Earn the Pack's Loyalty",
      description = "Return to the half-gnoll outcast and claim her place.",
      requirements = { "defeat_pack" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to outcast", npc_id = "half_gnoll_outcast" },
      },
    },
  },
  rewards = { experience = 140, gold = 35 },
})