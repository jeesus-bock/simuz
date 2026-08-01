-- baa_hamut.lua
-- Baa-Hamut the Blunderer: god of misplaced justice and broken scales.
-- A garbled parody of Bahamut.

local function try_divine_conception(events)
    if world.tick % 210 ~= 105 then return end
    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then return end
    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return end
    local nearby = world.nearby_entities()
    if not nearby then world.move_to(self.home) return end
    local females = {}
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.gender == "female" then
                table.insert(females, eid)
            end
        end
    end
    if #females == 0 then world.move_to(self.home) return end
    local target_id = females[util.rand_int(#females) + 1]
    local target_info = world.entity_info(target_id)
    local did_polymorph = false
    if self.species ~= target_info.species then
        world.polymorph(self.id, target_info.species)
        did_polymorph = true
    end
    if world.impregnate(self.id, target_id) then
        util.log("[DIVINE] " .. self.name .. " has impregnated " .. target_info.name .. " (" .. target_info.species .. ")")
        local cause = self.cause
        if (not cause or cause == "") then cause = util.mem_get("domain") end
        if cause and cause ~= "" then
            world.set_cause(target_id, cause)
        end
        table.insert(events, util.event("divine", {
            source = self.id,
            data = { mother = target_id, species = target_info.species, event = "divine_conception" }
        }))
    end
    if did_polymorph then
        world.revert_polymorph(self.id)
    end
    world.move_to(self.home)
end

function do_tick()
    local events = {}

    if world.tick % 210 ~= 0 then
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        return events
    end

    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return events end
    util.log(self.name .. " descends to " .. dest .. " with a loud, metallic CLANG. He has dropped his scales again.")

    local roll = util.rand_int(10)

    if roll < 3 then
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive then
                        local verdicts = {
                            "GUILTY! Of... wait, I forgot my glasses. What did you do again?",
                            "INNOCENT! Probably. The scales are broken. Everything reads 'guilty.'",
                            "I find you... RESPONSIBLE for the weather. Case closed.",
                            "You are hereby sentenced to... community service? Is that a thing here?",
                            "NOT GUILTY! Wait, that was the wrong person. GUILTY! No, wait..."
                        }
                        local verdict = verdicts[util.rand_int(#verdicts) + 1]
                        util.log(self.name .. " judges " .. info.name .. ": \"" .. verdict .. "\"")
                        table.insert(events, util.event("mood", {
                            source = self.id,
                            target = eid,
                            data = { event = "divine_judgment" }
                        }))
                        break
                    end
                end
            end
        end
    elseif roll < 5 then
        util.log(self.name .. " tries to protect the innocent at " .. dest)
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive and info.worship and info.worship > 10 then
                        if info.hp < info.max_hp then
                            local heal_amt = math.min(30, info.max_hp - info.hp)
                            world.heal(self.id, eid, heal_amt)
                            util.log(self.name .. " heals " .. info.name .. " by " .. heal_amt .. " HP! \"JUSTICE IS SERVED!\"")
                            util.log(self.name .. " accidentally knocks over a market stall in the process. \"Sorry! So sorry!\"")
                            table.insert(events, util.event("combat", {
                                source = self.id,
                                target = eid,
                                data = { event = "clumsy_healing", amount = heal_amt }
                            }))
                            break
                        end
                    end
                end
            end
        end
    elseif roll < 8 then
        util.log(self.name .. " searches for his reading glasses to better dispense justice")
        util.log("...checks pockets. Checks other pockets. Checks pockets that don't exist.")
        util.log("\"They were RIGHT HERE! I had them! Did someone steal my glasses?!\"")
        util.log("They are on his head. They have always been on his head.")
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { event = "glasses_search" }
        }))
    else
        util.log(self.name .. " attempts a dramatic entrance at " .. dest)
        util.log("Trips over his own tail. Crashes into a wall. The wall collapses.")
        world.damage_location(self.id, 10)
        util.log(self.name .. " stands up, dusts himself off. \"I MEANT to do that. Justice... required structural renovation.\"")
        table.insert(events, util.event("world", {
            source = self.id,
            data = { location = dest, event = "accidental_demolition" }
        }))
    end

    util.log(self.name .. " trips once more on the way out, scattering scales everywhere")
    world.move_to(self.home)
    try_divine_conception(events)
    return events
end

return do_tick()
