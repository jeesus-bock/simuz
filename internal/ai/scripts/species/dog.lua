-- dog.lua
-- Domestic canine script handling pathing tracking, alerting, and companion routines.

local MONSTERS = { "vampire", "werewolf", "cultist" }

local function run_canine_alerts()
    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                -- 1. Bark aggressively if a monster structure approaches
                for _, monster_prof in ipairs(MONSTERS) do
                    if info.profession == monster_prof or info.faction == "rust_walkers" then
                        util.log(self.name .. " growls fiercely and barks loudly at " .. info.name .. "!")
                        util.set_mood("furious", 10)

                        -- Alert the entire map location via low-level area noise if possible
                        world.damage_location(self.id, 1)
                        return
                    end
                end

                -- 2. Follow a friendly local human guard or town explorer if left alone
                if not util.mem_get("companion_target") and info.profession == "guard" then
                    util.log(self.name .. " wags its tail and begins following " .. info.name)
                    util.mem_set("companion_target", eid)
                    util.set_mood("happy", 40)
                end
            end
        end
    end
end

local function do_tick()
    if world.tick % 6 == 0 then
        run_canine_alerts()
    end

    -- Companion path tracking: if following someone, automatically drag/hop to their sector
    if world.tick % 12 == 0 then
        local master_id = util.mem_get("companion_target")
        if master_id then
            local master_info = world.entity_info(master_id)
            if master_info and master_info.alive and master_info.location_id ~= self.loc_id then
                world.move_to(master_info.location_id)
                util.log(self.name .. " trots behind its companion into " .. master_info.location_id)
            elseif not master_info or not master_info.alive then
                util.mem_set("companion_target", nil) -- Reset tracker if companion dies
            end
        end
    end
end

do_tick()
