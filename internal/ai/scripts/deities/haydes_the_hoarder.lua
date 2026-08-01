-- haydes_the_hoarder.lua
-- Haydes the Hoarder: god of cluttered graves and unpaid debts.
-- A garbled parody of Hades.

local function try_divine_conception(events)
    if world.tick % 220 ~= 110 then return end
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

    if world.tick % 220 ~= 0 then
        return events
    end

    local roll = util.rand_int(10)

    if roll < 4 then
        util.log(self.name .. " sits in the Underworld surrounded by piles of junk")
        local junk = {
            "37 mismatched left shoes",
            "a crate of unpaid IOUs from the 4th age",
            "200 pounds of unsorted ear bones",
            "a filing cabinet full of soul tax returns",
            "Cerberus's chew toy (all three heads want it)",
            "a broken Helm of Invisibility (it's just a regular helmet now)"
        }
        local item = junk[util.rand_int(#junk) + 1]
        util.log(self.name .. " sorts through " .. item .. ". \"This is VERY valuable. Don't touch it.\"")
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { event = "hoarding_sorting" }
        }))
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then return events end
    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return events end
    util.log(self.name .. " emerges from the ground at " .. dest .. " clutching a ledger")

    local nearby = world.nearby_entities()
    if not nearby then
        world.move_to(self.home)
        return events
    end

    if roll < 7 then
        local has_dead = false
        for _, eid in ipairs(nearby) do
            if eid ~= self.id then
                local info = world.entity_info(eid)
                if info and not info.alive then
                    util.log(self.name .. " guards the corpse of " .. info.name .. " jealously. \"MINE. My dead. Nobody touch.\"")
                    has_dead = true
                    table.insert(events, util.event("ambient", {
                        source = self.id,
                        target = eid,
                        data = { event = "dead_guarding" }
                    }))
                end
            end
        end
        if not has_dead then
            util.log(self.name .. " finds no dead to guard. Checks behind bushes. Nothing. Disappointing.")
        end
    else
        for _, eid in ipairs(nearby) do
            if eid ~= self.id then
                local info = world.entity_info(eid)
                if info and info.alive then
                    util.log(self.name .. " accosts " .. info.name .. ": \"You owe the Underworld 3 silver and 7 copper in soul processing fees!\"")
                    util.log(info.name .. " has no idea what " .. self.name .. " is talking about.")
                    util.log(self.name .. " makes a note in the ledger. \"Interest is compounding.\"")
                    world.attack(self.id, eid)
                    table.insert(events, util.event("combat", {
                        source = self.id,
                        target = eid,
                        data = { event = "debt_collection" }
                    }))
                    break
                end
            end
        end
    end

    util.log(self.name .. " sinks back into the earth, taking a suspicious amount of dirt with him")
    world.move_to(self.home)
    try_divine_conception(events)
    return events
end

return do_tick()
