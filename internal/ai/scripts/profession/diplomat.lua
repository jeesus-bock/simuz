-- diplomat.lua
-- Machiavellian Statecraft Script: Realpolitik negotiation,
-- calculated alliance shifts, and strategic destabilization.
-- Diplomats carry diplomatic immunity and seek out politicians.
--
-- Sub-categories:
--   diplomat   - Standard diplomatic rank (humans, half-elves, half-dwarves)
--   ambassador - Elite diplomatic rank (elves, dwarves) — stronger negotiation,
--                stronghold defense coordination, and diplomatic immunity enforcement

local function find_politicians()
    local nearby = world.nearby_entities()
    if not nearby then return {} end
    local politicians = {}
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.profession == "politician" then
                table.insert(politicians, info)
            end
        end
    end
    return politicians
end

local function negotiate_with_politician(pol)
    if world.can_communicate(pol.id) then
        world.dialog(self.id, pol.id, "negotiation")
        util.log(self.name .. " engages politician " .. pol.name .. " in formal diplomacy.")
        util.set_mood("diplomatic", 25)
        return true
    else
        util.log(self.name .. " presents written credentials to " .. pol.name .. ".")
        util.set_mood("formal", 15)
        return true
    end
end

local function practice_statecraft()
    local nearby = world.nearby_entities()
    if not nearby then return nil, nil end

    -- Realpolitik: "He who adapts his policy to the times prospers."
    -- If critically wounded, retreat immediately. Preservation of the state (self) is paramount.
    if self.hp and self.hp < 45 then
        util.log(self.name .. " calculates poor odds. 'A prince must know how to escape danger when necessary.'")
        util.set_mood("calculating", 15)
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then world.move_to(exits[util.rand_int(#exits) + 1]) end
        return nil, nil
    end

    local strongest_entity = nil
    local weakest_entity = nil
    local max_hp = -1
    local min_hp = 9999

    -- 1. Scan the room to map the local balance of power
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                -- Skip diplomats and politicians for power assessment
                if info.profession ~= "diplomat" and info.profession ~= "politician" then
                    if info.hp > max_hp then
                        max_hp = info.hp
                        strongest_entity = info
                    end
                    if info.hp < min_hp then
                        min_hp = info.hp
                        weakest_entity = info
                    end
                end
            end
        end
    end

    return strongest_entity, weakest_entity
end

local function negotiate(strongest, weakest)
    if strongest and type(strongest) == "table" and strongest.id ~= self.id then
        -- Flattery: Defer to the strongest entity present
        if world.can_communicate(strongest.id) then
            world.say_to(strongest.id, "Your strength precedes you. Perhaps we might find mutual benefit.")
        end
        util.log(self.name .. " addresses " .. strongest.name .. " with calculated deference.")
        util.set_mood("diplomatic", 20)
        return true
    end

    if weakest and type(weakest) == "table" and weakest.id ~= self.id and weakest.faction ~= self.faction then
        -- Destabilization: Seed distrust in the weakest non-aligned entity
        if world.can_communicate(weakest.id) then
            world.say_to(weakest.id, "I hear troubling whispers about your rivals...")
        end
        util.log(self.name .. " whispers veiled threats to " .. weakest.name .. ".")
        util.set_mood("scheming", 20)
        return true
    end

    return false
end

-- -----------------------------------------------------------------------
-- Ambassador sub-category: elite diplomatic behavior
-- -----------------------------------------------------------------------
local function is_ambassador()
    return self.diplomatic_rank == "ambassador"
end

local function find_ambassador_targets()
    local nearby = world.nearby_entities()
    if not nearby then return {} end
    local targets = {}
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                -- Ambassadors prioritize treaty negotiations and stronghold defense
                if info.profession == "diplomat" or info.profession == "politician" then
                    table.insert(targets, {id = eid, info = info, priority = 1})
                elseif info.profession == "merchant" or info.profession == "courier" then
                    table.insert(targets, {id = eid, info = info, priority = 2})
                end
            end
        end
    end
    table.sort(targets, function(a, b) return a.priority < b.priority end)
    return targets
end

local function coordinate_stronghold_defense()
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

    if allies_nearby >= 2 then
        util.log(self.name .. " rallies the stronghold defenders! +" .. (allies_nearby * 5) .. "% defensive coordination")
        util.set_mood("authoritative", 20)
        return true
    end
    return false
end

local function enforce_diplomatic_immunity()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and not world.is_hostile(self.faction, info.faction) then
                if info.profession == "thief" or info.profession == "bandit_chief" then
                    util.log(self.name .. " invokes diplomatic immunity against " .. info.name)
                    util.set_mood("authoritative", 15)
                end
            end
        end
    end
    return false
end

local function escort_ally()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.faction == self.faction then
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

function do_tick()
    local acted = false

    -- Ambassadors get additional elite behaviors
    if is_ambassador() then
        -- Priority 1: Enforce diplomatic immunity
        if world.tick % 5 == 0 then
            enforce_diplomatic_immunity()
        end

        -- Priority 2: Coordinate stronghold defense
        if world.tick % 15 == 0 then
            if coordinate_stronghold_defense() then
                acted = true
            end
        end

        -- Priority 3: Escort wounded allies
        if world.tick % 20 == 0 then
            if escort_ally() then
                acted = true
            end
        end
    end

    -- Standard diplomat behaviors
    if world.tick % 20 == 0 then
        -- Priority: Seek out politicians for formal diplomacy
        local politicians = find_politicians()
        if #politicians > 0 then
            acted = negotiate_with_politician(politicians[1])
        end

        -- Priority: General statecraft with non-politician entities
        if not acted then
            local strongest, weakest = practice_statecraft()
            if strongest ~= nil and type(strongest) == "table" then
                acted = negotiate(strongest, weakest)
            end
        end

        if not acted then
            -- If no targets, travel to a new location to scout
            local exits = world.exits_from(self.loc_id)
            if exits and #exits > 0 then
                local dest = exits[util.rand_int(#exits) + 1]
                if dest ~= self.loc_id then
                    util.log(self.name .. " leaves to sow regional stability—or profitable discord.")
                    world.move_to(dest)
                    acted = true
                end
            end
        end
    end

    if acted then
        return {util.event("profession_action", {profession = "diplomat", rank = self.diplomatic_rank or "diplomat"})}
    end
    return {}
end

return do_tick()
