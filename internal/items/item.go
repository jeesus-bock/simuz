// Package items defines item definitions, instances, and registry support for the simulation.
package items

type ItemType int

const (
	TypeMisc ItemType = iota
	TypeWeapon
	TypeArmor
	TypeConsumable
	TypeMaterial
	TypeQuest
	TypeCurrency
)

type ItemDef struct {
	ID          string           `json:"id" yaml:"id"`
	Name        string           `json:"name" yaml:"name"`
	Type        ItemType         `json:"type" yaml:"type"`
	Weight      float64          `json:"weight" yaml:"weight"`
	Value       int              `json:"value" yaml:"value"`
	Description string           `json:"description" yaml:"description"`
	Slot        string           `json:"slot,omitempty" yaml:"slot,omitempty"`
	Tags        []string         `json:"tags,omitempty" yaml:"tags,omitempty"`
	Stackable   bool             `json:"stackable" yaml:"stackable"`
	MaxStack    int              `json:"max_stack,omitempty" yaml:"max_stack,omitempty"`
	Substance   *SubstanceEffect `json:"substance,omitempty" yaml:"substance,omitempty"`
}

type ItemInstance struct {
	ID        string   `json:"id"`
	DefID     string   `json:"def_id"`
	Def       *ItemDef `json:"-"`
	Count     int      `json:"count"`
	Condition float64  `json:"condition"`
	Equipped  bool     `json:"equipped"`
}

func NewItemInstance(id, defID string, def *ItemDef, count int) ItemInstance {
	return ItemInstance{
		ID:        id,
		DefID:     defID,
		Def:       def,
		Count:     count,
		Condition: 1.0,
	}
}

type ItemRegistry struct {
	defs map[string]*ItemDef
}

var globalDefs = make(map[string]*ItemDef)

func RegisterDef(def *ItemDef) {
	globalDefs[def.ID] = def
}

func GetDef(id string) *ItemDef {
	return globalDefs[id]
}

func NewItemRegistry() *ItemRegistry {
	return &ItemRegistry{
		defs: make(map[string]*ItemDef),
	}
}

func (r *ItemRegistry) Register(def *ItemDef) {
	r.defs[def.ID] = def
}

func (r *ItemRegistry) Get(id string) *ItemDef {
	return r.defs[id]
}

func (r *ItemRegistry) All() []*ItemDef {
	result := make([]*ItemDef, 0, len(r.defs))
	for _, def := range r.defs {
		result = append(result, def)
	}
	return result
}

func (r *ItemRegistry) ByType(t ItemType) []*ItemDef {
	var result []*ItemDef
	for _, def := range r.defs {
		if def.Type == t {
			result = append(result, def)
		}
	}
	return result
}
