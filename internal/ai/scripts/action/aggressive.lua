-- Aggressive AI
-- Attacks hostiles in same location on sight.
-- Roams between adjacent locations when no target.
-- Rate-limited per-faction.
--
-- Intended for: Aggressive monsters, wild beasts, and any entity that
-- should engage hostiles proactively without needing a specific profession
-- or faction trigger.

-- faction_hash computes a deterministic numeric hash from the entity's
-- faction string. This is used to vary the attack rate per faction so
-- that different factions attack at different cadences.
-- Input: none (uses self.faction)
-- Output: integer hash value
local function faction_hash()
    local h = 0
    for i = 1, #self.faction do
        h = h + string.byte(self.faction, i)
    end
    return h
end

-- get_attack_rate returns the tick interval between attacks for this
-- entity, derived from the faction hash. Factions with different hash
-- values will attack at slightly different rates (5-9 ticks apart).
-- Input: none
-- Output: integer tick rate (5 + (faction_hash % 5))
local function get_attack_rate()
    return 5 + (faction_hash() % 5)
end

-- find_hostile_target scans nearby entities for the first hostile
-- target that is conscious (alive and not knocked out). If no conscious
-- hostile exists, it falls back to the first downed (knocked-out) hostile.
-- This ensures aggressive entities finish off weakened foes when no
-- active threat remains.
-- Input: none (uses world.nearby_entities and world.entity_info)
-- Output: (target_id string, target_info *EntityInfo) or (nil, nil)
local function find_hostile_target()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return nil
    end
    -- Prefer conscious hostiles; finish knocked-out foes only when no active threat remains.
    local downed_id, downed_info = nil, nil
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.species ~= "deity" and world.is_hostile(self.faction, info.faction) then
            if info.conscious ~= false then
                return eid, info
            end
            if not downed_id then
                downed_id, downed_info = eid, info
            end
        end
    end
    return downed_id, downed_info
end

-- wander picks a random adjacent location to move to when there is no
-- hostile target in range. It prefers real travel exits (camps, dens,
-- sibling buildings) over falling back to children of the parent location.
-- Input: none (uses self.loc_id, world.travel_exits, world.parent_location,
--         world.exits_from, world.move_to)
-- Output: none (side effect: moves entity to a new location)
local function wander()
    -- Prefer real exits (camps, dens, sibling buildings) before falling back to children of parent
    local exits = world.travel_exits(self.loc_id)
    local targets = {}
    if exits then
        for _, e in ipairs(exits) do
            local tid = e.target_id
            if tid and tid ~= self.loc_id then
                table.insert(targets, tid)
            end
        end
    end
    if #targets == 0 then
        local parent_id = world.parent_location(self.loc_id)
        if parent_id and parent_id ~= "" then
            local children = world.exits_from(parent_id)
            if children then
                for _, sid in ipairs(children) do
                    if sid ~= self.loc_id then
                        table.insert(targets, sid)
                    end
                end
            end
        end
    end
    if #targets == 0 then
        return
    end
    local dest = targets[util.rand_int(#targets) + 1]
    local ok = world.move_to(dest)
    if ok then
        util.log(self.name .. " prowled to " .. dest)
    end
end

-- do_tick is the main entry point called every tick by the AI runtime.
-- It rate-limits attacks per faction, finds and attacks hostile targets,
-- or wanders between locations when no target is present.
-- Input: none (uses world.tick, world.attack, world.move_to, etc.)
-- Output: none (side effects: attacks, movement, logging)
local function do_tick()
    local tick = world.tick
    local rate = get_attack_rate()

    -- Rate-limit: skip this tick if it's not our attack interval
    if tick % rate ~= 0 then
        return
    end

    -- Find a hostile target in the same location
    local target, info = find_hostile_target()
    if target then
        -- Attack the target; log if the target is killed
        local ok = world.attack(self.id, target)
        if ok and info and info.hp <= 0 then
            util.log(self.name .. " killed " .. (info.name or target))
        end
        return
    end

    -- No target found: wander to a random adjacent location periodically
    if tick % 30 == 0 and util.rand_int(100) < 40 then
        wander()
    end
end

do_tick()
