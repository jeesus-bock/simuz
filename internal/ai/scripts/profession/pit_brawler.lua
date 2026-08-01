-- pit_brawler.lua
-- Constantly challenges nearby entities to physical combat within closed rooms.

local function provoke_fight()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                util.log(self.name .. " headbutts " .. info.name .. "! 'You look soft! Let's see what you're made of!'")
                util.set_mood("furious", 40)

                -- Immediate localized bar-fight challenge
                world.attack(self.id, eid)
                return true
            end
        end
    end
    return false
end

function do_tick()
    local acted = false
    -- Highly aggressive, checks for fights constantly
    if world.tick % 4 == 0 and util.rand_int(100) < 70 then acted = provoke_fight() end
    if world.tick % 40 == 0 and util.rand_int(100) < 30 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            world.move_to(exits[util.rand_int(#exits) + 1])
            acted = true
        end
    end
    if acted then return { util.event("profession_action", { profession = "pit_brawler" }) } end
    return {}
end

return do_tick()
