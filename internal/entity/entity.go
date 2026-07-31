// Package entity defines the simulation entities, their attributes, and related behaviors.
package entity

import (
	"hash/fnv"
	"math/rand"

	"simuz/internal/items"
	"simuz/internal/relation"
	"simuz/internal/species"
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type EntityAI struct {
	Type         string   `json:"type"`
	ScriptIDs    []string `json:"script_ids,omitempty"`
	FactionID    string   `json:"faction_id"`
	AggroRange   float64  `json:"aggro_range"`
	AggroWarning int      `json:"aggro_warning"`
	HomeLocation string   `json:"home_location"`
	SleepCycle   string   `json:"sleep_cycle"` // "diurnal" (default), "nocturnal", "none" (never sleeps)
	Brave        bool     `json:"brave,omitempty"`
}

type ActivityType string

const (
	ActivityIdle     ActivityType = "idle"
	ActivitySleep    ActivityType = "sleeping"
	ActivityWork     ActivityType = "working"
	ActivitySit      ActivityType = "sitting"
	ActivityTravel   ActivityType = "traveling"
	ActivityTrade    ActivityType = "trading"
	ActivityPaint    ActivityType = "painting"
	ActivityEat      ActivityType = "eating"
	ActivityWorship  ActivityType = "worshipping"
	ActivityPatrol   ActivityType = "patrolling"
	ActivityDrink    ActivityType = "drinking"
	ActivityCraft    ActivityType = "crafting"
	ActivityMeditate ActivityType = "meditating"
	ActivityHunt     ActivityType = "hunting"
	ActivityGather   ActivityType = "gathering"
)

// Gender represents the biological sex of an entity.
const (
	GenderMale   = "male"
	GenderFemale = "female"
	GenderOther  = "other"
)

// Faction constants define true voluntary group memberships — cults, religions,
// political movements, and similar organized groups. Species and professions are
// tracked separately on the Entity (Species, Profession fields).
const (
	FactionCivilian = "civilian"
	FactionCult     = "cult"
	FactionDeity    = "deity"
)

type EntityActivity struct {
	Type      ActivityType `json:"type"`
	SinceTick uint64       `json:"since_tick"`
	UntilTick uint64       `json:"until_tick"`
}

type Equipment struct {
	Head    *items.ItemInstance `json:"head,omitempty"`
	Body    *items.ItemInstance `json:"body,omitempty"`
	Weapon  *items.ItemInstance `json:"weapon,omitempty"`
	Offhand *items.ItemInstance `json:"offhand,omitempty"`
	Feet    *items.ItemInstance `json:"feet,omitempty"`
	Hands   *items.ItemInstance `json:"hands,omitempty"`
	Neck    *items.ItemInstance `json:"neck,omitempty"`
	Finger1 *items.ItemInstance `json:"finger1,omitempty"`
	Finger2 *items.ItemInstance `json:"finger2,omitempty"`
}

type Entity struct {
	ID                   string                        `json:"id"`
	Name                 string                        `json:"name"`
	Species              string                        `json:"species"`
	Gender               string                        `json:"gender"`
	Profession           string                        `json:"profession"`
	Level                int                           `json:"level"`
	Age                  int                           `json:"age"`           // simulation ticks: increments by 1 per tick
	MaxAge               int                           `json:"max_age"`       // game-days: natural lifespan (converted to ticks at comparison)
	LastMealTick         int                           `json:"last_meal"`     // simulation tick of last meal
	Alive                bool                          `json:"alive"`
	Immortal             bool                          `json:"immortal"`
	Attributes           Attributes                    `json:"attributes"`
	Skills               map[string]int                `json:"skills"`
	MaxHP                int                           `json:"max_hp"`
	HP                   int                           `json:"hp"`
	MaxFP                int                           `json:"max_fp"`
	FP                   int                           `json:"fp"`
	XP                   int                           `json:"xp"`
	SkillProgressXP      map[string]int                `json:"skillProgressXP"`
	LocationID           string                        `json:"location_id"`
	Position             Position                      `json:"position"`
	Equipment            Equipment                     `json:"equipment"`
	Inventory            []items.ItemInstance          `json:"inventory"`
	AI                   EntityAI                      `json:"ai"`
	Activity             EntityActivity                `json:"activity"`
	Faction              string                        `json:"faction"`
	Conscious            bool                          `json:"conscious"`
	Effects              []ActiveEffect                `json:"effects,omitempty"`
	Flags                map[string]any                `json:"flags,omitempty"`
	Mood                 string                        `json:"mood,omitempty"`
	MoodModifiers        []MoodModifier                `json:"mood_modifiers,omitempty"`
	LeashedBy            string                        `json:"leashed_by,omitempty"`
	RescueState          string                        `json:"rescue_state,omitempty"`
	Reproduction         Reproduction                  `json:"reproduction,omitempty"`
	Relationships        map[string]EntityRelationship `json:"relationships,omitempty"`
	LastReproductionTick uint64                        `json:"last_reproduction_tick,omitempty"` // simulation tick
	KnockedOutTick       uint64                        `json:"knocked_out_tick,omitempty"`       // simulation tick
	TimeOfDeath          uint64                        `json:"timeOfDeath"`                      // simulation tick
	relation.Relation
	Memory     map[string]string `json:"memory,omitempty"`
	BioProfile *species.Species  `json:"bioProfile"`
}

func NewEntity(id, name, speciesID string, attrs Attributes, level int, rel relation.Relation) *Entity {
	maxHP := attrs.CON*2 + level*2
	maxFP := attrs.CON + attrs.STR/2
	if maxHP < 1 {
		maxHP = 1
	}
	if maxFP < 1 {
		maxFP = 1
	}
	maxAge := 0
	if sp, exists := species.GetByID(speciesID); exists {
		maxAge = sp.MaxAge
	}
	return &Entity{
		ID:            id,
		Name:          name,
		Species:       speciesID,
		Gender:        GetRndGender(),
		Profession:    "",
		Level:         level,
		Age:           0,
		MaxAge:        maxAge,
		LastMealTick:  0,
		Alive:         true,
		Conscious:     true,
		Attributes:    attrs,
		Skills:        make(map[string]int),
		MaxHP:         maxHP,
		HP:            maxHP,
		MaxFP:         maxFP,
		FP:            maxFP,
		Inventory:     make([]items.ItemInstance, 0),
		Flags:         make(map[string]any),
		Effects:       make([]ActiveEffect, 0),
		Relationships: make(map[string]EntityRelationship),
		AI: EntityAI{
			Type: "passive",
		},
		Relation: rel,
	}
}

func (e *Entity) GetID() string         { return e.ID }
func (e *Entity) GetFaction() string    { return e.Faction }
func (e *Entity) GetSpecies() string    { return e.Species }
func (e *Entity) GetProfession() string { return e.Profession }
func GetRndGender() string {
	choices := []string{GenderMale, GenderFemale, GenderOther}
	return choices[rand.Intn(len(choices))]
}

// fertilityKey returns a deterministic 0-99 value derived from the entity ID.
// Uses FNV-1a so it is stable across ticks, saves, and process restarts.
func fertilityKey(id string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(id))
	return h.Sum32() % 100
}

