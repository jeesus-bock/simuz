-- Hunting AI
-- Actively pursues and attacks prey when in range.
-- Tracks hostile entities across nearby locations.

local CHASE_INTERVAL = 5
local ATTACK_CHANCE = 40

local function find_hostile_prey()
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

local function do_pursuit(eid)
    if not world.attack then return false end

    local chance = util.rand_int(100)
    if chance < ATTACK_CHANCE then
        local hit = world.attack(self.id, eid)
        if hit then
            local info = world.entity_info(eid)
            util.log(self.name .. " pursued and attacked " .. (info and info.name or eid))
        end
        return true
    end
    return false
end

local function do_tick()
    local tick = world.tick
    local phase = world.phase

    if phase == "night" or phase == "dusk" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
        end
        return
    end

    -- Check for prey at current location
    local prey_id, prey_info = find_hostile_prey()
    if prey_id then
        do_pursuit(prey_id)
        return
    end

    -- Hunt across nearby locations
    if tick % CHASE_INTERVAL == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            for _, eid in ipairs(exits) do
                local info = world.entity_info(eid)
                if info and info.alive and world.is_hostile(self.faction, info.faction) then
                    world.move_to(eid)
                    util.log(self.name .. " chased prey to " .. eid)
                    return
                end
            end
        end

        -- Wander toward a random adjacent location
        local nearby = world.nearby_entities()
        if not nearby or #nearby == 0 then
            if self.home then
                local exits = world.exits_from(self.loc_id)
                if exits and #exits > 0 then
                    local dest = exits[util.rand_int(#exits) + 1]
                    if dest ~= self.loc_id then
                        world.move_to(dest)
                    end
                end
            end
        end
    end
end

do_tick()
