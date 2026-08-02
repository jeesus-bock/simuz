-- The Gnoll Nomad Raid
-- Gnoll nomads have been raiding caravans in the waste lands.

quest.define({
  id = "gnoll_nomad_raid",
  title = "The Gnoll Nomad Raid",
  type = "side",
  level = 3,
  description = "Gnoll nomads have been attacking trade caravans in the waste lands. Hunt them down and end their raids.",
  source = { type = "npc", npc_id = "frosthold_guard_captain" },
  stages = {
    {
      id = "track_nomads",
      name = "Track the Nomads",
      description = "Follow the gnoll trail through the waste lands.",
      objectives = {
        { id = "visit_waste", type = "visit_location", description = "Reach the waste lands", location_id = "ash_desert" },
      },
    },
    {
      id = "hunt_nomads",
      name = "Hunt the Nomads",
      description = "Defeat the gnoll raiding party.",
      requirements = { "track_nomads" },
      objectives = {
        { id = "kill_gnolls", type = "kill_entities", description = "Gnolls defeated", count = 4, entity_template = "gnoll" },
      },
    },
    {
      id = "return_reward",
      name = "Return for Reward",
      description = "Report back to the guard captain with news of your victory.",
      requirements = { "hunt_nomads" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to captain", npc_id = "frosthold_guard_captain" },
      },
    },
  },
  rewards = { experience = 140, gold = 35 },
})