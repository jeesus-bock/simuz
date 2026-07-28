// Package quest defines quest definitions, state handling, and quest progression logic.
package quest

import (
	"log"
	"sort"
	"strconv"
	"sync"
)

type EntityQuestState struct {
	QuestID         string          `json:"quest_id"`
	State           State           `json:"state"`
	CurrentStage    string          `json:"current_stage,omitempty"`
	CompletedStages []string        `json:"completed_stages,omitempty"`
	Objectives      map[string]int  `json:"objectives,omitempty"`
	Variables       map[string]any  `json:"variables,omitempty"`
	AcceptedTick    uint64          `json:"accepted_tick,omitempty"`
	Activity        []QuestActivity `json:"activity,omitempty"`
}

// QuestActivity records an observable decision or outcome while an entity is
// working on a quest. It is kept on the quest state so it can be shown in the
// UI and survive a save/load cycle.
type QuestActivity struct {
	Tick    uint64 `json:"tick"`
	Message string `json:"message"`
}

type CompleteFn func(string, string, *Rewards)

type Manager struct {
	mu              sync.RWMutex
	defs            map[string]*QuestDef
	states          map[string]map[string]*EntityQuestState
	OnQuestComplete CompleteFn
	currentTick     uint64
}

func (m *Manager) SetTick(tick uint64) {
	m.mu.Lock()
	m.currentTick = tick
	m.mu.Unlock()
}

func NewManager() *Manager {
	return &Manager{
		defs:   make(map[string]*QuestDef),
		states: make(map[string]map[string]*EntityQuestState),
	}
}

func (m *Manager) Register(def *QuestDef) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defs[def.ID] = def
}

func (m *Manager) GetDef(id string) *QuestDef {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.defs[id]
}

func (m *Manager) AllDefs() []*QuestDef {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*QuestDef, 0, len(m.defs))
	for _, def := range m.defs {
		result = append(result, def)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].ID != result[j].ID {
			return result[i].ID < result[j].ID
		}
		return result[i].Title < result[j].Title
	})
	return result
}

func (m *Manager) GetState(entityID, questID string) *EntityQuestState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if states, ok := m.states[entityID]; ok {
		return states[questID]
	}
	return nil
}

func (m *Manager) Accept(entityID, questID string, entityLevel int, currentTick uint64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	def := m.defs[questID]
	if def == nil {
		return false
	}

	if !m.canAcceptLocked(entityID, questID, entityLevel) {
		return false
	}

	if _, ok := m.states[entityID]; !ok {
		m.states[entityID] = make(map[string]*EntityQuestState)
	}

	m.states[entityID][questID] = &EntityQuestState{
		QuestID:         questID,
		State:           StateActive,
		Objectives:      make(map[string]int),
		Variables:       make(map[string]any),
		CompletedStages: make([]string, 0),
		AcceptedTick:    currentTick,
		Activity:        []QuestActivity{{Tick: currentTick, Message: "Quest accepted"}},
	}

	if len(def.Stages) > 0 {
		m.states[entityID][questID].CurrentStage = def.Stages[0].ID
	}

	return true
}

func (m *Manager) canAcceptLocked(entityID, questID string, entityLevel int) bool {
	def := m.defs[questID]
	if def == nil {
		return false
	}

	if states, ok := m.states[entityID]; ok {
		if _, ok := states[questID]; ok {
			return false
		}
	}

	if def.Prereqs != nil {
		if def.Prereqs.LevelMin > 0 && entityLevel < def.Prereqs.LevelMin {
			return false
		}
		if def.Prereqs.LevelMax > 0 && entityLevel > def.Prereqs.LevelMax {
			return false
		}
	}

	return true
}

func (m *Manager) CanAccept(entityID, questID string, entityLevel int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	def := m.defs[questID]
	if def == nil {
		return false
	}

	if states, ok := m.states[entityID]; ok {
		if _, ok := states[questID]; ok {
			return false
		}
	}

	if def.Prereqs != nil {
		if def.Prereqs.LevelMin > 0 && entityLevel < def.Prereqs.LevelMin {
			return false
		}
		if def.Prereqs.LevelMax > 0 && entityLevel > def.Prereqs.LevelMax {
			return false
		}
	}

	return true
}

