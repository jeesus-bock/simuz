-- The Fairy's Request
-- Help Sparkle gather magical ingredients.

quest.define({
  id = "fairy_escort",
  title = "The Fairy's Request",
  type = "side",
  level = 2,
  description = "A fairy named Sparkle asks for help gathering magical ingredients from the Crystal Forest.",
  source = { type = "npc", npc_id = "fairy_sparkle" },
  stages = {
    {
      id = "gather_ingredients",
      name = "Gather Ingredients",
      description = "Collect rare herbs and flowers from the forest.",
      objectives = {
        { id = "collect_herbs", type = "collect_items", description = "Herbs gathered", count = 3, location_id = "crystal_forest" },
      },
    },
    {
      id = "deliver_fairy",
      name = "Deliver to Sparkle",
      description = "Bring the ingredients to the fairy.",
      requirements = { "gather_ingredients" },
      objectives = {
        { id = "deliver_fairy", type = "deliver_item", description = "Deliver ingredients", npc_id = "fairy_sparkle" },
      },
    },
  },
  rewards = { experience = 60, gold = 10 },
})
