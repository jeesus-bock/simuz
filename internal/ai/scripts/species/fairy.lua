-- Fairy AI - Sleep/Charm/Steal
-- Puts enemies to sleep
-- Charms mortals to change faction
-- Steals items on hit
-- Has chance to evade melee

local SLEEP_CHANCE = 30
local CHARM_CHANCE = 20
local STEAL_CHANCE = 20
local EVADE_CHANCE = 40
local SLEEP_DURATION = 5
local CHARM_DURATION = 50
local WANDER_INTERVAL = 15

local function find_hostile()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return nil
    end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and world.is_hostile(self.faction, info.faction) then
            return eid, info
        end
    end
    return nil
end

local function do_sleep(eid)
    if util.rand_int(100) < SLEEP_CHANCE then
        util.mem_set("sleep_target_" .. eid, world.tick + SLEEP_DURATION)
        local info = world.entity_info(eid)
        util.log(self.name .. " put " .. (info and info.name or eid) .. " to sleep for " .. SLEEP_DURATION .. " ticks")
        return true
    end
    return false
end

local function do_charm(eid)
    if util.rand_int(100) < CHARM_CHANCE then
        util.mem_set("charm_target_" .. eid, world.tick + CHARM_DURATION)
        local info = world.entity_info(eid)
        util.log(self.name .. " charmed " .. (info and info.name or eid) .. " for " .. CHARM_DURATION .. " ticks")
        return true
    end
    return false
end

local function do_steal(eid)
    if util.rand_int(100) < STEAL_CHANCE then
        local items = world.entity_items(eid)
        if items and #items > 0 then
            local item = items[util.rand_int(#items) + 1]
            world.add_item(item)
            local info = world.entity_info(eid)
            util.log(self.name .. " stole " .. item .. " from " .. (info and info.name or eid))
            return true
        end
    end
    return false
end

local function try_evade()
    if util.rand_int(100) < EVADE_CHANCE then
        util.log(self.name .. " evaded an attack with flight")
        return true
    end
    return false
end

local function do_tick(self)
    -- 1. Fail-safe: Ensure the actor executing the tick actually exists
    if not self then return false end

    local phase = world.phase

    if phase == "night" or phase == "dusk" then
        -- 2. Added safety check for self.home and self.loc_id
        if self.home and self.loc_id and self.loc_id ~= self.home then
            world.move_to(self.home)
            return true
        end
        return false
    end

    local target_id, target_info = find_hostile()
    -- 3. Fail-safe: Only interact if a target and target_info were successfully found
    if target_id and target_info then
        local attacked = util.rand_int(100) < 60
        if attacked then
            local hit = world.attack(self.id, target_id)
            if hit then
                do_steal(target_id)
            end
        end
        do_sleep(target_id)
        if target_info.species == "human" then
            do_charm(target_id)
        end
        return true
    end

    -- 4. Added safe-navigation check for world.tick and WANDER_INTERVAL
    if world.tick and WANDER_INTERVAL and (world.tick % WANDER_INTERVAL == 0) then
        -- Ensure we have a valid location ID to query exits from
        if self.loc_id then
            local exits = world.exits_from(self.loc_id)
            if exits and #exits > 0 then
                local dest = exits[util.rand_int(#exits) + 1]
                -- 5. Added explicit nil check for the picked destination
                if dest and dest ~= self.loc_id then
                    world.move_to(dest)
                    return true
                end
            end
        end
    end
    return false
end

-- When executing the tick, pass the current entity context
return do_tick(self)