func (m *Manager) EntityStates(entityID string) []*EntityQuestState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := m.states[entityID]
	result := make([]*EntityQuestState, 0, len(states))
	for _, s := range states {
		result = append(result, s)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].AcceptedTick != result[j].AcceptedTick {
			return result[i].AcceptedTick < result[j].AcceptedTick
		}
		if result[i].QuestID != result[j].QuestID {
			return result[i].QuestID < result[j].QuestID
		}
		return result[i].CurrentStage < result[j].CurrentStage
	})
	return result
}

func (m *Manager) LoadState(entityID, questID string, state State, currentStage string, completedStages []string, objectives map[string]int, variables map[string]any, acceptedTick uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.states[entityID]; !ok {
		m.states[entityID] = make(map[string]*EntityQuestState)
	}

	m.states[entityID][questID] = &EntityQuestState{
		QuestID:         questID,
		State:           state,
		CurrentStage:    currentStage,
		CompletedStages: completedStages,
		Objectives:      objectives,
		Variables:       variables,
		AcceptedTick:    acceptedTick,
	}
}

func (m *Manager) ProgressObjective(entityID, questID, objID string, delta int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.states[entityID][questID]
	if !ok || state.State != StateActive {
		return
	}

	state.Objectives[objID] += delta
	m.addActivityLocked(state, "Progressed objective '"+objID+"' by "+itoa(delta))
	m.checkStageCompletion(entityID, questID)
}

func (m *Manager) SetObjective(entityID, questID, objID string, value int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.states[entityID][questID]
	if !ok || state.State != StateActive {
		return
	}

	state.Objectives[objID] = value
	m.addActivityLocked(state, "Set objective '"+objID+"' to "+itoa(value))
	m.checkStageCompletion(entityID, questID)
}

func (m *Manager) CheckCollectItem(entityID, itemID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	states := m.states[entityID]
	for _, state := range states {
		if state.State != StateActive {
			continue
		}
		def := m.defs[state.QuestID]
		if def == nil {
			continue
		}
		for _, stage := range def.Stages {
			if stage.ID != state.CurrentStage {
				continue
			}
			for _, obj := range stage.Objectives {
				if obj.Type == "collect_items" && obj.ItemTemplate == itemID {
					state.Objectives[obj.ID]++
					m.addActivityLocked(state, "Collected '"+itemID+"' for objective '"+obj.ID+"'")
					m.checkStageCompletion(entityID, state.QuestID)
				}
			}
		}
	}
}

func (m *Manager) CheckVisitLocation(entityID, locationID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	states := m.states[entityID]
	for _, state := range states {
		if state.State != StateActive {
			continue
		}
		def := m.defs[state.QuestID]
		if def == nil {
			continue
		}
		for _, stage := range def.Stages {
			if stage.ID != state.CurrentStage {
				continue
			}
			for _, obj := range stage.Objectives {
				if obj.Type == "visit_location" && obj.LocationID == locationID {
					state.Objectives[obj.ID] = 1
					m.addActivityLocked(state, "Visited '"+locationID+"' for objective '"+obj.ID+"'")
					m.checkStageCompletionLocked(entityID, state.QuestID)
				}
			}
		}
	}
}

func (m *Manager) CheckTalkToNPC(entityID, npcID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	states := m.states[entityID]
	for _, state := range states {
		if state.State != StateActive {
			continue
		}
		def := m.defs[state.QuestID]
		if def == nil {
			continue
		}
		for _, stage := range def.Stages {
			if stage.ID != state.CurrentStage {
				continue
			}
			for _, obj := range stage.Objectives {
				if obj.Type == "talk_to_npc" && obj.NPCID == npcID {
					state.Objectives[obj.ID] = 1
					m.addActivityLocked(state, "Talked to '"+npcID+"' for objective '"+obj.ID+"'")
					m.checkStageCompletionLocked(entityID, state.QuestID)
				}
			}
		}
	}
}

