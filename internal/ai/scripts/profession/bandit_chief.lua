-- bandit_chief.lua
-- Advanced tactical script for orchestrating roadside robberies and gang command.
-- Orc bandits may target diplomats, but diplomatic immunity blocks attacks
-- unless a nearby politician has revoked the diplomat's immunity.

local function coordinate_raid()
    local nearby = world.nearby_entities()
    if not nearby then return end

    -- Check health status; if critically wounded, fallback immediately
    if self.hp and self.hp < 40 then
        util.log(self.name .. " is heavily injured! Ordering a tactical retreat!")
        util.set_mood("panicked", 15)
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then world.move_to(exits[util.rand_int(#exits) + 1]) end
        return
    end

    local diplomat_target = nil

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                -- Rob vulnerable classes like merchants, couriers, or basic travelers
                if info.profession == "merchant" or info.profession == "courier" or info.profession == "traveler" then
                    util.log(self.name .. " signals the ambush! 'Cut their coin purses open!'")
                    util.set_mood("furious", 30)

                    -- Area effect assault to overwhelm the targets
                    world.damage_location(self.id, 12)
                    world.attack(self.id, eid)

                    -- Attempt extortion shakedown mechanics
                    local plunder = world.try_buy(eid, "gold_pouch")
                    if plunder and plunder.done then
                        util.log(self.name .. " extracted the caravan spoils from " .. info.name)
                    end
                    return
                end

                -- Track diplomat targets for potential attack
                if info.profession == "diplomat" then
                    diplomat_target = info
                end
            end
        end
    end

    -- Attempt to attack a diplomat if found (may be blocked by diplomatic immunity)
    if diplomat_target and util.rand_int(100) < 30 then
        util.log(self.name .. " lunges at diplomat " .. diplomat_target.name .. "!")
        util.set_mood("aggressive", 20)
        world.attack(self.id, diplomat_target.id)
    end
end

function do_tick()
    local acted = false

    if world.tick % 8 == 0 then
        coordinate_raid()
        acted = true
    end

    -- Periodically move territory layout boundaries to find fresh caravans
    if world.tick % 50 == 0 and util.rand_int(100) < 40 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            world.move_to(exits[util.rand_int(#exits) + 1])
            acted = true
        end
    end

    if acted then
        return {util.event("profession_action", {profession = "bandit_chief"})}
    end
    return {}
end

return do_tick()
