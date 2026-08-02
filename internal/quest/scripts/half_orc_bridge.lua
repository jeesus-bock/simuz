-- The Half-Orc's Bridge
-- A half-orc needs help clearing a bridge of hostile creatures.

quest.define({
  id = "half_orc_bridge",
  title = "The Half-Orc's Bridge",
  type = "side",
  level = 2,
  description = "A half-orc merchant needs the bridge to the Northern Highlands cleared of hostile creatures so she can trade safely.",
  source = { type = "npc", npc_id = "half_orc_merchant" },
  stages = {
    {
      id = "clear_bridge",
      name = "Clear the Bridge",
      description = "Defeat the hostile creatures blocking the bridge.",
      objectives = {
        { id = "kill_hostiles", type = "kill_entities", description = "Hostiles cleared", count = 3, entity_template = "goblin" },
      },
    },
    {
      id = "escort_merchant",
      name = "Escort the Merchant",
      description = "Accompany the half-orc merchant to Frosthold.",
      requirements = { "clear_bridge" },
      objectives = {
        { id = "visit_frosthold", type = "visit_location", description = "Reach Frosthold", location_id = "frosthold" },
      },
    },
  },
  rewards = { experience = 60, gold = 20 },
})