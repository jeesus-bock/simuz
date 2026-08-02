-- The Zombie Plague
-- A zombie plague has broken out near the Frosthold cemetery.

quest.define({
  id = "zombie_plague",
  title = "The Zombie Plague",
  type = "side",
  level = 3,
  description = "Zombies have risen from the Frosthold cemetery, spreading a plague. Contain the outbreak.",
  source = { type = "npc", npc_id = "frosthold_priest" },
  stages = {
    {
      id = "contain_outbreak",
      name = "Contain the Outbreak",
      description = "Reach the cemetery and stop the zombie plague.",
      objectives = {
        { id = "visit_cemetery", type = "visit_location", description = "Reach the cemetery", location_id = "frosthold" },
      },
    },
    {
      id = "slay_zombies",
      name = "Slay the Zombies",
      description = "Defeat the undead horde.",
      requirements = { "contain_outbreak" },
      objectives = {
        { id = "kill_zombies", type = "kill_entities", description = "Zombies slain", count = 5, entity_template = "zombie" },
      },
    },
    {
      id = "purify_ground",
      name = "Purify the Ground",
      description = "Return to the priest and help purify the cursed ground.",
      requirements = { "slay_zombies" },
      objectives = {
        { id = "report_priest", type = "talk_to_npc", description = "Report to priest", npc_id = "frosthold_priest" },
      },
    },
  },
  rewards = { experience = 150, gold = 35 },
})