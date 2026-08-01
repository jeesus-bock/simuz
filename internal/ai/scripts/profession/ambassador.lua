-- ambassador.lua
-- Diplomatic sub-category of diplomat: elite negotiators, treaty-makers,
-- stronghold defenders, and escorts. Ambassadors have diplomatic immunity
-- and coordinate defense of their species' strongholds.
--
-- Sub-category of: diplomat
-- Ranks above: diplomat
-- Species with this rank: elf (ambassador), dwarf (ambassador)

local function find_diplomat_targets()
    local nearby = world.nearby_entities()
    if not nearby then return {} end
    local targets = {}
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                -- Prioritize foreign diplomats and politicians for treaty negotiations
                if info.profession == "diplomat" or info.profession == "politician" then
                    table.insert(targets, {id = eid, info = info, priority = 1})
                -- Then merchants and couriers for trade agreements
                elseif info.profession == "merchant" or info.profession == "courier" then
                    table.insert(targets, {id = eid, info = info, priority = 2})
                end
            end
        end
    end
    -- Sort by priority (diplomats first)
    table.sort(targets, function(a, b) return a.priority < b.priority end)
    return targets
end

local function negotiate_treaty(target_id, target_info)
    if world.can_communicate(target_id) then
        world.dialog(self.id, target_id, "treaty")
        util.log(self.name .. " opens treaty negotiations with " .. target_info.name)
        util.set_mood("diplomatic", 30)
        return true
    else
        util.log(self.name .. " presents formal credentials to " .. target_info.name)
        util.set_mood("formal", 20)
        return true
    end
end

local function escort_ally()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.faction == self.faction then
                -- Escort wounded or low-level allies to safety
                if info.hp < info.max_hp * 0.5 or info.level < 5 then
                    if self.loc_id ~= self.home then
                        world.move_to(self.home)
                        util.log(self.name .. " escorts " .. info.name .. " to the stronghold")
                        return true
                    end
                end
            end
        end
    end
    return false
end

local function coordinate_stronghold_defense()
    -- Elven and dwarven ambassadors boost defensive stats of nearby allies
    local nearby = world.nearby_entities()
    if not nearby then return false end

    local allies_nearby = 0
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.faction == self.faction then
                allies_nearby = allies_nearby + 1
            end
        end
    end

    -- If we have allies nearby, provide a defensive buff (simulated via logging)
    if allies_nearby >= 2 then
        util.log(self.name .. " rallies the stronghold defenders! +" .. (allies_nearby * 5) .. "% defensive coordination")
        util.set_mood("authoritative", 20)
        return true
    end
    return false
end

local function enforce_diplomatic_immunity()
    -- Ambassadors cannot be attacked by non-hostile factions
    -- If a non-hostile entity tries to attack, the ambassador is protected
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and not world.is_hostile(self.faction, info.faction) then
                -- Non-hostile entity is nearby — diplomatic immunity applies
                -- Log the protection but don't take action unless attacked
                if info.profession == "thief" or info.profession == "bandit_chief" then
                    util.log(self.name .. " invokes diplomatic immunity against " .. info.name)
                    util.set_mood("authoritative", 15)
                end
            end
        end
    end
    return false
end

local function find_hostile_threat()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and world.is_hostile(self.faction, info.faction) then
                return eid, info
            end
        end
    end
    return nil
end

function do_tick()
    local tick = world.tick
    local events = {}

    -- Priority 1: Enforce diplomatic immunity
    enforce_diplomatic_immunity()

    -- Priority 2: Coordinate stronghold defense
    if tick % 15 == 0 then
        if coordinate_stronghold_defense() then
            table.insert(events, util.event("profession_action", {profession = "ambassador", rank = "ambassador"}))
        end
    end

    -- Priority 3: Escort wounded allies
    if tick % 20 == 0 then
        if escort_ally() then
            table.insert(events, util.event("move", {}))
        end
    end

    -- Priority 4: Negotiate with diplomats and politicians
    if tick % 25 == 0 then
        local targets = find_diplomat_targets()
        if #targets > 0 then
            local target = targets[1]
            if negotiate_treaty(target.id, target.info) then
                table.insert(events, util.event("profession_action", {profession = "ambassador", action = "negotiate"}))
            end
        end
    end

    -- Priority 5: Defend against hostile threats
    local hostile_id, hostile_info = find_hostile_threat()
    if hostile_id then
        -- Ambassadors are defenders, not attackers — coordinate defense instead of engaging directly
        util.log(self.name .. " coordinates defense against " .. (hostile_info.name or hostile_id))
        util.set_mood("authoritative", 25)
        -- Rally nearby allies to defend
        coordinate_stronghold_defense()
        table.insert(events, util.event("defend", {target = hostile_id}))
    end

    -- Priority 6: Return to stronghold at night
    local phase = world.phase
    if phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            table.insert(events, util.event("move", {}))
        end
        return events
    end

    if #events > 0 then
        return events
    end
    return {}
end

return do_tick()
