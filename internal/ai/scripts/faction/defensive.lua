-- defensive.lua
-- High-alert survival matrix for fleeing danger and avoiding confrontation.

local function evaluate_threats()
    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            -- Escape from active combat classes or aggressive monstrous factions
            if info and info.alive and (info.profession == "guard" or info.profession == "bandit_chief" or info.faction == "rust_walkers") then
                util.log(self.name .. " senses extreme structural danger from " .. info.name .. "! Fleeing sector!")
                util.set_mood("panicked", 15)

                local exits = world.exits_from(self.loc_id)
                if exits and #exits > 0 then
                    world.move_to(exits[util.rand_int(#exits) + 1])
                end
                return true
            end
        end
    end
    return false
end

function do_tick()
    if world.tick % 4 == 0 then
        local fled = evaluate_threats()

        if not fled and world.tick % 30 == 0 then
            util.set_mood("relaxed", 20)
        end
        return fled and {util.event("flee", {})} or {}
    end
    return {}
end

return do_tick()