// CanSire returns true when the entity can father/get another entity pregnant.
// Males always can; "other" can; females cannot.
func (e *Entity) CanSire() bool {
	return e.Gender == GenderMale || e.Gender == GenderOther
}

// CanGetPregnant returns true when the entity can carry a pregnancy.
// Females: 95% fertile (5% infertile, deterministic per entity ID).
// Other:   50% can carry (deterministic per entity ID).
// Males:   never.
func (e *Entity) CanGetPregnant() bool {
	switch e.Gender {
	case GenderFemale:
		return fertilityKey(e.ID) < 95
	case GenderOther:
		return fertilityKey(e.ID) < 50
	default:
		return false
	}
}

// IsAdult returns true if the entity is old enough to reproduce.
// speed is the game speed (game-minutes per tick), used to convert AdultAge from game-days to ticks.
func (e *Entity) IsAdult(speed int) bool {
	if sp, exists := species.GetByID(e.Species); exists {
		adultAgeTicks := sp.AdultAge * 1440 / speed
		if adultAgeTicks < 1 {
			adultAgeTicks = 1
		}
		return e.Age >= adultAgeTicks || e.Level >= 3
	}
	return false
}
func (e *Entity) TakeDamage(amount int) {
	if amount < 0 {
		return
	}
	e.HP -= amount
	if e.HP <= 0 {
		e.HP = 0
		if !e.Immortal {
			// Dropping to 0 HP knocks out; combat decides whether to kill.
			e.Alive = true
			e.Conscious = false
		}
	}
}

// Kill permanently slays a non-immortal entity.
func (e *Entity) Kill() {
	if e.Immortal {
		return
	}
	e.HP = 0
	e.Alive = false
	e.Conscious = false
}

func (e *Entity) Heal(amount int) {
	if !e.Alive {
		return
	}
	e.HP += amount
	if e.HP > e.MaxHP {
		e.HP = e.MaxHP
	}
	if e.HP > 0 {
		e.Conscious = true
	}
}

func (e *Entity) SpendFP(amount int) bool {
	if e.FP < amount {
		return false
	}
	e.FP -= amount
	return true
}

func (e *Entity) RestFP(amount int) {
	e.FP += amount
	if e.FP > e.MaxFP {
		e.FP = e.MaxFP
	}
}

func (e *Entity) AddXP(amount int) {
	e.XP += amount
	e.CheckLevelUp()
}

func (e *Entity) IsBlessed(currentTick uint64) bool {
	expiry, ok := e.Flags["blessed"]
	if !ok {
		return false
	}
	exp, ok := expiry.(float64)
	if !ok {
		return false
	}
	return currentTick < uint64(exp)
}

