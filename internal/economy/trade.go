// Package economy contains trade pricing and economy-related helpers used by the simulation.
package economy

import (
	"math/rand"
	"sort"

	"simuz/internal/entity"
	"simuz/internal/items"
)

const (
	SellerMarkup  = 1.5
	BuyerDiscount = 0.6
)

func haggleThreshold(savvy, slick int) int {
	return savvy - slick
}

func SellPrice(base int, seller, buyer *entity.Entity) (int, bool) {
	asking := max(1, int(float64(base)*SellerMarkup+0.5))
	buyerAttrs := buyer.EffectiveAttrs()
	sellerAttrs := seller.EffectiveAttrs()
	savvy := buyerAttrs.INT + buyerAttrs.WIS
	slick := sellerAttrs.CHA + sellerAttrs.INT
	diff := haggleThreshold(savvy, slick)
	if diff >= 5 {
		fair := max(1, base)
		return fair, true
	}
	if diff >= 0 {
		compromise := max(1, (asking+base)/2)
		return compromise, true
	}
	return asking, false
}

func BuyPrice(base int, seller, buyer *entity.Entity) (int, bool) {
	offer := max(1, int(float64(base)*BuyerDiscount+0.5))
	sellerAttrs := seller.EffectiveAttrs()
	buyerAttrs := buyer.EffectiveAttrs()
	savvy := sellerAttrs.INT + sellerAttrs.WIS
	slick := buyerAttrs.CHA + buyerAttrs.INT
	diff := haggleThreshold(savvy, slick)
	if diff >= 5 {
		fair := max(1, base)
		return fair, true
	}
	if diff >= 0 {
		compromise := max(1, (offer+base)/2)
		return compromise, true
	}
	return offer, false
}

func TotalCurrency(e *entity.Entity) int {
	total := 0
	for _, inst := range e.Inventory {
		if inst.Def != nil && inst.Def.Type == items.TypeCurrency {
			total += inst.Def.Value * inst.Count
		}
	}
	return total
}

func RemoveCurrency(e *entity.Entity, amount int) int {
	type currSlot struct {
		idx   int
		value int
		count int
	}
	var slots []currSlot
	for i, inst := range e.Inventory {
		if inst.Def != nil && inst.Def.Type == items.TypeCurrency && inst.Def.Value > 0 {
			slots = append(slots, currSlot{idx: i, value: inst.Def.Value, count: inst.Count})
		}
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i].idx > slots[j].idx
	})
	remaining := amount
	removed := 0
	for _, s := range slots {
		if remaining <= 0 {
			break
		}
		perUnit := s.value
		needed := (remaining + perUnit - 1) / perUnit
		if needed > s.count {
			needed = s.count
		}
		takeValue := needed * perUnit
		remaining -= takeValue
		removed += takeValue
		if needed >= s.count {
			e.RemoveItem(s.idx)
		} else {
			e.Inventory[s.idx].Count -= needed
		}
	}
	return removed
}

var coinDefs = []struct {
	id    string
	value int
}{
	{"ep", 100000},
	{"mp", 10000},
	{"tp", 1000},
	{"gp", 100},
	{"sp", 10},
	{"cp", 1},
}

func AddCurrency(e *entity.Entity, amount int) {
	if amount <= 0 {
		return
	}
	for _, cd := range coinDefs {
		if amount >= cd.value {
			count := amount / cd.value
			amount -= count * cd.value
			def := &items.ItemDef{
				ID:        cd.id,
				Name:      coinName(cd.id),
				Type:      items.TypeCurrency,
				Weight:    0.01,
				Value:     cd.value,
				Stackable: true,
				MaxStack:  99999,
			}
			addOrStack(e, def, count)
		}
	}
}

func coinName(id string) string {
	switch id {
	case "cp":
		return "Copper Piece"
	case "sp":
		return "Silver Piece"
	case "gp":
		return "Gold Piece"
	case "tp":
		return "Titanium Piece"
	case "mp":
		return "Mithril Piece"
	case "ep":
		return "Electronium Piece"
	}
	return id
}

func TransferItem(from, to *entity.Entity, idx int) bool {
	if idx < 0 || idx >= len(from.Inventory) {
		return false
	}
	inst := from.Inventory[idx]
	if inst.Equipped {
		return false
	}
	from.RemoveItem(idx)
	to.AddItem(inst)
	return true
}

func addOrStack(e *entity.Entity, def *items.ItemDef, count int) {
	for i := range e.Inventory {
		if e.Inventory[i].DefID == def.ID && e.Inventory[i].Def != nil && e.Inventory[i].Def.Stackable {
			e.Inventory[i].Count += count
			return
		}
	}
	e.AddItem(items.ItemInstance{
		ID:    def.ID + "_stack",
		DefID: def.ID,
		Def:   def,
		Count: count,
	})
}

func CanAfford(e *entity.Entity, price int) bool {
	return TotalCurrency(e) >= price
}

func HasItem(e *entity.Entity, defID string) (int, bool) {
	for i, inst := range e.Inventory {
		if inst.DefID == defID && !inst.Equipped {
			return i, true
		}
	}
	return -1, false
}

func GenerateWares(rng *rand.Rand, count int) []string {
	allTrade := []string{
		"common_clothes", "fine_clothes", "simple_robe", "work_tunic",
		"dagger", "short_sword", "cudgel", "tankard",
		"leather_helmet", "leather_boots", "leather_gloves",
		"wooden_shield", "holy_symbol",
	}
	rng.Shuffle(len(allTrade), func(i, j int) {
		allTrade[i], allTrade[j] = allTrade[j], allTrade[i]
	})
	if count > len(allTrade) {
		count = len(allTrade)
	}
	return allTrade[:count]
}
