-- Child AI
-- A harmless town wanderer who plays during the day, runs from any hostiles,
-- and returns home at night. Never attacks.

local function find_hostile()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and world.is_hostile("civilian", info.faction) then
            return eid, info
        end
    end
    return nil
end

local function find_home_parent()
    if not self.home then return nil end
    local parent = world.parent_location(self.home)
    if parent then
        return parent
    end
    return self.home
end

local function flee_home()
    if self.home then
        if self.loc_id ~= self.home then
            world.move_to(self.home)
            util.log(self.name .. " ran home scared")
        end
    else
        local parent = world.parent_location(self.loc_id)
        if parent and parent ~= self.loc_id then
            world.move_to(parent)
        end
    end
end

local function wander_town()
    local parent = world.parent_location(self.loc_id)
    if not parent then
        parent = find_home_parent()
    end
    if not parent then return end
    local buildings = world.exits_from(parent)
    if buildings and #buildings > 0 then
        local dest = buildings[util.rand_int(#buildings) + 1]
        if dest ~= self.loc_id then
            world.move_to(dest)
            util.log(self.name .. " wandered to " .. dest)
        end
    end
end

function do_tick()
    local phase = world.phase
    local tick = world.tick

    if phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            return {util.event("move", {})}
        end
        return {}
    end

    local hostile_id, hostile_info = find_hostile()
    if hostile_id then
        flee_home()
        util.set_mood("fearful")
        return {util.event("flee", {})}
    end

    local acted = false
    if tick % 25 == 0 then
        if util.rand_int(100) < 60 then
            wander_town()
            acted = true
        end
    end

    if tick % 40 == 0 then
        local roll = util.rand_int(100)
        if roll < 40 then
            util.set_mood("happy")
        elseif roll < 70 then
            util.set_mood("neutral")
        else
            util.set_mood("inspired")
        end
    end
    return acted and {util.event("species_action", {})} or {}
end

return do_tick()
