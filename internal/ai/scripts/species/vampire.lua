-- vampire.lua
-- High-tier nocturnal predator script tracking blood values and avoiding sunlight.

local function feed_on_living()
    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            -- Target healthy human or elven targets, avoiding fellow undead or werewolves
            if info and info.alive and info.species ~= "undead" and info.faction ~= "werewolves" then
                util.log(self.name .. " sinks fangs into " .. info.name .. "'s neck to drain life fluid!")
                util.set_mood("ecstatic", 40)

                world.attack(self.id, eid)
                world.heal(self.id, 25) -- Regain structural hit points from feeding
                return true
            end
        end
    end
    return false
end

local function do_tick()
    -- 1. Daylight survival emergency check: Fatal vulnerability
    if world.phase == "day" then
        util.log(self.name .. " is burning in the sun! Retreating to deep crypt catacombs!")
        util.set_mood("panicked", 10)

        -- Take structural environment damage or apply penalites
        world.damage_location(self.id, 5)

        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
        end
        return true
    end

    -- 2. Night Hunting Operations Loop
    if world.tick % 8 == 0 then
        local fed = feed_on_living()

        -- If the current location map sector is drained of living targets, migrate down paths
        if not fed and util.rand_int(100) < 40 then
            local exits = world.exits_from(self.loc_id)
            if exits and #exits > 0 then
                world.move_to(exits[util.rand_int(#exits) + 1])
                return true
            end
        end
        return fed
    end
    return false
end

return do_tick()
