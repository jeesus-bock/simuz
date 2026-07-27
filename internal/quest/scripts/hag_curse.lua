-- The Hag's Curse
-- Break Mirelda's curse on Golden Gate farms.

quest.define({
  id = "hag_curse",
  title = "The Hag's Curse",
  type = "side",
  level = 3,
  description = "Farmers near Golden Gate report a hag cursing their crops. Find her cottage and break the curse.",
  source = { type = "npc", npc_id = "golden_gate_farmer" },
  stages = {
    {
      id = "find_cottage",
      name = "Locate the Hag's Cottage",
      description = "Find the hag's dwelling in the Golden Plains.",
      objectives = {
        { id = "visit_cottage", type = "visit_location", description = "Find the cottage", location_id = "hag_cottage" },
      },
    },
    {
      id = "defeat_hag",
      name = "Defeat the Hag",
      description = "Defeat Mirelda and break her curse.",
      requirements = { "find_cottage" },
      objectives = {
        { id = "kill_hag", type = "kill_entities", description = "Hag defeated", count = 1, entity_template = "hag" },
      },
    },
    {
      id = "return_farmer",
      name = "Return to the Farmer",
      description = "Tell the farmer the curse is broken.",
      requirements = { "defeat_hag" },
      objectives = {
        { id = "report_farmer", type = "talk_to_npc", description = "Report to farmer", npc_id = "golden_gate_farmer" },
      },
    },
  },
  rewards = { experience = 125, gold = 35 },
})
