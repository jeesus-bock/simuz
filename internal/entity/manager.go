// Package entity defines the simulation entities, their attributes, and related behaviors.
package entity

import "sync"

type Manager struct {
	mu       sync.RWMutex
	entities map[string]*Entity
}

var EntityManager *Manager

func NewManager() *Manager {
	EntityManager = &Manager{
		entities: make(map[string]*Entity),
	}
	return EntityManager
}

func (m *Manager) Add(e *Entity) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entities[e.ID] = e
}

func (m *Manager) Get(id string) *Entity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.entities[id]
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entities, id)
}

func (m *Manager) All() []*Entity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Entity, 0, len(m.entities))
	for _, e := range m.entities {
		result = append(result, e)
	}
	return result
}

func (m *Manager) ByLocation(locID string) []*Entity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Entity
	for _, e := range m.entities {
		if e.LocationID == locID {
			result = append(result, e)
		}
	}
	return result
}

func (m *Manager) Filter(fn func(*Entity) bool) []*Entity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Entity
	for _, e := range m.entities {
		if fn(e) {
			result = append(result, e)
		}
	}
	return result
}

func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entities)
}
