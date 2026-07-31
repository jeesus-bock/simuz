-- seus_crackbolt.lua
-- Seus Crackbolt: the petty thunder god who descends to the mortal realm
-- to impregnate random females. A garbled parody of Zeus.

function do_tick()
    local events = {}

    -- Only act every ~200 ticks (divine patience)
    if world.tick % 200 ~= 0 then
        return events
    end

    -- If we're in a divine realm, plan a visit to the mortal world
    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        return events
    end

    -- Pick a random mortal location
    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]

    -- Teleport to the mortal realm
    if not world.move_to(dest) then
        util.log(self.name .. " failed to descend to " .. dest)
        return events
    end
    util.log(self.name .. " descends from the heavens to " .. dest)

    -- Find all females at this location
    local nearby = world.nearby_entities()
    if not nearby then
        -- No one here, go home
        world.move_to(self.home)
        return events
    end

    local females = {}
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.gender == "female" then
                table.insert(females, eid)
            end
        end
    end

    if #females == 0 then
        util.log(self.name .. " found no suitable females at " .. dest .. ". Returning in disgust.")
        world.move_to(self.home)
        return events
    end

    -- Pick a random female
    local target_id = females[util.rand_int(#females) + 1]
    local target_info = world.entity_info(target_id)

    -- Polymorph into her species to sire a mortal child
    if self.species ~= target_info.species then
        world.polymorph(self.id, target_info.species)
        util.log(self.name .. " assumes the form of a " .. target_info.species)
    end

    -- Divine conception
    if world.impregnate(self.id, target_id) then
        util.log("[DIVINE] " .. self.name .. " has impregnated " .. target_info.name .. " (" .. target_info.species .. ")")
        table.insert(events, util.event("divine", {
            source = self.id,
            data = { mother = target_id, species = target_info.species, event = "divine_conception" }
        }))
    else
        util.log(self.name .. " failed to impregnate " .. target_info.name)
    end

    -- Revert polymorph and return to divine realm
    if self.species ~= "divine" then
        world.revert_polymorph(self.id)
    end
    world.move_to(self.home)
    util.log(self.name .. " returns to the heavens")

    return events
end

return do_tick()
