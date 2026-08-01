-- raider_vanguard.lua
-- Aggressive territory runner that plunders while moving messages or orders.

local function pillage_en_route()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            -- Target weak human/civilian professions on the road
            if info and info.alive and (info.profession == "merchant" or info.profession == "traveler" or info.profession == "farmer") then
                util.log(self.name .. " intercepts a weakling! 'Hand over your rations or feed the crows!'")
                util.set_mood("furious", 25)

                world.attack(self.id, eid)
                local loot = world.try_buy(eid, "grain")
                if loot and loot.done then
                    util.log(self.name .. " stole supplies from " .. info.name)
                end
                return true
            end
        end
    end
    return false
end

function do_tick()
    local acted = false
    if world.tick % 5 == 0 then acted = pillage_en_route() end
    -- Moves at a blistering pace (every 15 ticks) compared to ordinary entities
    if world.tick % 15 == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            world.move_to(exits[util.rand_int(#exits) + 1])
            acted = true
        end
    end
    if acted then return { util.event("profession_action", { profession = "raider_vanguard" }) } end
    return {}
end

return do_tick()
