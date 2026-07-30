-- deity_core.lua
-- Data-driven master AI script for the tenants of the Fractured Vault.

-- Helper to extract what this specific god actually needs from its pre-seeded memory strings
local function get_divine_needs()
    local needs = {}
    -- We scan standard items your gods crave
    local possible_needs = {
        "cheap_wine", "lard", "melted_copper", "brackish_water", "dead_frogs",
        "salt_pork", "ink_vials", "herbal_tea", "wax_candles", "raw_meat",
        "bandages", "loud_shouting", "secret_journals", "old_cheese", "earplugs",
        "roasted_marrow", "ale_barrels", "anvil_to_smash", "greased_strings",
        "rotten_eggs", "attention", "violet_spores", "honey_pots", "gossip",
        "stamped_deeds", "dried_ink", "absolute_silence", "soothing_balm",
        "quiet_crying", "stolen_peaches", "flea_combs", "stale_bread",
        "tallow_candles", "blankets", "clear_night", "dry_pillows", "salted_fish",
        "raw_hide", "wooden_mallets", "zinc_coins", "rusty_nails", "cold_gruel",
        "copper_scraps", "rotten_meat", "charcoal", "polishing_wax", "parchment_deeds",
        "moldy_cheese", "bird_seed"
    }

    for _, item in ipairs(possible_needs) do
        if util.mem_get("need_" .. item) == "true" then
            table.insert(needs, item)
        end
    end
    return needs
end

local function process_divine_neglect()
    local needs = get_divine_needs()
    local state_changed = false

    -- Check if we have our required offerings in our inventory
    for _, item in ipairs(needs) do
        -- count_item wrapper or inventory scanning if exposed to your world API
        if world.has_material(item) then
            world.use_item(item)
            util.log(self.name .. " greedily consumed an offering of " .. item .. ".")

            -- Lower frustration levels upon a successful sacrifice consumption pass
            local anger = tonumber(util.mem_get("divine_anger") or "0")
            anger = math.max(0, anger - 20)
            util.mem_set("divine_anger", tostring(anger))
            state_changed = true
        else
            -- If a specific critical need is completely missing, tick up their cosmic fury
            local anger = tonumber(util.mem_get("divine_anger") or "0")
            anger = anger + 5
            util.mem_set("divine_anger", tostring(anger))
        end
    end

    -- Evaluate their resulting mood based on localized frustration numbers
    local final_anger = tonumber(util.mem_get("divine_anger") or "0")
    if final_anger > 80 then
        util.set_mood("furious", 30)
    elseif final_anger > 40 then
        util.set_mood("annoyed", 30)
    else
        util.set_mood("relaxed", 30)
    end
end

local function unleash_celestial_consequences()
    if self.mood ~= "furious" then return end

    -- Roll a random chance so they don't break the world every single tick pass
    if util.rand_int(100) > 15 then return end

    util.log("[WRATH] " .. self.name .. " is completely neglected and unleashes a localized divine penalty!")

    -- Direct conditional behavioral triggers mapping out their distinct structural domains
    if self.id == "seus_crackbolt" or self.id == "choke_the_drenched" then
        -- Force dynamic weather anomalies across random mortal paths
        world.damage_location(self.id, 5) -- Crack static lightning inside their own room
        util.log(self.name .. " caused a leaky cosmic pipe breakdown. Thunderstorms rip across the realm.")
    elseif self.id == "vaicna_the_unwashed" or self.id == "haydes_the_hoarder" then
        -- Boost the undead skeleton spawning loops or rot metrics
        util.log(self.name .. " curses the mortal realms. Corpses inside local crypts rot faster.")
    elseif self.id == "othena_the_pedantic" then
        -- Grammatical frustration causes local wizards' spells to fail completely
        util.log(self.name .. " alters the laws of grammar. Magic users feel a deep, throbbing headache.")
    elseif self.id == "oriz_the_bloodshot" then
        -- Drive nearby room occupants into unprovoked fistfights
        world.damage_location(self.id, 10)
        util.log(self.name .. " incites an unwarranted brawl inside the High Hall!")
    end
end

local function do_tick()
    -- 1. Gods process their internal consumption or neglect pools every 20 ticks
    if world.tick % 20 == 0 then
        process_divine_neglect()
    end

    -- 2. Unleash consequences if neglected for too long
    if world.tick % 40 == 0 then
        unleash_celestial_consequences()
    end

    -- 3. High Hall Room Dynamics: Brawling gods will naturally scrap with each other if angry
    if world.tick % 15 == 0 and self.mood == "furious" then
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive and info.profession == "deity" then
                        util.log(self.name ..
                        " blindly throws an object at " .. info.name .. " out of sheer petty frustration!")
                        world.attack(self.id, eid)
                        break
                    end
                end
            end
        end
    end
end

do_tick()
