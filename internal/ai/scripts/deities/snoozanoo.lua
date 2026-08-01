-- snoozanoo.lua
-- Snoozano-o the Lethargic: god of damp gales and clogged drains.
-- A garbled parody of Susanoo.

local function try_divine_conception(events)
    if world.tick % 400 ~= 200 then return end
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
        if self.cause and self.cause ~= "" then
            world.set_cause(target_id, self.cause)
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

    if world.tick % 400 ~= 0 then
        util.log(self.name .. " sleeps. Zzzzzz. A damp wind escapes his nostrils.")
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        return events
    end

    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return events end
    util.log(self.name .. " AWAKENS! \"WHERE AM I?! WHAT YEAR IS IT?!\"")

    local roll = util.rand_int(10)

    if roll < 4 then
        util.log(self.name .. " summons a TERRIBLE STORM! Wind howls! Rain... is lukewarm and slightly greasy.")
        world.damage_location(self.id, 35)
        util.log("Everything at " .. dest .. " is now wet, windswept, and smells faintly of old drain water.")
        table.insert(events, util.event("world", {
            source = self.id,
            data = { location = dest, event = "damp_gale" }
        }))
    elseif roll < 7 then
        util.log(self.name .. " rages through " .. dest .. " looking for an eight-headed serpent to fight!")
        local nearby = world.nearby_entities()
        if nearby then
            local found_serpent = false
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive and info.species == "hydra" then
                        util.log(self.name .. " FINDS A HYDRA! \"CLOSE ENOUGH! EN GARDE, MULTI-HEADED SCUM!\"")
                        world.attack(self.id, eid)
                        found_serpent = true
                        table.insert(events, util.event("combat", {
                            source = self.id,
                            target = eid,
                            data = { event = "serpent_hunt" }
                        }))
                        break
                    end
                end
            end
            if not found_serpent then
                util.log(self.name .. " finds no serpent. Gets angry at a tree instead.")
                world.damage_location(self.id, 10)
            end
        end
    else
        util.log(self.name .. " accidentally clogs every drain at " .. dest .. " with divine hair")
        world.damage_location(self.id, 15)
        util.log("Floodwater rises ankle-deep. Mortals are not pleased.")
        table.insert(events, util.event("world", {
            source = self.id,
            data = { location = dest, event = "clogged_drains" }
        }))
    end

    util.log(self.name .. " yawns enormously and gets banished back to the divine realm for being too loud")
    world.move_to(self.home)
    util.log(self.name .. " falls asleep immediately upon arrival. Zzzzzz.")
    try_divine_conception(events)
    return events
end

return do_tick()
