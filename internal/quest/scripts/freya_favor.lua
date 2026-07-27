-- Freya's Favor
-- Gather crystal blooms as tribute for Freya.

quest.define({
  id = "freya_favor",
  title = "Freya's Favor",
  type = "side",
  level = 3,
  description = "The goddess Freya requests a tribute of rare flowers from the Crystal Forest.",
  source = { type = "npc", npc_id = "freya" },
  stages = {
    {
      id = "gather_flowers",
      name = "Gather Crystal Blooms",
      description = "Collect rare crystal blooms from the Crystal Forest.",
      objectives = {
        { id = "gather", type = "collect_items", description = "Crystal blooms collected", count = 5, location_id = "crystal_forest" },
      },
    },
    {
      id = "offer",
      name = "Present the Tribute",
      description = "Offer the crystal blooms at a shrine.",
      requirements = { "gather_flowers" },
      objectives = {
        { id = "offer", type = "deliver_item", description = "Present tribute", npc_id = "freya" },
      },
    },
  },
  rewards = { experience = 200, gold = 50 },
})
