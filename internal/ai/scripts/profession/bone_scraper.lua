-- bone_scraper.lua
-- Battlefield scavenger that strips weapons and valuables off the fallen.

local function scrape_corpses()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            -- Explicitly target dying or heavily crippled targets to steal their weapons
            if info and info.alive and info.hp < 30 then
                util.log(self.name ..
                " circles the wounded " .. info.name .. ". 'You won't be needing those weapons much longer...'")
                util.set_mood("calculating", 15)

                local weapon = world.try_buy(eid, "iron_sword") or world.try_buy(eid, "orc_cleaver")
                if weapon and weapon.done then
                    util.log(self.name .. " successfully disarmed and looted the target.")
                else
                    -- If they refuse to give it up, finish them off
                    world.attack(self.id, eid)
                end
                return true
            end
        end
    end
    return false
end

function do_tick()
    local acted = false
    if world.tick % 7 == 0 then acted = scrape_corpses() end
    if world.tick % 35 == 0 and util.rand_int(100) < 50 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            world.move_to(exits[util.rand_int(#exits) + 1])
            acted = true
        end
    end
    if acted then return { util.event("profession_action", { profession = "bone_scraper" }) } end
    return {}
end

return do_tick()
