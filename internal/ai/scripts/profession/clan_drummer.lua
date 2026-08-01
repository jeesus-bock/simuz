-- clan_drummer.lua
-- Beats tribal war drums to alter room psychological statuses drastically.

local function sound_drums()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    util.log(self.name .. " pounds the war drums! *BOOM-CRASH-BOOM*")

    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive then
            if info.faction == self.faction then
                -- Buff allies into a furious battle state
                world.set_mood_target(eid, "furious", 20)
            else
                -- Terrify civilian or hostile outsiders
                world.set_mood_target(eid, "terrified", 15)
            end
        end
    end
    return true
end

function do_tick()
    local acted = false
    if world.tick % 10 == 0 then acted = sound_drums() end
    if world.tick % 50 == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            world.move_to(exits[util.rand_int(#exits) + 1])
            acted = true
        end
    end
    if acted then return { util.event("profession_action", { profession = "clan_drummer" }) } end
    return {}
end

return do_tick()
