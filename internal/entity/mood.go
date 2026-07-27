package entity

type MoodModifier struct {
	Source      string `json:"source"`
	Mood        string `json:"mood"`
	DecayAtTick uint64 `json:"decay_at_tick"`
}

var moodStatMods = map[string]Attributes{
	"neutral":  {},
	"happy":    {CHA: 1},
	"angry":    {STR: 2, WIS: -1},
	"fearful":  {STR: -2, DEX: 2},
	"stressed": {CON: -1, WIS: -1},
	"relaxed":  {CON: 1, WIS: 1},
	"inspired": {INT: 2, WIS: 2},
	"tired":    {STR: -1, DEX: -1, CON: -1},
}

func (e *Entity) AddMoodModifier(source, mood string, duration uint64) {
	if duration == 0 {
		return
	}
	e.MoodModifiers = append(e.MoodModifiers, MoodModifier{
		Source:      source,
		Mood:        mood,
		DecayAtTick: uint64(e.Age) + duration,
	})
}

func (e *Entity) EffectiveMood() string {
	if len(e.MoodModifiers) == 0 {
		return "neutral"
	}
	counts := make(map[string]int)
	for _, m := range e.MoodModifiers {
		counts[m.Mood]++
	}
	best := "neutral"
	bestCount := 0
	for mood, c := range counts {
		if c > bestCount {
			best = mood
			bestCount = c
		}
	}
	return best
}

func (e *Entity) MoodStatMods() Attributes {
	var total Attributes
	for _, m := range e.MoodModifiers {
		if mods, ok := moodStatMods[m.Mood]; ok {
			total.STR += mods.STR
			total.DEX += mods.DEX
			total.CON += mods.CON
			total.INT += mods.INT
			total.WIS += mods.WIS
			total.CHA += mods.CHA
		}
	}
	return total
}

func (e *Entity) TickMoods(tick uint64) {
	var active []MoodModifier
	for _, m := range e.MoodModifiers {
		if tick < m.DecayAtTick {
			active = append(active, m)
		}
	}
	e.MoodModifiers = active
	if len(e.MoodModifiers) > 0 {
		e.Mood = e.EffectiveMood()
	} else {
		e.Mood = "neutral"
	}
}
