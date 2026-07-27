package quest

type State string

const (
	StateInactive  State = "inactive"
	StateActive    State = "active"
	StateCompleted State = "completed"
	StateFailed    State = "failed"
)

type QuestType string

const (
	TypeMain      QuestType = "main"
	TypeSide      QuestType = "side"
	TypeFaction   QuestType = "faction"
	TypeDaily     QuestType = "daily"
	TypeRepeatable QuestType = "repeatable"
)

type QuestDef struct {
	ID            string              `yaml:"id"`
	Title         string              `yaml:"title"`
	Type          QuestType           `yaml:"type"`
	Level         int                 `yaml:"level"`
	Description   string              `yaml:"description,omitempty"`
	Source        *QuestSource        `yaml:"source,omitempty"`
	Prereqs       *Prerequisites      `yaml:"prerequisites,omitempty"`
	Stages        []StageDef          `yaml:"stages"`
	Rewards       *Rewards            `yaml:"rewards,omitempty"`
	FailConditions []FailCondition    `yaml:"failure_conditions,omitempty"`
}

type QuestSource struct {
	Type       string `yaml:"type"`
	NPCID      string `yaml:"npc_id,omitempty"`
	LocationID string `yaml:"location_id,omitempty"`
	Dialog     *QuestDialog `yaml:"dialog,omitempty"`
}

type QuestDialog struct {
	Accept   string `yaml:"accept"`
	Progress string `yaml:"progress"`
	Complete string `yaml:"complete"`
}

type Prerequisites struct {
	QuestsCompleted []string            `yaml:"quests_completed,omitempty"`
	QuestsActive    []string            `yaml:"quests_active,omitempty"`
	LevelMin        int                 `yaml:"level_min,omitempty"`
	LevelMax        int                 `yaml:"level_max,omitempty"`
	FactionRep      map[string]int      `yaml:"faction_reputation,omitempty"`
	Flags           []FlagCondition     `yaml:"flags,omitempty"`
}

type FlagCondition struct {
	Flag  string `yaml:"flag"`
	Value any    `yaml:"value,omitempty"`
}

type StageDef struct {
	ID           string        `yaml:"id"`
	Name         string        `yaml:"name"`
	Description  string        `yaml:"description"`
	Requirements []string      `yaml:"requirements,omitempty"`
	Objectives   []ObjectiveDef `yaml:"objectives"`
}

type ObjectiveDef struct {
	ID             string `yaml:"id"`
	Type           string `yaml:"type"`
	Description    string `yaml:"description"`
	Optional       bool   `yaml:"optional,omitempty"`
	Count          int    `yaml:"count,omitempty"`
	EntityTemplate string `yaml:"entity_template,omitempty"`
	LocationID     string `yaml:"location_id,omitempty"`
	NPCID          string `yaml:"npc_id,omitempty"`
	ItemTemplate   string `yaml:"item_template,omitempty"`
}

type Rewards struct {
	Experience int               `yaml:"experience,omitempty"`
	Gold       int               `yaml:"gold,omitempty"`
	Items      []RewardItem      `yaml:"items,omitempty"`
	FactionRep map[string]int    `yaml:"faction_reputation,omitempty"`
	Unlocks    *Unlocks          `yaml:"unlocks,omitempty"`
}

type RewardItem struct {
	Template string `yaml:"template"`
	Count    int    `yaml:"count"`
}

type Unlocks struct {
	Quests    []string `yaml:"quests,omitempty"`
	Locations []string `yaml:"locations,omitempty"`
	Recipes   []string `yaml:"recipes,omitempty"`
}

type FailCondition struct {
	Type     string `yaml:"type"`
	Hours    int    `yaml:"hours,omitempty"`
	EntityID string `yaml:"entity_id,omitempty"`
	Flag     string `yaml:"flag,omitempty"`
}
