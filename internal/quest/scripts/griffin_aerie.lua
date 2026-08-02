-- The Griffin's Aerie
-- A griffin has been attacking livestock in the highlands.

quest.define({
  id = "griffin_aerie",
  title = "The Griffin's Aerie",
  type = "side",
  level = 4,
  description = "A griffin has been swooping down on livestock in the Northern Highlands. Find its aerie and calm the beast.",
  source = { type = "npc", npc_id = "northern_highlands_rancher" },
  stages = {
    {
      id = "find_aerie",
      name = "Find the Aerie",
      description = "Locate the griffin's aerie in the highlands.",
      objectives = {
        { id = "visit_highlands", type = "visit_location", description = "Reach the highlands", location_id = "northern_highlands" },
      },
    },
    {
      id = "calm_griffin",
      name = "Calm the Griffin",
      description = "Subdue the griffin without killing it.",
      requirements = { "find_aerie" },
      objectives = {
        { id = "tame_griffin", type = "kill_entities", description = "Griffin subdued", count = 1, entity_template = "griffin" },
      },
    },
    {
      id = "return_rancher",
      name = "Return to the Rancher",
      description = "Report back to the rancher that the griffin is contained.",
      requirements = { "calm_griffin" },
      objectives = {
        { id = "report_rancher", type = "talk_to_npc", description = "Report to rancher", npc_id = "northern_highlands_rancher" },
      },
    },
  },
  rewards = { experience = 200, gold = 50 },
})