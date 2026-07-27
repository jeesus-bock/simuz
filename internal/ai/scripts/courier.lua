-- Courier AI
-- A civilian messenger who roams town buildings, flees danger, and cooperates
-- with rescue/leash behavior.

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

local function is_leashed()
    local leashed, dragger = world.is_leashed()
    return leashed, dragger
end

local function maybe_wander_town()
    local parent = world.parent_location(self.loc_id)
    if not parent then
        return
    end
    local buildings = world.exits_from(parent)
    if not buildings or #buildings == 0 then
        return
    end
    local dest = buildings[util.rand_int(#buildings) + 1]
    if dest ~= self.loc_id then
        world.move_to(dest)
    end
end

local function flee_home()
    if self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
    end
end

local function do_tick()
    if world.defend_self and world.defend_self() then
        return
    end

    local leashed = is_leashed()
    if leashed then
        util.set_mood("fearful")
        return
    end

    local hostile_id = find_hostile()
    if hostile_id then
        util.set_mood("stressed")
        flee_home()
        return
    end

    if world.phase == "night" or world.phase == "dusk" then
        flee_home()
        return
    end

    if world.tick % 20 == 0 then
        maybe_wander_town()
    end

    if world.tick % 60 == 0 then
        util.set_mood("neutral")
    end
end

do_tick()
