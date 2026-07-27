package quest

import (
	"testing"
)

func TestLoadScripts(t *testing.T) {
	defs, err := LoadScripts()
	if err != nil {
		t.Fatalf("LoadScripts: %v", err)
	}
	if len(defs) < 12 {
		t.Fatalf("expected at least 12 quests, got %d", len(defs))
	}

	byID := map[string]*QuestDef{}
	for _, d := range defs {
		if d.ID == "" {
			t.Fatal("quest with empty id")
		}
		if _, dup := byID[d.ID]; dup {
			t.Fatalf("duplicate quest id %s", d.ID)
		}
		byID[d.ID] = d
		if len(d.Stages) == 0 {
			t.Fatalf("quest %s has no stages", d.ID)
		}
		for _, st := range d.Stages {
			if st.ID == "" {
				t.Fatalf("quest %s has stage with empty id", d.ID)
			}
			if len(st.Objectives) == 0 {
				t.Fatalf("quest %s stage %s has no objectives", d.ID, st.ID)
			}
		}
	}

	required := []string{
		"rat_problem",
		"deliver_sword",
		"lost_heirlooms",
		"deity_whispers",
		"freya_favor",
		"zeus_crazy_task",
		"kobold_menace",
		"vampire_hunt",
		"hag_curse",
		"fairy_escort",
		"bard_ballad",
		"taken_courier",
	}
	for _, id := range required {
		if byID[id] == nil {
			t.Errorf("missing quest %s", id)
		}
	}

	rat := byID["rat_problem"]
	if rat.Title != "The Rat Problem" {
		t.Errorf("rat_problem title = %q", rat.Title)
	}
	if rat.Source == nil || rat.Source.NPCID != "frosthold_greta" {
		t.Fatalf("rat_problem source = %+v", rat.Source)
	}
	if rat.Rewards == nil || rat.Rewards.Experience != 50 || rat.Rewards.Gold != 10 {
		t.Fatalf("rat_problem rewards = %+v", rat.Rewards)
	}
	kill := rat.Stages[0].Objectives[0]
	if kill.Type != "kill_entities" || kill.Count != 5 || kill.EntityTemplate != "rat" {
		t.Fatalf("rat kill objective = %+v", kill)
	}
	talk := rat.Stages[1].Objectives[0]
	if talk.Type != "talk_to_npc" || talk.Count != 1 || talk.NPCID != "frosthold_greta" {
		t.Fatalf("rat talk objective = %+v", talk)
	}

	taken := byID["taken_courier"]
	if taken.Title != "The Taken Courier" {
		t.Errorf("taken_courier title = %q", taken.Title)
	}
	if taken.Source == nil || taken.Source.NPCID != "frosthold_guard_captain" {
		t.Fatalf("taken_courier source = %+v", taken.Source)
	}
	if taken.Rewards == nil || taken.Rewards.Experience != 120 || taken.Rewards.Gold != 25 {
		t.Fatalf("taken_courier rewards = %+v", taken.Rewards)
	}
	kobold := taken.Stages[1].Objectives[0]
	if kobold.Type != "kill_entities" || kobold.Count != 3 || kobold.EntityTemplate != "kobold" {
		t.Fatalf("taken_courier kobold objective = %+v", kobold)
	}
}
