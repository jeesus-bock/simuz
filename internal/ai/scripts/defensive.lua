-- Defensive AI
-- Stays near home, attacks only when directly threatened.
-- Flees or retreats when HP is low.

local DODGE_TICKS = 4
local FLEE_HP_THRESHOLD = 0.25

local function do_tick()
    local tick = world.tick
    local phase = world.phase

    if phase == "night" or phase == "dusk" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
        end
        return
    end

    if tick % DODGE_TICKS ~= 0 then
        return
    end

    if self.hp < self.max_hp * FLEE_HP_THRESHOLD then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            util.log(self.name .. " is retreating to safety")
        end
        return
    end

    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return
    end

    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and world.is_hostile(self.faction, info.faction) then
            if world.attack then
                local hit = world.attack(self.id, eid)
                if hit then
                    util.log(self.name .. " defended against " .. (info.name or eid))
                end
            end
            return
        end
    end

    if self.home and self.loc_id ~= self.home and tick % 30 == 0 then
        world.move_to(self.home)
    end
end

do_tick()