-- werewolf.lua
-- Shape-shifting script that activates lethal combat arrays under night cycles.

local function run_primal_hunt()
    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            -- Attack ABSOLUTELY ANYONE that is living when transformed
            if info and info.alive and info.faction ~= "werewolves" then
                util.log(self.name .. " howls under the night cycle and tears into " .. info.name .. " with razor claws!")
                util.set_mood("furious", 20)
                world.attack(self.id, eid)
                return
            end
        end
    end
end

local function do_tick()
    -- Day Phase: Act completely innocent to avoid town guards
    if world.phase == "day" then
        local was_transformed = util.mem_get("transformed")
        if was_transformed then
            util.log(self.name .. " shifts back into human form, breathing heavily in exhaustion.")
            util.mem_set("transformed", nil)
            util.set_mood("relaxed", 30)
        end

        -- Normal peasant pathing routines
        if world.tick % 40 == 0 and util.rand_int(100) < 30 then
            local exits = world.exits_from(self.loc_id)
            if exits and #exits > 0 then world.move_to(exits[util.rand_int(#exits) + 1]) end
        end
        return
    end

    -- Night Phase: Immediate transformation and bloodlust
    if world.phase == "night" then
        if not util.mem_get("transformed") then
            util.log(self.name .. "'s eyes turn crimson as thick fur erupts across their flesh!")
            util.mem_set("transformed", true)
        end

        if world.tick % 6 == 0 then
            run_primal_hunt()
        end
    end
end

do_tick()
