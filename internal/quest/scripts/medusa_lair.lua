-- The Medusa's Lair
-- A medusa has been turning travelers to stone in the caves.

quest.define({
  id = "medusa_lair",
  title = "The Medusa's Lair",
  type = "side",
  level = 5,
  description = "A medusa has taken up residence in the caves, petrifying anyone who enters. End its reign of stone.",
  source = { type = "npc", npc_id = "frosthold_priest" },
  stages = {
    {
      id = "enter_caves",
      name = "Enter the Caves",
      description = "Brave the caves and locate the medusa's lair.",
      objectives = {
        { id = "visit_caves", type = "visit_location", description = "Reach the caves", location_id = "rat_king_lair" },
      },
    },
    {
      id = "slay_medusa",
      name = "Slay the Medusa",
      description = "Defeat the medusa and end its petrifying gaze.",
      requirements = { "enter_caves" },
      objectives = {
        { id = "kill_medusa", type = "kill_entities", description = "Medusa slain", count = 1, entity_template = "medusa" },
      },
    },
    {
      id = "restore_petrified",
      name = "Restore the Petrified",
      description = "Return to the priest and report the medusa's defeat.",
      requirements = { "slay_medusa" },
      objectives = {
        { id = "report_priest", type = "talk_to_npc", description = "Report to priest", npc_id = "frosthold_priest" },
      },
    },
  },
  rewards = { experience = 300, gold = 75 },
})