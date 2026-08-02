-- The Cockatrice Nest
-- Cockatrices have been nesting near the Golden Gate farms.

quest.define({
  id = "cockatrice_nest",
  title = "The Cockatrice Nest",
  type = "side",
  level = 3,
  description = "Cockatrices have been nesting near the Golden Gate farms, turning livestock to stone. Destroy their nest.",
  source = { type = "npc", npc_id = "golden_gate_farmer" },
  stages = {
    {
      id = "find_nest",
      name = "Find the Nest",
      description = "Locate the cockatrice nest in the Golden Plains.",
      objectives = {
        { id = "visit_plains", type = "visit_location", description = "Reach the Golden Plains", location_id = "hag_cottage" },
      },
    },
    {
      id = "destroy_nest",
      name = "Destroy the Nest",
      description = "Slay the cockatrices and destroy their eggs.",
      requirements = { "find_nest" },
      objectives = {
        { id = "kill_cockatrices", type = "kill_entities", description = "Cockatrices slain", count = 3, entity_template = "cockatrice" },
      },
    },
    {
      id = "protect_farms",
      name = "Protect the Farms",
      description = "Return to the farmer and assure them the threat is over.",
      requirements = { "destroy_nest" },
      objectives = {
        { id = "report_farmer", type = "talk_to_npc", description = "Report to farmer", npc_id = "golden_gate_farmer" },
      },
    },
  },
  rewards = { experience = 130, gold = 30 },
})