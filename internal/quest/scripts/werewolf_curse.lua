-- The Werewolf Curse
-- A werewolf has been terrorizing the village of Stillwater.

quest.define({
  id = "werewolf_curse",
  title = "The Werewolf Curse",
  type = "side",
  level = 4,
  description = "A werewolf prowls the nights around Stillwater, attacking villagers. Find and end the curse.",
  source = { type = "npc", npc_id = "stillwater_merchant" },
  stages = {
    {
      id = "investigate_attacks",
      name = "Investigate the Attacks",
      description = "Gather information about the werewolf's whereabouts.",
      objectives = {
        { id = "visit_stillwater", type = "visit_location", description = "Visit Stillwater", location_id = "stillwater" },
      },
    },
    {
      id = "hunt_werewolf",
      name = "Hunt the Werewolf",
      description = "Track and defeat the werewolf in the forest.",
      requirements = { "investigate_attacks" },
      objectives = {
        { id = "kill_werewolf", type = "kill_entities", description = "Werewolf slain", count = 1, entity_template = "werewolf" },
      },
    },
    {
      id = "protect_village",
      name = "Protect the Village",
      description = "Return to Stillwater and assure the villagers the threat is over.",
      requirements = { "hunt_werewolf" },
      objectives = {
        { id = "report_merchant", type = "talk_to_npc", description = "Report to merchant", npc_id = "stillwater_merchant" },
      },
    },
  },
  rewards = { experience = 220, gold = 55 },
})