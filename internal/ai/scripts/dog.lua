-- Dog AI
-- A farm or town companion that stays near home, follows leashes cleanly,
-- and otherwise wanders a little during the day.

local function is_leashed()
    local leashed, dragger = world.is_leashed()
    return leashed, dragger
end

local function find_home_parent()
    if not self.home then return nil end
    return world.parent_location(self.home) or self.home
end

local function wander_near_home()
    local parent = world.parent_location(self.loc_id)
    if not parent then
        parent = find_home_parent()
    end
    if not parent then return end
    local exits = world.exits_from(parent)
    if exits and #exits > 0 then
        local dest = exits[util.rand_int(#exits) + 1]
        if dest ~= self.loc_id then
            world.move_to(dest)
        end
    end
end

local function do_tick()
    local leashed = is_leashed()
    if leashed then
        util.set_mood("happy")
        return
    end

    if world.phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
        end
        return
    end

    if world.tick % 30 == 0 and util.rand_int(100) < 50 then
        wander_near_home()
    end

    if world.tick % 45 == 0 then
        util.set_mood("relaxed")
    end
end

do_tick()
