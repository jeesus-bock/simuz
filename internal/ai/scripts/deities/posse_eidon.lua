-- posse_eidon.lua
-- Posse-Eidon the Silt-King: god of stagnant puddles and well collapses.
-- A garbled parody of Poseidon.

local function try_divine_conception(events)
    if world.tick % 250 ~= 125 then return end
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
    if self.species ~= target_info.species then
        world.polymorph(self.id, target_info.species)
    end
    if world.impregnate(self.id, target_id) then
        util.log("[DIVINE] " .. self.name .. " has impregnated " .. target_info.name .. " (" .. target_info.species .. ")")
        table.insert(events, util.event("divine", {
            source = self.id,
            data = { mother = target_id, species = target_info.species, event = "divine_conception" }
        }))
    end
    if self.species ~= "divine" then
        world.revert_polymorph(self.id)
    end
    world.move_to(self.home)
end

function do_tick()
    local events = {}

    if world.tick % 250 ~= 0 then
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        return events
    end

    local roll = util.rand_int(10)
    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]

    if roll < 3 then
        if not world.move_to(dest) then return events end
        util.log(self.name .. " rises from a stagnant puddle at " .. dest)
        world.damage_location(self.id, 15)
        util.log(self.name .. " shakes the ground! Wells collapse, mud sprays everywhere!")
        table.insert(events, util.event("world", {
            source = self.id,
            data = { location = dest, event = "silt_quake" }
        }))
        world.move_to(self.home)
    elseif roll < 6 then
        if not world.move_to(dest) then return events end
        util.log(self.name .. " summons a tidal wave of brown, murky water at " .. dest)
        world.damage_location(self.id, 20)
        util.log(self.name .. " floods the area! Everything smells of pond scum!")
        table.insert(events, util.event("world", {
            source = self.id,
            data = { location = dest, event = "mud_flood" }
        }))
        world.move_to(self.home)
    elseif roll < 8 then
        util.log(self.name .. " transforms into a sad, wet horse and gallops clumsily through the divine realm")
        world.polymorph(self.id, "horse")
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { event = "horse_polymorph" }
        }))
        world.move_to(dest)
        util.log(self.name .. " as a horse, slips in the mud at " .. dest .. " and falls over")
        world.revert_polymorph(self.id)
        world.move_to(self.home)
    else
        if not world.move_to(dest) then return events end
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive and info.worship and info.worship < 10 then
                        util.log(self.name .. " glares at " .. info.name .. " for insufficient worship! DROWNS THEM IN SILT!")
                        world.damage_location(self.id, 30)
                        table.insert(events, util.event("combat", {
                            source = self.id,
                            target = eid,
                            data = { event = "silt_punishment" }
                        }))
                        break
                    end
                end
            end
        end
        world.move_to(self.home)
    end

    util.log(self.name .. " sinks back into the muck and returns to the divine realm")
    try_divine_conception(events)
    return events
end

return do_tick()
