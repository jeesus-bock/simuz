-- The Basilisk's Gaze
-- A basilisk has been petrifying travelers in the Ash Desert.

quest.define({
  id = "basilisk_petrification",
  title = "The Basilisk's Gaze",
  type = "side",
  level = 4,
  description = "A basilisk lurks in the Ash Desert, turning travelers to stone with its deadly gaze. End its reign.",
  source = { type = "npc", npc_id = "ash_desert_scout" },
  stages = {
    {
      id = "enter_desert",
      name = "Enter the Ash Desert",
      description = "Travel to the Ash Desert and locate the basilisk's den.",
      objectives = {
        { id = "visit_desert", type = "visit_location", description = "Reach the Ash Desert", location_id = "ash_desert" },
      },
    },
    {
      id = "slay_basilisk",
      name = "Slay the Basilisk",
      description = "Defeat the basilisk and end its petrifying gaze.",
      requirements = { "enter_desert" },
      objectives = {
        { id = "kill_basilisk", type = "kill_entities", description = "Basilisk slain", count = 1, entity_template = "basilisk" },
      },
    },
    {
      id = "rescue_travelers",
      name = "Rescue the Petrified",
      description = "Return to the desert scout and report your victory.",
      requirements = { "slay_basilisk" },
      objectives = {
        { id = "report_scout", type = "talk_to_npc", description = "Report to scout", npc_id = "ash_desert_scout" },
      },
    },
  },
  rewards = { experience = 280, gold = 70 },
})