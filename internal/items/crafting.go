// Package items defines item definitions, instances, and registry support for the simulation.
package items

type RecipeInput struct {
	DefID string
	Count int
}

type RecipeOutput struct {
	DefID string
	Count int
}

type Recipe struct {
	ID      string
	Name    string
	Inputs  []RecipeInput
	Output  RecipeOutput
	Station string
}

var Recipes = []*Recipe{
	{
		ID: "smelt_iron", Name: "Smelt Iron",
		Inputs:  []RecipeInput{{DefID: "iron_ore", Count: 2}, {DefID: "coal", Count: 1}},
		Output:  RecipeOutput{DefID: "iron_ingot", Count: 1},
		Station: "forge",
	},
	{
		ID: "craft_bandage", Name: "Craft Bandage",
		Inputs:  []RecipeInput{{DefID: "cloth", Count: 2}, {DefID: "leather_strips", Count: 1}},
		Output:  RecipeOutput{DefID: "bandage", Count: 2},
		Station: "workbench",
	},
	{
		ID: "cook_poultice", Name: "Prepare Herbal Poultice",
		Inputs:  []RecipeInput{{DefID: "herb", Count: 2}},
		Output:  RecipeOutput{DefID: "herbal_poultice", Count: 1},
		Station: "campfire",
	},
	{
		ID: "refine_salve", Name: "Refine Healing Salve",
		Inputs:  []RecipeInput{{DefID: "herb", Count: 4}, {DefID: "coal", Count: 1}},
		Output:  RecipeOutput{DefID: "healing_salve", Count: 1},
		Station: "cauldron",
	},
}

func HasMaterials(inv []ItemInstance, inputs []RecipeInput) bool {
	for _, req := range inputs {
		if countInInventory(inv, req.DefID) < req.Count {
			return false
		}
	}
	return true
}

func countInInventory(inv []ItemInstance, defID string) int {
	total := 0
	for _, inst := range inv {
		if inst.DefID == defID && !inst.Equipped {
			total += inst.Count
		}
	}
	return total
}

func RemoveInputs(inv []ItemInstance, inputs []RecipeInput) []ItemInstance {
	for _, req := range inputs {
		remaining := req.Count
		for i := 0; i < len(inv) && remaining > 0; i++ {
			if inv[i].DefID == req.DefID && !inv[i].Equipped {
				if inv[i].Count <= remaining {
					remaining -= inv[i].Count
					inv = append(inv[:i], inv[i+1:]...)
					i--
				} else {
					inv[i].Count -= remaining
					remaining = 0
				}
			}
		}
	}
	return inv
}
