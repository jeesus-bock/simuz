// Package storage defines the persistence interfaces and SQLite-backed implementation used by the simulation.
package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"simuz/internal/engine"
	"simuz/internal/entity"
	"simuz/internal/items"
	"simuz/internal/quest"
	"simuz/internal/relation"
	"simuz/internal/world"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db      *sql.DB
	enabled bool
	path    string
}

func NewSQLiteStore(path string) *SQLiteStore {
	return &SQLiteStore{
		path:    path,
		enabled: false,
	}
}

func (s *SQLiteStore) Open() error {
	db, err := sql.Open("sqlite", s.path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	s.db = db
	s.enabled = true

	if err := s.migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLiteStore) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS world_state (
			id INTEGER PRIMARY KEY DEFAULT 1,
			tick INTEGER NOT NULL DEFAULT 0,
			day INTEGER NOT NULL DEFAULT 1,
			hour INTEGER NOT NULL DEFAULT 6,
			minute INTEGER NOT NULL DEFAULT 0,
			speed INTEGER NOT NULL DEFAULT 24,
			faction_relations_json TEXT NOT NULL DEFAULT '{}'
		)`,
		`CREATE TABLE IF NOT EXISTS locations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type INTEGER NOT NULL,
			parent_id TEXT,
			pos_x REAL NOT NULL DEFAULT 0,
			pos_y REAL NOT NULL DEFAULT 0,
			area REAL NOT NULL DEFAULT 0,
			is_outside INTEGER NOT NULL DEFAULT 1,
			weather_json TEXT,
			tags_json TEXT,
			controlling_faction TEXT DEFAULT '',
			control_strength INTEGER DEFAULT 0,
			exits_json TEXT DEFAULT '[]'
		)`,
		`CREATE TABLE IF NOT EXISTS entities (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			species TEXT NOT NULL DEFAULT 'human',
			level INTEGER NOT NULL DEFAULT 1,
			age INTEGER NOT NULL DEFAULT 0,
			max_age INTEGER NOT NULL DEFAULT 0,
			last_meal INTEGER NOT NULL DEFAULT 0,
			alive INTEGER NOT NULL DEFAULT 1,
			attrs_json TEXT NOT NULL,
			skills_json TEXT,
			hp INTEGER NOT NULL,
			max_hp INTEGER NOT NULL,
			fp INTEGER NOT NULL,
			max_fp INTEGER NOT NULL,
			xp INTEGER NOT NULL DEFAULT 0,
			location_id TEXT,
			pos_x REAL NOT NULL DEFAULT 0,
			pos_y REAL NOT NULL DEFAULT 0,
			ai_json TEXT,
			faction TEXT DEFAULT '',
			flags_json TEXT,
			inventory_json TEXT,
			equipment_json TEXT,
			effects_json TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS entity_quests (
			entity_id TEXT NOT NULL,
			quest_id TEXT NOT NULL,
			state TEXT NOT NULL DEFAULT 'inactive',
			current_stage TEXT DEFAULT '',
			objectives_json TEXT,
			variables_json TEXT,
			activity_json TEXT DEFAULT '[]',
			accepted_tick INTEGER DEFAULT 0,
			PRIMARY KEY (entity_id, quest_id)
		)`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("exec %q: %w", q[:40], err)
		}
	}
	// Migration: add new columns to entities table if they don't exist
	alterQueries := []string{
		"ALTER TABLE entities ADD COLUMN age INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE entities ADD COLUMN max_age INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE entities ADD COLUMN last_meal INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE entity_quests ADD COLUMN completed_stages_json TEXT DEFAULT '[]'",
		"ALTER TABLE entity_quests ADD COLUMN activity_json TEXT DEFAULT '[]'",
		"ALTER TABLE locations ADD COLUMN controlling_faction TEXT DEFAULT ''",
		"ALTER TABLE locations ADD COLUMN control_strength INTEGER DEFAULT 0",
		"ALTER TABLE locations ADD COLUMN exits_json TEXT DEFAULT '[]'",
		"ALTER TABLE entities ADD COLUMN language_skills_json TEXT DEFAULT '{}'",
	}
	for _, q := range alterQueries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("alter %q: %w", q[:40], err)
		}
	}
	log.Printf("Database migrated at %s", s.path)
	return nil
}

func (s *SQLiteStore) Enabled() bool {
	return s.enabled
}

func (s *SQLiteStore) Save(sim *engine.Simulation) error {
	if !s.enabled || s.db == nil {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.Exec(`INSERT OR REPLACE INTO world_state (id, tick, day, hour, minute, speed) VALUES (1, ?, ?, ?, ?, ?)`,
		sim.Tick, sim.Time.Day, sim.Time.Hour, sim.Time.Minute, sim.Time.Speed, "")
	if err != nil {
		return err
	}

	for _, loc := range sim.World.AllLocations() {
		weatherJSON := "null"
		if loc.Weather != nil {
			b, _ := json.Marshal(loc.Weather)
			weatherJSON = string(b)
		}
		tagsJSON := "[]"
		if len(loc.Tags) > 0 {
			b, _ := json.Marshal(loc.Tags)
			tagsJSON = string(b)
		}
		exitsJSON := "[]"
		if len(loc.Exits) > 0 {
			b, _ := json.Marshal(loc.Exits)
			exitsJSON = string(b)
		}
		_, err = tx.Exec(`INSERT OR REPLACE INTO locations (id, name, type, parent_id, pos_x, pos_y, area, is_outside, weather_json, tags_json, controlling_faction, control_strength, exits_json) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			loc.ID, loc.Name, int(loc.Type), nullString(loc.ParentID), loc.Position.X, loc.Position.Y, loc.Area, boolInt(loc.IsOutside), weatherJSON, tagsJSON, loc.ControllingFaction, loc.ControlStrength, exitsJSON)
		if err != nil {
			return err
		}
	}

	for _, ent := range sim.Entities.All() {
		attrsJSON, _ := json.Marshal(ent.Attributes)
		skillsJSON, _ := json.Marshal(ent.Skills)
		aiJSON, _ := json.Marshal(ent.AI)
		flagsJSON, _ := json.Marshal(ent.Flags)
		inventoryJSON, _ := json.Marshal(ent.Inventory)
		equipmentJSON, _ := json.Marshal(ent.Equipment)
		effectsJSON, _ := json.Marshal(ent.Effects)
		langSkillsJSON, _ := json.Marshal(ent.LanguageSkills)

		_, err = tx.Exec(`INSERT OR REPLACE INTO entities (id, name, species, level, age, max_age, last_meal, alive, attrs_json, skills_json, hp, max_hp, fp, max_fp, xp, location_id, pos_x, pos_y, ai_json, faction, flags_json, inventory_json, equipment_json, effects_json, language_skills_json) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			ent.ID, ent.Name, ent.Species, ent.Level, ent.Age, ent.MaxAge, ent.LastMealTick, boolInt(ent.Alive),
			string(attrsJSON), string(skillsJSON),
			ent.HP, ent.MaxHP, ent.FP, ent.MaxFP, ent.XP,
			nullString(ent.LocationID), ent.Position.X, ent.Position.Y,
			string(aiJSON), ent.Faction,
			string(flagsJSON), string(inventoryJSON), string(equipmentJSON),
			string(effectsJSON), string(langSkillsJSON))
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`DELETE FROM entity_quests`)
	if err != nil {
		return err
	}

	for _, ent := range sim.Entities.All() {
		states := sim.Quests.EntityStates(ent.ID)
		for _, state := range states {
			objectivesJSON, _ := json.Marshal(state.Objectives)
			variablesJSON, _ := json.Marshal(state.Variables)
			completedStagesJSON, _ := json.Marshal(state.CompletedStages)
			activityJSON, _ := json.Marshal(state.Activity)
			_, err = tx.Exec(`INSERT OR REPLACE INTO entity_quests (entity_id, quest_id, state, current_stage, completed_stages_json, objectives_json, variables_json, activity_json, accepted_tick) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				ent.ID, state.QuestID, string(state.State), state.CurrentStage,
				string(completedStagesJSON), string(objectivesJSON), string(variablesJSON), string(activityJSON), state.AcceptedTick)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) Load() (*engine.Simulation, error) {
	if !s.enabled || s.db == nil {
		return nil, fmt.Errorf("storage not enabled")
	}

	var tick uint64
	var day, hour, minute, speed int
	var factionRelationsJSON sql.NullString
	err := s.db.QueryRow(`SELECT tick, day, hour, minute, speed, faction_relations_json FROM world_state WHERE id=1`).
		Scan(&tick, &day, &hour, &minute, &speed, &factionRelationsJSON)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no saved world state found")
	}
	if err != nil {
		return nil, fmt.Errorf("read world_state: %w", err)
	}

	if factionRelationsJSON.Valid {
		_ = factionRelationsJSON.String
	}

	w := world.NewWorld()
	gt := world.NewGameTime(speed)
	gt.Tick = tick
	gt.Day = day
	gt.Hour = hour
	gt.Minute = minute

	rows, err := s.db.Query(`SELECT id, name, type, parent_id, pos_x, pos_y, area, is_outside, weather_json, tags_json,
		COALESCE(controlling_faction,''), COALESCE(control_strength,0), COALESCE(exits_json,'[]') FROM locations`)
	if err != nil {
		return nil, fmt.Errorf("read locations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		var locType int
		var parentID sql.NullString
		var posX, posY, area float64
		var isOutside int
		var weatherJSON, tagsJSON sql.NullString
		var ctrlFaction string
		var ctrlStrength int
		var exitsJSON string

		if err := rows.Scan(&id, &name, &locType, &parentID, &posX, &posY, &area, &isOutside, &weatherJSON, &tagsJSON, &ctrlFaction, &ctrlStrength, &exitsJSON); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}

		pID := ""
		if parentID.Valid {
			pID = parentID.String
		}

		loc := world.NewLocation(id, name, world.LocationType(locType), pID, world.Position{X: posX, Y: posY})
		loc.IsOutside = isOutside == 1
		loc.Area = area
		loc.ControllingFaction = ctrlFaction
		loc.ControlStrength = ctrlStrength

		if weatherJSON.Valid && weatherJSON.String != "null" {
			var wth world.Weather
			if err := json.Unmarshal([]byte(weatherJSON.String), &wth); err == nil {
				loc.Weather = &wth
			}
		}
		if tagsJSON.Valid && tagsJSON.String != "[]" {
			var tags []string
			if err := json.Unmarshal([]byte(tagsJSON.String), &tags); err == nil {
				loc.Tags = tags
			}
		}
		if exitsJSON != "" && exitsJSON != "[]" {
			var exits []world.Exit
			if err := json.Unmarshal([]byte(exitsJSON), &exits); err == nil {
				loc.Exits = exits
			}
		}

		w.AddLocation(loc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read locations: %w", err)
	}

	sim := engine.NewSimulation(w, entity.NewEntityManager())
	sim.Tick = tick
	sim.Time = gt

	entRows, err := s.db.Query(`SELECT id, name, species, level, age, max_age, last_meal, alive, attrs_json, skills_json, hp, max_hp, fp, max_fp, xp, location_id, pos_x, pos_y, ai_json, faction, flags_json, inventory_json, equipment_json, effects_json, COALESCE(language_skills_json,'{}') FROM entities`)
	if err != nil {
		return nil, fmt.Errorf("read entities: %w", err)
	}
	defer entRows.Close()

	for entRows.Next() {
		var id, name, species string
		var level, age, maxAge, lastMeal, alive, hp, maxHp, fp, maxFp, xp int
		var attrsJSON, skillsJSON, aiJSON, faction, flagsJSON, inventoryJSON, equipmentJSON, langSkillsJSON string
		var effectsJSON sql.NullString
		var locID sql.NullString
		var posX, posY float64

		if err := entRows.Scan(&id, &name, &species, &level, &age, &maxAge, &lastMeal, &alive, &attrsJSON, &skillsJSON,
			&hp, &maxHp, &fp, &maxFp, &xp, &locID, &posX, &posY,
			&aiJSON, &faction, &flagsJSON, &inventoryJSON, &equipmentJSON, &effectsJSON, &langSkillsJSON); err != nil {
			return nil, fmt.Errorf("scan entity: %w", err)
		}

		var attrs entity.Attributes
		if err := json.Unmarshal([]byte(attrsJSON), &attrs); err != nil {
			return nil, fmt.Errorf("unmarshal attrs for %s: %w", id, err)
		}

		ent := entity.NewEntity(id, name, species, attrs, level, relation.Relation{})
		ent.Alive = alive == 1
		ent.Age = age
		ent.MaxAge = maxAge
		ent.LastMealTick = lastMeal
		ent.HP = hp
		ent.MaxHP = maxHp
		ent.FP = fp
		ent.MaxFP = maxFp
		ent.XP = xp
		if locID.Valid {
			ent.LocationID = locID.String
		}
		ent.Position = entity.Position{X: posX, Y: posY}

		var ai entity.EntityAI
		if err := json.Unmarshal([]byte(aiJSON), &ai); err != nil {
			return nil, fmt.Errorf("unmarshal ai for %s: %w", id, err)
		}
		ent.AI = ai

		ent.Faction = faction
		ent.Skills = make(map[string]int)
		if err := json.Unmarshal([]byte(skillsJSON), &ent.Skills); err != nil {
			return nil, fmt.Errorf("unmarshal skills for %s: %w", id, err)
		}

		ent.Flags = make(map[string]any)
		if flagsJSON != "null" {
			if err := json.Unmarshal([]byte(flagsJSON), &ent.Flags); err != nil {
				return nil, fmt.Errorf("unmarshal flags for %s: %w", id, err)
			}
		}

		ent.Inventory = make([]items.ItemInstance, 0)
		if inventoryJSON != "null" {
			if err := json.Unmarshal([]byte(inventoryJSON), &ent.Inventory); err != nil {
				return nil, fmt.Errorf("unmarshal inventory for %s: %w", id, err)
			}
		}

		if equipmentJSON != "null" {
			var eq entity.Equipment
			if err := json.Unmarshal([]byte(equipmentJSON), &eq); err != nil {
				return nil, fmt.Errorf("unmarshal equipment for %s: %w", id, err)
			}
			ent.Equipment = eq
		}

		if effectsJSON.Valid && effectsJSON.String != "null" && effectsJSON.String != "" {
			var effects []entity.ActiveEffect
			if err := json.Unmarshal([]byte(effectsJSON.String), &effects); err == nil {
				ent.Effects = effects
			}
		}

		ent.LanguageSkills = make(map[string]int)
		if langSkillsJSON != "{}" && langSkillsJSON != "" {
			if err := json.Unmarshal([]byte(langSkillsJSON), &ent.LanguageSkills); err != nil {
				return nil, fmt.Errorf("unmarshal language_skills for %s: %w", id, err)
			}
		}

		sim.Entities.Add(ent)
	}
	if err := entRows.Err(); err != nil {
		return nil, fmt.Errorf("read entities: %w", err)
	}

	questRows, err := s.db.Query(`SELECT entity_id, quest_id, state, current_stage, completed_stages_json, objectives_json, variables_json, activity_json, accepted_tick FROM entity_quests`)
	if err != nil {
		log.Printf("Warning: could not read entity_quests: %v", err)
	} else {
		defer questRows.Close()
		for questRows.Next() {
			var entityID, questID, stateStr, currentStage string
			var completedStagesJSON, objectivesJSON, variablesJSON, activityJSON sql.NullString
			var acceptedTick uint64

			if err := questRows.Scan(&entityID, &questID, &stateStr, &currentStage, &completedStagesJSON, &objectivesJSON, &variablesJSON, &activityJSON, &acceptedTick); err != nil {
				log.Printf("Warning: could not scan quest row: %v", err)
				continue
			}

			var completedStages []string
			if completedStagesJSON.Valid && completedStagesJSON.String != "" {
				if err := json.Unmarshal([]byte(completedStagesJSON.String), &completedStages); err != nil {
					log.Printf("Warning: could not unmarshal completed stages for %s/%s: %v", entityID, questID, err)
				}
			}

			objectives := make(map[string]int)
			if objectivesJSON.Valid && objectivesJSON.String != "" {
				if err := json.Unmarshal([]byte(objectivesJSON.String), &objectives); err != nil {
					log.Printf("Warning: could not unmarshal objectives for %s/%s: %v", entityID, questID, err)
				}
			}

			variables := make(map[string]any)
			if variablesJSON.Valid && variablesJSON.String != "" {
				if err := json.Unmarshal([]byte(variablesJSON.String), &variables); err != nil {
					log.Printf("Warning: could not unmarshal variables for %s/%s: %v", entityID, questID, err)
				}
			}

			var activity []quest.QuestActivity
			if activityJSON.Valid && activityJSON.String != "" {
				if err := json.Unmarshal([]byte(activityJSON.String), &activity); err != nil {
					log.Printf("Warning: could not unmarshal activity for %s/%s: %v", entityID, questID, err)
				}
			}

			sim.Quests.LoadState(entityID, questID, quest.State(stateStr), currentStage, completedStages, objectives, variables, acceptedTick)
			if state := sim.Quests.GetState(entityID, questID); state != nil {
				state.Activity = activity
			}
		}
		if err := questRows.Err(); err != nil {
			return nil, fmt.Errorf("read quest states: %w", err)
		}
	}

	log.Printf("Loaded %d entities from %s", len(sim.Entities.All()), s.path)
	return sim, nil
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
