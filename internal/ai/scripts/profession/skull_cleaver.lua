-- skull_cleaver.lua
-- Brutal enforcer that executes cowards and attacks trespassers instantly.

local function enforce_discipline()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                -- Execution of internal cowards
                if info.faction == self.faction and (info.mood == "terrified" or info.mood == "panicked") then
                    util.log(self.name .. " roars at " .. info.name .. "! 'Cowardice is punishable by death!'")
                    world.attack(self.id, eid)
                    return true
                    -- Obliteration of dynamic enemy trespassers
                elseif info.faction ~= self.faction and info.faction ~= "civilian" then
                    util.log(self.name .. " defends the territory with absolute brutality!")
                    world.attack(self.id, eid)
                    return true
                end
            end
        end
    end
    return false
end

function do_tick()
    local acted = false
    if world.tick % 4 == 0 then acted = enforce_discipline() end
    -- Guards rarely leave their designated post/room location
    if world.tick % 100 == 0 and util.rand_int(100) < 10 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            world.move_to(exits[util.rand_int(#exits) + 1])
            acted = true
        end
    end
    if acted then return { util.event("profession_action", { profession = "skull_cleaver" }) } end
    return {}
end

return do_tick()
