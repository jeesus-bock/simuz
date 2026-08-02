-- The Lich's Tower
-- A lich has been raising undead in the tower ruins.

quest.define({
  id = "lich_tower",
  title = "The Lich's Tower",
  type = "side",
  level = 6,
  description = "A lich has taken up residence in the tower ruins, raising an army of undead. Destroy the lich and end its necromancy.",
  source = { type = "npc", npc_id = "frosthold_priest" },
  stages = {
    {
      id = "find_tower",
      name = "Find the Tower",
      description = "Locate the tower ruins where the lich resides.",
      objectives = {
        { id = "visit_tower", type = "visit_location", description = "Reach the tower", location_id = "rat_king_lair" },
      },
    },
    {
      id = "slay_lich",
      name = "Slay the Lich",
      description = "Defeat the undead lich and end its necromancy.",
      requirements = { "find_tower" },
      objectives = {
        { id = "kill_lich", type = "kill_entities", description = "Lich destroyed", count = 1, entity_template = "lich" },
      },
    },
    {
      id = "purify_tower",
      name = "Purify the Tower",
      description = "Return to the priest and have the tower purified.",
      requirements = { "slay_lich" },
      objectives = {
        { id = "report_priest", type = "talk_to_npc", description = "Report to priest", npc_id = "frosthold_priest" },
      },
    },
  },
  rewards = { experience = 400, gold = 100 },
})