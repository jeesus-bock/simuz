-- Lost Heirlooms
-- Father Luan's ceremonial items were taken into the Rat King's lair.

quest.define({
  id = "lost_heirlooms",
  title = "Lost Heirlooms",
  type = "side",
  level = 2,
  description = "Father Luan at the Temple of Light has lost several ceremonial heirlooms. They may have been taken to the Rat King's lair.",
  source = { type = "npc", npc_id = "frosthold_priest" },
  stages = {
    {
      id = "search_dungeon",
      name = "Search the Rat King's Lair",
      description = "Find the heirlooms in the Rat King's lair.",
      objectives = {
        { id = "find_heirlooms", type = "collect_items", description = "Heirlooms found", count = 3, location_id = "rat_king_lair" },
      },
    },
    {
      id = "return_heirlooms",
      name = "Return to Father Luan",
      description = "Bring the heirlooms back to the temple.",
      requirements = { "search_dungeon" },
      objectives = {
        { id = "return_heirlooms", type = "deliver_item", description = "Return heirlooms", npc_id = "frosthold_priest" },
      },
    },
  },
  rewards = { experience = 100, gold = 25 },
})
