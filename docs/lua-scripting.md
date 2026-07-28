# Lua Scripting Guide

This document describes the Lua scripting system used for entity AI behavior in the simulation engine.

## Overview

The AI system allows entities to run Lua scripts that control their behavior each tick. Scripts are loaded at runtime and can interact with the entity, world, and other systems through a set of exposed bindings.

## Script Loading

Scripts are stored as compiled bytecode in the scripts directory. The engine loads them into a global script manager. Each script is identified by a name that matches the file name without the extension. The manager provides a simple API for loading and retrieving scripts.

### Example script structure

The following is a minimal example of a Lua script that logs a message each tick:

    local function do_tick()
        util.log(self.name .. " is thinking")
    end

    do_tick()

The script is executed once per tick for the entity it belongs to. The engine passes the entity reference, world state, and other helpers into the script's global namespace.

## Available Bindings

The engine exposes several tables to the script: self, world, and util. These tables contain functions and fields that allow the script to query and modify the simulation.

### self

The self table contains information about the entity running the script. It includes fields such as id, name, species, faction, profession, hp, max_hp, loc_id, and home. It also provides methods for relationship management, inventory inspection, and status effects.

Key methods on self:

- get_relationship(other_id) – returns the relationship data between this entity and another.
- add_relationship(other_id, type, since_tick) – establishes a new relationship.
- remove_relationship(other_id) – removes an existing relationship.
- has_relationship(other_id) – checks whether a relationship exists.
- set_faction(new_faction) – changes the entity's faction.
- set_profession(new_profession) – changes the entity's profession.

### world

The world table provides access to the simulation world, locations, entities, weather, and combat functions. It includes functions such as:

- nearby_entities() – returns a list of entity IDs in the same location.
- entity_info(entity_id) – returns a table with basic info about an entity.
- move_to(location_id) – attempts to move the entity to a new location.
- exits_from(location_id) – returns a list of adjacent location IDs.
- weather(location_id) – returns current weather data for a location.
- attack(attacker_id, target_id) – performs a combat attack.
- damage_location(attacker_id, amount) – applies area damage to nearby entities.
- heal(entity_id, amount) – restores hit points.
- add_item(entity_id, def_id, count) – gives an item to an entity.
- try_buy(seller_id, item_def_id) – attempts to purchase an item from a seller.
- try_sell(buyer_id, item_def_id) – attempts to sell an item to a buyer.
- use_item(item_def_id) – consumes a usable item.
- talk_to(npc_id) – initiates dialogue with an NPC.
- deliver_item(npc_id, item_def_id) – attempts to hand an item to an NPC.
- feed(target_id) – feeds another entity.
- give_quest(target_id, quest_id) – assigns a quest to another entity.
- quest_progress(quest_id, objective_id, delta) – advances a quest objective.
- quest_set(quest_id, objective_id, value) – sets a quest objective directly.
- craft(recipe_id) – attempts to craft an item using available materials.
- has_material(def_id) – checks whether the entity possesses a specific item definition.
- recipe_info(recipe_id) – returns details about a recipe, including required inputs and outputs.
- drag_entity(target_id) – starts dragging or leashing another entity.
- undrag_entity(target_id) – releases a dragged entity.
- is_leashed(target_id) – checks whether an entity is being dragged.
- start_rescue(target_id) – begins a rescue operation.
- complete_rescue(target_id) – finishes a rescue.

### util

The util table contains helper functions for randomization, logging, memory storage, and JSON serialization.

Key functions:

- rand_int(max) – returns a random integer between 0 and max-1.
- log(message) – prints a message to the simulation log.
- mem_set(key, value) – stores a value in the entity's persistent memory.
- mem_get(key) – retrieves a value from the entity's memory.
- set_mood(mood, duration) – changes the entity's mood for a given duration.
- json_encode(value) – converts a Lua table or value to a JSON string.
- json_decode(json_string) – parses a JSON string into a Lua value.

## Event System

Scripts can emit simulation events by returning a table of events. Each event is a table with the following fields:

- type – an integer representing the event type.
- tick – the simulation tick when the event occurred.
- source – the entity ID that generated the event.
- data – a table of additional data relevant to the event.

The engine collects these events and makes them available to other systems such as quest managers, AI, and UI.

## Example: Simple Guard AI

The following script implements a basic guard that patrols between two locations and attacks any hostile that enters its range.

    local function find_hostile()
        local nearby = world.nearby_entities()
        if not nearby then return nil end
        for _, eid in ipairs(nearby) do
            local info = world.entity_info(eid)
            if info and info.alive and world.is_hostile(self.faction, info.faction) then
                return eid, info
            end
        end
        return nil
    end

    local function patrol()
        local exits = world.exits_from(self.loc_id)
        if not exits or #exits == 0 then return end
        local dest = exits[util.rand_int(#exits) + 1]
        world.move_to(dest)
    end

    local function do_tick()
        local hostile_id, hostile_info = find_hostile()
        if hostile_id then
            world.attack(self.id, hostile_id)
            return
        end
        patrol()
    end

    do_tick()

## Best Practices

- Keep scripts focused on a single behavior. Complex logic can be split into named local functions.
- Use util.log for debugging; avoid spamming the log in tight loops.
- Store persistent state in util.mem_set rather than global variables.
- Respect the entity's mood and avoid aggressive actions when the mood is fearful.
- Check for nil values before using data returned by world functions.
- Use util.rand_int for any probabilistic decisions to ensure consistent randomness across runs.

## Further Reading

- The ai/scripts directory contains many example scripts for different species and professions.
- The internal/ai/scripts folder holds the core AI logic for factions, combat, and interactions.
- The internal/api package provides HTTP endpoints for external tools to query entity and world state.