func (m *Manager) CheckDeliverItem(entityID, npcID, itemID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	states := m.states[entityID]
	for _, state := range states {
		if state.State != StateActive {
			continue
		}
		def := m.defs[state.QuestID]
		if def == nil {
			continue
		}
		for _, stage := range def.Stages {
			if stage.ID != state.CurrentStage {
				continue
			}
			for _, obj := range stage.Objectives {
				if obj.Type == "deliver_item" && obj.NPCID == npcID && obj.ItemTemplate == itemID {
					state.Objectives[obj.ID] = 1
					m.addActivityLocked(state, "Delivered '"+itemID+"' to '"+npcID+"' for objective '"+obj.ID+"'")
					m.checkStageCompletionLocked(entityID, state.QuestID)
				}
			}
		}
	}
}

func (m *Manager) triggerUnlocks(def *QuestDef, entityID string) {
	if def.Rewards == nil || def.Rewards.Unlocks == nil {
		return
	}
	for _, qid := range def.Rewards.Unlocks.Quests {
		if states, ok := m.states[entityID]; ok {
			if _, exists := states[qid]; exists {
				continue
			}
		} else {
			m.states[entityID] = make(map[string]*EntityQuestState)
		}
		m.states[entityID][qid] = &EntityQuestState{
			QuestID:         qid,
			State:           StateInactive,
			Objectives:      make(map[string]int),
			Variables:       make(map[string]any),
			CompletedStages: make([]string, 0),
			Activity:        []QuestActivity{{Tick: m.currentTick, Message: "Quest unlocked"}},
		}
		log.Printf("[quest] unlocked quest '%s' for %s", qid, entityID)
	}
}

func (m *Manager) addActivityLocked(state *EntityQuestState, message string) {
	state.Activity = append(state.Activity, QuestActivity{Tick: m.currentTick, Message: message})
	if len(state.Activity) > 50 {
		state.Activity = state.Activity[len(state.Activity)-50:]
	}
}

func itoa(n int) string { return strconv.Itoa(n) }

func (m *Manager) RecentActivity(entityID, questID string) []QuestActivity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state := m.states[entityID][questID]
	if state == nil || len(state.Activity) == 0 {
		return nil
	}
	out := make([]QuestActivity, len(state.Activity))
	copy(out, state.Activity)
	return out
}

// RecordActivity adds a narrative line to every active quest an entity is
// pursuing. Callers use this for attempts such as travel, talking, and
// delivery, including attempts that do not advance an objective.
func (m *Manager) RecordActivity(entityID, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, state := range m.states[entityID] {
		if state.State == StateActive {
			m.addActivityLocked(state, message)
		}
	}
}

func (m *Manager) checkStageCompletion(entityID, questID string) {
	def := m.defs[questID]
	if def == nil {
		return
	}

	state := m.states[entityID][questID]
	if state == nil {
		return
	}

	stageID := state.CurrentStage
	var stage *StageDef
	for i := range def.Stages {
		if def.Stages[i].ID == stageID {
			stage = &def.Stages[i]
			break
		}
	}
	if stage == nil {
		return
	}

	for _, req := range stage.Requirements {
		found := false
		for _, cs := range state.CompletedStages {
			if cs == req {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	for _, obj := range stage.Objectives {
		progress := state.Objectives[obj.ID]
		if progress < obj.Count {
			return
		}
	}

	state.CompletedStages = append(state.CompletedStages, stageID)

	nextIdx := -1
	for i, s := range def.Stages {
		if s.ID == stageID {
			nextIdx = i + 1
			break
		}
	}

	if nextIdx >= 0 && nextIdx < len(def.Stages) {
		state.CurrentStage = def.Stages[nextIdx].ID
	} else {
		state.State = StateCompleted
		if m.OnQuestComplete != nil {
			m.OnQuestComplete(entityID, questID, def.Rewards)
		}
	}
	m.triggerUnlocks(def, entityID)
}
