-- The Ancient Tree
-- Elven elders ask for help protecting a sacred tree.

quest.define({
  id = "elf_ancient_tree",
  title = "The Ancient Tree",
  type = "side",
  level = 4,
  description = "The elven elders of the Crystal Forest seek protection for their ancient tree from encroaching dark forces.",
  source = { type = "npc", npc_id = "elf_elder" },
  stages = {
    {
      id = "reach_tree",
      name = "Reach the Ancient Tree",
      description = "Travel to the heart of the Crystal Forest and find the ancient tree.",
      objectives = {
        { id = "visit_tree", type = "visit_location", description = "Find the ancient tree", location_id = "crystal_forest" },
      },
    },
    {
      id = "defend_tree",
      name = "Defend the Tree",
      description = "Protect the ancient tree from dark creatures.",
      requirements = { "reach_tree" },
      objectives = {
        { id = "kill_dark", type = "kill_entities", description = "Dark creatures slain", count = 5, entity_template = "wraith" },
      },
    },
    {
      id = "bless_tree",
      name = "Bless the Tree",
      description = "Return to the elven elder and report your success.",
      requirements = { "defend_tree" },
      objectives = {
        { id = "report_elder", type = "talk_to_npc", description = "Report to elder", npc_id = "elf_elder" },
      },
    },
  },
  rewards = { experience = 200, gold = 50 },
})