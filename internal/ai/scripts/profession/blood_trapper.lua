-- blood_trapper.lua
-- Aggressive monster tracker that hunts down high-value targets for hide and blood.

local function hunt_beasts()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            -- Actively hunt down beasts or wild monsters in the area
            if info and info.alive and (info.faction == "beast" or info.species == "wolf" or info.species == "bear") then
                util.log(self.name .. " spots tracks! 'Fresh blood in the wind! Time to skin a beast!'")
                util.set_mood("furious", 20)

                -- Ambush bonus damage
                world.attack(self.id, eid)
                world.attack(self.id, eid)

                local harvest = world.try_buy(eid, "beast_hide")
                if harvest and harvest.done then
                    util.log(self.name .. " strips the raw pelt directly from the target.")
                end
                return true
            end
        end
    end
    return false
end

function do_tick()
    local acted = false
    if world.tick % 6 == 0 then acted = hunt_beasts() end
    if world.tick % 30 == 0 and util.rand_int(100) < 60 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then world.move_to(exits[util.rand_int(#exits) + 1]) acted = true end
    end
    if acted then return {util.event("profession_action", {profession = "blood_trapper"})} end
    return {}
end
return do_tick()
