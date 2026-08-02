-- The Half-Kobold Tunnel
-- A half-kobold needs help securing a tunnel network.

quest.define({
  id = "half_kobold_tunnel",
  title = "The Half-Kobold Tunnel",
  type = "side",
  level = 2,
  description = "A half-kobold tunnel scout needs the tunnel network cleared of dangerous creatures.",
  source = { type = "npc", npc_id = "half_kobold_scout" },
  stages = {
    {
      id = "enter_tunnels",
      name = "Enter the Tunnels",
      description = "Descend into the kobold tunnel network.",
      objectives = {
        { id = "visit_tunnels", type = "visit_location", description = "Reach the tunnels", location_id = "kobold_warren" },
      },
    },
    {
      id = "clear_tunnels",
      name = "Clear the Tunnels",
      description = "Defeat the creatures infesting the tunnels.",
      requirements = { "enter_tunnels" },
      objectives = {
        { id = "kill_creatures", type = "kill_entities", description = "Creatures cleared", count = 4, entity_template = "rat" },
      },
    },
    {
      id = "secure_network",
      name = "Secure the Network",
      description = "Return to the half-kobold scout and confirm the tunnels are safe.",
      requirements = { "clear_tunnels" },
      objectives = {
        { id = "report", type = "talk_to_npc", description = "Report to scout", npc_id = "half_kobold_scout" },
      },
    },
  },
  rewards = { experience = 70, gold = 15 },
})