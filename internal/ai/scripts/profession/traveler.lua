-- Traveler AI
-- Wandering civilian that moves between locations casually.
-- Avoids combat, returns home at night.

local WANDER_INTERVAL = 25

function do_tick()
    local phase = world.phase

    if phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            return {util.event("profession_action", {profession = "traveler"})}
        end
        return {}
    end

    if world.tick % WANDER_INTERVAL ~= 0 then
        return {}
    end

    if util.rand_int(100) < 40 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            if dest ~= self.loc_id then
                world.move_to(dest)
            return {util.event("profession_action", {profession = "traveler"})}
            end
        end
    end
    return {}
end

return do_tick()
