-- The Ogre's Gluttony
-- An ogre has been devouring livestock near Golden Gate.

quest.define({
  id = "ogre_gluttony",
  title = "The Ogre's Gluttony",
  type = "side",
  level = 5,
  description = "A ravenous ogre has been destroying farms and devouring livestock near Golden Gate. Put an end to its appetite.",
  source = { type = "npc", npc_id = "golden_gate_farmer" },
  stages = {
    {
      id = "find_ogre",
      name = "Find the Ogre",
      description = "Track the ogre to its lair in the Golden Plains.",
      objectives = {
        { id = "visit_lair", type = "visit_location", description = "Find the ogre's lair", location_id = "hag_cottage" },
      },
    },
    {
      id = "slay_ogre",
      name = "Slay the Ogre",
      description = "Defeat the ogre and end its reign of terror.",
      requirements = { "find_ogre" },
      objectives = {
        { id = "kill_ogre", type = "kill_entities", description = "Ogre slain", count = 1, entity_template = "ogre" },
      },
    },
    {
      id = "salvage_farms",
      name = "Salvage the Farms",
      description = "Return to the farmers and help them rebuild.",
      requirements = { "slay_ogre" },
      objectives = {
        { id = "report_farmer", type = "talk_to_npc", description = "Report to farmer", npc_id = "golden_gate_farmer" },
      },
    },
  },
  rewards = { experience = 300, gold = 80 },
})