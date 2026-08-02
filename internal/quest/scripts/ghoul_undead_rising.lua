-- The Ghoul Undead Rising
-- Ghouls have been rising from the underground graveyard.

quest.define({
  id = "ghoul_undead_rising",
  title = "The Ghoul Undead Rising",
  type = "side",
  level = 4,
  description = "Ghoul undead have been rising from the underground graveyard. Put them to rest before they spread.",
  source = { type = "npc", npc_id = "frosthold_priest" },
  stages = {
    {
      id = "enter_underground",
      name = "Enter the Underground",
      description = "Descend into the underground graveyard.",
      objectives = {
        { id = "visit_underground", type = "visit_location", description = "Reach the underground", location_id = "rat_king_lair" },
      },
    },
    {
      id = "destroy_ghouls",
      name = "Destroy the Ghouls",
      description = "Defeat the ghoul undead.",
      requirements = { "enter_underground" },
      objectives = {
        { id = "kill_ghouls", type = "kill_entities", description = "Ghoul undead destroyed", count = 4, entity_template = "ghoul" },
      },
    },
    {
      id = "purify_graveyard",
      name = "Purify the Graveyard",
      description = "Return to the priest and have the graveyard purified.",
      requirements = { "destroy_ghouls" },
      objectives = {
        { id = "report_priest", type = "talk_to_npc", description = "Report to priest", npc_id = "frosthold_priest" },
      },
    },
  },
  rewards = { experience = 200, gold = 50 },
})