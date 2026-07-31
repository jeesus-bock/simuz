-- Gathering AI
-- Collects resources and avoids combat.
-- Prefers outdoor resource-rich locations.

local GATHER_CHANCE = 20
local FLEE_HP_THRESHOLD = 0.3

local function is_hostile_nearby()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return false
    end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.species ~= "deity" then
            if world.is_hostile(self.faction, info.faction) then
                return true
            end
        end
    end
    return false
end

local function do_gather()
    local nearby = world.nearby_entities()
    if nearby and #nearby > 0 then
        for _, eid in ipairs(nearby) do
            local info = world.entity_info(eid)
            if info and info.alive and world.is_hostile(self.faction, info.faction) then
                if self.home and self.loc_id ~= self.home then
                    world.move_to(self.home)
                    util.log(self.name .. " fled from hostile")
                end
                return
            end
        end
    end

    local roll = util.rand_int(100)
    if roll < GATHER_CHANCE then
        util.log(self.name .. " gathered resources")
    end
end

local function do_tick()
    local tick = world.tick
    local phase = world.phase
    local acted = false

    if phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            return true
        end
        return false
    end

    if is_hostile_nearby() then
        if self.hp < self.max_hp * FLEE_HP_THRESHOLD then
            if self.home and self.loc_id ~= self.home then
                world.move_to(self.home)
                return true
            end
        end
        return false
    end

    if tick % 20 == 0 then
        do_gather()
        acted = true
    end

    if self.home and self.loc_id ~= self.home and tick % 30 == 0 then
        world.move_to(self.home)
        acted = true
    end
    return acted
end

return do_tick()