func (e *Entity) Equip(item *items.ItemInstance) bool {
	if item.Equipped {
		return false
	}
	slot := item.Def.Slot
	if slot == "" {
		return false
	}
	item.Equipped = true
	switch slot {
	case "head":
		e.Equipment.Head = item
	case "body":
		e.Equipment.Body = item
	case "weapon":
		e.Equipment.Weapon = item
	case "offhand":
		e.Equipment.Offhand = item
	case "feet":
		e.Equipment.Feet = item
	case "hands":
		e.Equipment.Hands = item
	case "neck":
		e.Equipment.Neck = item
	case "finger":
		if e.Equipment.Finger1 == nil {
			e.Equipment.Finger1 = item
		} else if e.Equipment.Finger2 == nil {
			e.Equipment.Finger2 = item
		} else {
			item.Equipped = false
			return false
		}
	default:
		item.Equipped = false
		return false
	}
	return true
}

func (e *Entity) Unequip(slot string) *items.ItemInstance {
	var item *items.ItemInstance
	switch slot {
	case "head":
		item = e.Equipment.Head
		e.Equipment.Head = nil
	case "body":
		item = e.Equipment.Body
		e.Equipment.Body = nil
	case "weapon":
		item = e.Equipment.Weapon
		e.Equipment.Weapon = nil
	case "offhand":
		item = e.Equipment.Offhand
		e.Equipment.Offhand = nil
	case "feet":
		item = e.Equipment.Feet
		e.Equipment.Feet = nil
	case "hands":
		item = e.Equipment.Hands
		e.Equipment.Hands = nil
	case "neck":
		item = e.Equipment.Neck
		e.Equipment.Neck = nil
	case "finger1":
		item = e.Equipment.Finger1
		e.Equipment.Finger1 = nil
	case "finger2":
		item = e.Equipment.Finger2
		e.Equipment.Finger2 = nil
	}
	if item != nil {
		item.Equipped = false
	}
	return item
}

func (e *Entity) AddItem(itm items.ItemInstance) {
	e.Inventory = append(e.Inventory, itm)
}

func (e *Entity) RemoveItem(idx int) *items.ItemInstance {
	if idx < 0 || idx >= len(e.Inventory) {
		return nil
	}
	itm := e.Inventory[idx]
	e.Inventory = append(e.Inventory[:idx], e.Inventory[idx+1:]...)
	return &itm
}

func (e *Entity) Encumbrance() float64 {
	var total float64
	for _, itm := range e.Inventory {
		total += itm.Def.Weight
	}
	if e.Equipment.Head != nil {
		total += e.Equipment.Head.Def.Weight
	}
	if e.Equipment.Body != nil {
		total += e.Equipment.Body.Def.Weight
	}
	if e.Equipment.Weapon != nil {
		total += e.Equipment.Weapon.Def.Weight
	}
	if e.Equipment.Offhand != nil {
		total += e.Equipment.Offhand.Def.Weight
	}
	if e.Equipment.Feet != nil {
		total += e.Equipment.Feet.Def.Weight
	}
	return total
}

// AddRelationship records or updates a relationship between this entity and another.
func (e *Entity) AddRelationship(otherID string, relType RelationshipType, tick uint64) {
	e.Relationships[otherID] = EntityRelationship{
		OtherID:   otherID,
		Type:      relType,
		SinceTick: tick,
	}
}

// GetRelationship returns the relationship to another entity, if one exists.
func (e *Entity) GetRelationship(otherID string) (EntityRelationship, bool) {
	rel, ok := e.Relationships[otherID]
	return rel, ok
}

// GetRelationshipTo returns who the given relationship is to entity, returns only first entityID.
func (e *Entity) GetRelationshipTo(rst RelationshipType) (EntityRelationship, bool) {
	for _, val := range e.Relationships {
		if val.Type == rst {
			return val, true
		}
	}
	return EntityRelationship{}, false
}

// GetChildren returns all parent-child relationships from this entity's perspective.
func (e *Entity) GetChildren() []EntityRelationship {
	var result []EntityRelationship
	for _, rel := range e.Relationships {
		if rel.Type == RelationshipChild {
			result = append(result, rel)
		}
	}
	return result
}

// GetParents returns all parent relationships from this entity's perspective.
func (e *Entity) GetParents() []EntityRelationship {
	var result []EntityRelationship
	for _, rel := range e.Relationships {
		if rel.Type == RelationshipParent {
			result = append(result, rel)
		}
	}
	return result
}

// GetPartner returns the mate relationship, if one exists.
func (e *Entity) GetPartner() (EntityRelationship, bool) {
	for _, rel := range e.Relationships {
		if rel.Type == RelationshipMate {
			return rel, true
		}
	}
	return EntityRelationship{}, false
}
