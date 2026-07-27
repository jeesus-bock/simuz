-- Scouting AI
-- Explores territory, reports findings, flees when threatened.
-- Avoids direct combat at all costs.

local EXPLORE_INTERVAL = 8
local FLEE_THRESHOLD = 0.4
local RETURN_HOME_INTERVAL = 50

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

local function do_explore()
    local exits = world.exits_from(self.loc_id)
    if exits and #exits > 0 then
        local dest = exits[util.rand_int(#exits) + 1]
        if dest ~= self.loc_id then
            world.move_to(dest)
            util.log(self.name .. " scouted " .. dest)
        end
    end
end

local function do_tick()
    local tick = world.tick
    local phase = world.phase

    if phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
        end
        return
    end

    -- Flee from hostiles immediately
    local hostile_id, hostile_info = find_hostile()
    if hostile_id then
        if self.hp < self.max_hp * FLEE_THRESHOLD then
            if self.home and self.loc_id ~= self.home then
                world.move_to(self.home)
                util.log(self.name .. " fled home from threat")
            end
            return
        end
        -- Avoid combat, just stay alert
        util.log(self.name .. " spotted hostile, avoiding contact")
        return
    end

    -- Explore periodically
    if tick % EXPLORE_INTERVAL == 0 then
        do_explore()
    end

    -- Return home periodically
    if tick % RETURN_HOME_INTERVAL == 0 and self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
    end
end

do_tick()