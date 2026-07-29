-- bread_weaver.lua
-- Radical peasant militia script managing logistics, staging strikes, and defending granaries.

local OPPRESSORS = { "tax_man", "customs", "royal_mage" }

local function audit_the_slums()
    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                -- 1. Defend the neighborhood from incoming Crown oppressors instantly
                for _, target_prof in ipairs(OPPRESSORS) do
                    if info.profession == target_prof then
                        util.log(self.name ..
                            " sounds the iron dinner bell! 'The bloodsuckers are here! Protect the grain!'")
                        util.set_mood("furious", 20)

                        -- Use area damage mechanics if available, or straight attack
                        world.damage_location(self.id, 5)
                        world.attack(self.id, eid)
                        return
                    end
                end

                -- 2. Material logistics loop: Buy grain from poor local farmers to keep it out of royal hands
                if info.profession == "farmer" and info.faction ~= "bleeding_quill" then
                    local trade = world.try_buy(eid, "raw_grain")
                    if trade and trade.done then
                        util.log(self.name .. " secured a raw grain sack from a local farm comrade.")
                        util.set_mood("relaxed", 15)
                        return
                    end
                end
            end
        end
    end
end

local function manage_bakery_and_strikes()
    -- Check the status of current regional exploitation fields in memory
    local strike_active = util.mem_get("general_strike")

    if not strike_active then
        -- Normal Operations: Bake bread to feed the community and stock up resources
        if world.has_material("raw_grain") then
            world.craft("recipe_emergency_loaf")
            util.log(self.name .. " pulled a steaming, dense peasant loaf out of the brick oven.")
        end

        -- Trigger condition: If the crown's heretical bureaucrats squeeze too hard, call a strike
        -- Simulated here by checking if a local tax man has recently harassed someone nearby
        local recent_taxations = util.mem_get("crown_aggression_counter") or 0
        if recent_taxations > 3 or util.rand_int(100) < 5 then
            util.log(self.name .. " hangs the 'CLOSED DUE TO TYRANNY' sign. STRIKE INITIATED!")
            util.mem_set("general_strike", true)
            util.set_mood("furious", 60)
            world.quest_progress("starve_the_castle_q", "shut_down_mills", 1)
        end
    else
        -- Strike Operations: Refuse to serve food to anyone outside the working class
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                local info = world.entity_info(eid)
                -- If a rich noble or merchant walks into the bakery during a strike, throw them out
                if info and info.alive and (info.profession == "noble" or info.profession == "merchant") then
                    world.talk_to(eid)
                    util.log(self.name ..
                        " brandishes a massive iron rolling pin at " ..
                        info.name .. ": 'No bread for hoarders until taxes fall!'")
                    world.attack(self.id, eid)
                    return
                end
            end
        end

        -- Strike resolution safety: Cool down or resolve the labor issue over time
        if util.rand_int(100) < 15 then
            util.log(self.name .. " received concessions from the guilds. Ending strike operations.")
            util.mem_set("general_strike", nil)
            util.mem_set("crown_aggression_counter", 0)
            util.set_mood("happy", 30)
        end
    end
end

local function do_tick()
    -- 1. Audit local area for tax men or raw grain buying every 10 ticks
    if world.tick % 10 == 0 then
        audit_the_slums()
    end

    -- 2. Process culinary production or coordinate political resistance actions
    if world.tick % 18 == 0 then
        manage_bakery_and_strikes()
    end

    -- 3. Mutual aid distribution: Feed hungry/injured dynamic allies nearby
    if world.tick % 25 == 0 and world.has_material("emergency_loaf") then
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                local info = world.entity_info(eid)
                -- If they are a low-income comrade or injured ally faction, distribute rations
                if info and info.alive and (info.profession == "beggar" or info.faction == "smog_iron_cartel") then
                    if info.hp and info.hp < 50 then
                        world.feed(eid)
                        world.heal(eid, 15)
                        util.log(self.name .. " shared an emergency loaf to heal a wounded " .. info.profession)
                        break
                    end
                end
            end
        end
    end
end

do_tick()
