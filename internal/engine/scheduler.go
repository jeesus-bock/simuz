// Package engine contains the simulation engine, tick processing, and related systems.
package engine

type ScheduledEvent struct {
	ID       string
	Tick     uint64
	Interval uint64
	Action   func()
	OneShot  bool
}

type Scheduler struct {
	events []ScheduledEvent
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		events: make([]ScheduledEvent, 0),
	}
}

func (s *Scheduler) Add(event ScheduledEvent) {
	s.events = append(s.events, event)
}

func (s *Scheduler) Remove(id string) {
	for i, e := range s.events {
		if e.ID == id {
			s.events = append(s.events[:i], s.events[i+1:]...)
			return
		}
	}
}

func (s *Scheduler) ProcessDue(currentTick uint64) {
	for i := 0; i < len(s.events); i++ {
		e := s.events[i]
		if currentTick >= e.Tick {
			if e.Action != nil {
				e.Action()
			}
			if e.Interval > 0 {
				s.events[i].Tick = currentTick + e.Interval
			} else {
				s.events = append(s.events[:i], s.events[i+1:]...)
				i--
			}
		}
	}
}

func (s *Scheduler) At(tick uint64, action func()) string {
	id := "evt_" + formatTick(tick)
	s.Add(ScheduledEvent{
		ID:      id,
		Tick:    tick,
		Action:  action,
		OneShot: true,
	})
	return id
}

func (s *Scheduler) Every(interval uint64, action func()) string {
	id := "evt_every_" + formatTick(interval)
	s.Add(ScheduledEvent{
		ID:       id,
		Tick:     interval,
		Interval: interval,
		Action:   action,
	})
	return id
}
