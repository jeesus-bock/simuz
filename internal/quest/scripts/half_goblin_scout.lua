-- The Half-Goblin Scout
-- A half-goblin scout needs help locating a hidden kobold warren.

quest.define({
  id = "half_goblin_scout",
  title = "The Half-Goblin Scout",
  type = "side",
  level = 2,
  description = "A half-goblin scout has spotted a hidden kobold warren and needs backup to scout it safely.",
  source = { type = "npc", npc_id = "half_goblin_scout" },
  stages = {
    {
      id = "find_warren",
      name = "Find the Kobold Warren",
      description = "Follow the half-goblin scout to the hidden warren.",
      objectives = {
        { id = "visit_warren", type = "visit_location", description = "Reach the kobold warren", location_id = "kobold_warren" },
      },
    },
    {
      id = "scout_warren",
      name = "Scout the Warren",
      description = "Map the warren interior and report back.",
      requirements = { "find_warren" },
      objectives = {
        { id = "kill_kobolds", type = "kill_entities", description = "Kobolds scouted", count = 2, entity_template = "kobold" },
      },
    },
    {
      id = "report_findings",
      name = "Report Findings",
      description = "Return to the half-goblin scout with your findings.",
      requirements = { "scout_warren" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to scout", npc_id = "half_goblin_scout" },
      },
    },
  },
  rewards = { experience = 70, gold = 15 },
})