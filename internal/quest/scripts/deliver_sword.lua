-- The Iron Anvil Delivery
-- Sven wants a sword delivered to Captain Halvar.

quest.define({
  id = "deliver_sword",
  title = "The Iron Anvil Delivery",
  type = "side",
  level = 1,
  description = "Sven the blacksmith in Frosthold has crafted a fine sword for Captain Halvar at the guardhouse. Deliver it.",
  source = { type = "npc", npc_id = "frosthold_sven" },
  stages = {
    {
      id = "pickup",
      name = "Collect the Sword",
      description = "Pick up the sword from Sven at the blacksmith.",
      objectives = {
        { id = "collect_sword", type = "collect_items", description = "Collect the sword", count = 1, item_template = "iron_sword" },
      },
    },
    {
      id = "deliver",
      name = "Deliver to Halvar",
      description = "Bring the sword to Captain Halvar at the guardhouse.",
      requirements = { "pickup" },
      objectives = {
        {
          id = "deliver_sword",
          type = "deliver_item",
          description = "Deliver the sword",
          npc_id = "frosthold_guard_captain",
          item_template = "iron_sword",
        },
      },
    },
  },
  rewards = { experience = 35, gold = 5 },
})

return {
  util.event("quest_accepted", { quest_id = "deliver_sword", source = "frosthold_sven" }),
}
