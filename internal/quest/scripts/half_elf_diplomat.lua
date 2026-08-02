-- The Half-Elf Diplomat
-- A half-elf diplomat needs rare herbs for a peace negotiation.

quest.define({
  id = "half_elf_diplomat",
  title = "The Half-Elf Diplomat",
  type = "side",
  level = 3,
  description = "A half-elf diplomat seeks rare crystal blooms from the Crystal Forest to use as a peace offering.",
  source = { type = "npc", npc_id = "half_elf_diplomat" },
  stages = {
    {
      id = "gather_blooms",
      name = "Gather Crystal Blooms",
      description = "Collect rare crystal blooms from the Crystal Forest.",
      objectives = {
        { id = "collect_blooms", type = "collect_items", description = "Blooms gathered", count = 5, location_id = "crystal_forest" },
      },
    },
    {
      id = "deliver_blooms",
      name = "Deliver the Blooms",
      description = "Present the blooms to the half-elf diplomat.",
      requirements = { "gather_blooms" },
      objectives = {
        { id = "deliver", type = "deliver_item", description = "Deliver blooms", npc_id = "half_elf_diplomat" },
      },
    },
  },
  rewards = { experience = 100, gold = 30 },
})