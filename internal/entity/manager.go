// Package entity defines the simulation entities, their attributes, and related behaviors.
package entity

import "sync"

type EntityManager struct {
	mu       sync.RWMutex
	entities map[string]*Entity
}

var EntityMemgtanager *EntityManager

func NewEntityManager() *EntityManager {
	var ems = &EntityManager{
		entities: make(map[string]*Entity),
	}
	return ems
}
func (m *EntityManager) Add(e *Entity) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entities[e.ID] = e
}

func (m *EntityManager) Get(id string) *Entity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.entities[id]
}

func (m *EntityManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entities, id)
}

func (m *EntityManager) All() []*Entity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Entity, 0, len(m.entities))
	for _, e := range m.entities {
		result = append(result, e)
	}
	return result
}

func (m *EntityManager) ByLocation(locID string) []*Entity {
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

func (m *EntityManager) Filter(fn func(*Entity) bool) []*Entity {
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

func (m *EntityManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entities)
}

func (m *EntityManager) GetEntitiesInLocation(locationID string) []*Entity {
	ret := make([]*Entity, 0)
	for _, es := range m.entities {
		if es.LocationID == locationID {
			ret = append(ret, es)
		}
	}
	return ret
}
