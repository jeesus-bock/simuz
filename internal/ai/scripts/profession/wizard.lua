-- wizard_scholar.lua
-- Rogue Wizard Academia script utilizing astrology, hexes, and black-market scrying.

local function read_the_stars()
    local phase = world.phase
    local weather = world.weather(self.loc_id)

    -- Wizards can only map cosmic alignments under clear night skies
    if phase == "night" and weather == "clear" then
        util.set_mood("focused", 20)

        -- Generate research progression points via persistence keys
        local star_notes = util.mem_get("astral_notes") or 0
        star_notes = star_notes + 1
        util.mem_set("astral_notes", star_notes)
        util.log(self.name .. " is charting celestial alignments. Total research points: " .. star_notes)

        -- If they have enough research, progress the global forbidden magic quest
        if star_notes % 5 == 0 then
            world.quest_progress("cosmic_cataclysm_q", "gather_star_charts", 1)
            util.log(self.name .. " unlocked a hidden cosmic transit calculation!")
        end
        return true
    end
    return false
end

local function harvest_cosmic_reagents()
    local weather = world.weather(self.loc_id)

    -- High-level cosmic energies crystallize during violent ambient storms
    if weather == "thunderstorm" or weather == "storm" then
        util.log(self.name .. " is raising their copper staff to harvest lightning residuals!")
        world.add_item(self.id, "charged_quartz", 1)
        util.set_mood("ecstatic", 10)

        -- Instantly try to craft a high-value Scrying Mirror if components match
        if world.has_material("charged_quartz") and world.has_material("glass_pane") then
            world.craft("recipe_scrying_mirror")
            util.log(self.name .. " constructed a Black-Market Scrying Mirror from residual storm matrixes.")
        end
        return true
    end
    return false
end

local function handle_academic_rivalry()
    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                -- 1. Ruthlessly attack or hex mainstream institutional Royal Mages
                if info.profession == "royal_mage" then
                    util.log(self.name ..
                    " screams at " ..
                    info.name .. ": 'Institutional puppet! Your structural spell theories are ancient lies!'")
                    world.attack(self.id, eid)
                    util.set_mood("furious", 15)
                    return
                end

                -- 2. Run blackmail operations on wealthy traveling Nobles or Merchants
                if info.profession == "noble" or info.profession == "merchant" then
                    -- Check if we have an item to blackmail them with
                    if world.has_material("scrying_mirror") then
                        world.talk_to(eid)
                        util.log(self.name ..
                        " uses a hidden scrying mirror to threaten " .. info.name .. " with leaked dark secrets.")

                        -- Force them to pay hush money (try_sell dirty laundry text documents)
                        local extort = world.try_sell(eid, "hush_money_scroll")
                        if extort and extort.done then
                            util.log(self.name ..
                            " successfully blackmailed " .. info.name .. " for dynamic guild funding gold.")
                            world.quest_progress("fund_the_revolution_q", "extort_wealth", 1)
                        end
                        return
                    end
                end
            end
        end
    end
end

local function do_tick()
    -- 1. Academic rivalries and extortion checks occur frequently
    if world.tick % 10 == 0 then
        handle_academic_rivalry()
    end

    -- 2. Environmental harvesting checks
    if world.tick % 15 == 0 then
        harvest_cosmic_reagents()
    end

    -- 3. Nocturnal sky observation phase
    if world.tick % 20 == 0 then
        local observing = read_the_stars()

        -- If it's cloudy or raining at night, the wizard gets deeply frustrated
        if not observing and world.phase == "night" then
            util.log(self.name .. " mutters curses at the overcast skies obscuring the stars.")
            util.set_mood("annoyed", 10)

            -- Wander to an adjacent sector hoping for clearer viewing coordinates
            local exits = world.exits_from(self.loc_id)
            if exits and #exits > 0 then
                world.move_to(exits[util.rand_int(#exits) + 1])
            end
        end
    end

    return false
end

return do_tick()
