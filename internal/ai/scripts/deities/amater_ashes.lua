-- amater_ashes.lua
-- Amater-Ashes the Dim: goddess of flickering lanterns and bad sunburns.
-- A garbled parody of Amaterasu.

local function try_divine_conception(events)
    if world.tick % 280 ~= 140 then return end
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

    if world.tick % 280 ~= 0 then
        return events
    end

    local in_cave = util.mem_get("in_cave")
    local roll = util.rand_int(10)

    if in_cave and roll < 7 then
        util.log(self.name .. " remains hidden in the Celestial Cave. Darkness spreads across the land.")
        util.log("Mortals stumble around in the dark. Someone walks into a pillar. It's getting embarrassing.")
        table.insert(events, util.event("world", {
            source = self.id,
            data = { event = "divine_darkness" }
        }))
        return events
    end

    if in_cave and roll >= 7 then
        util.log("Someone brought a slightly shiny rock near the cave entrance!")
        util.log(self.name .. " peeks out. \"Is that... is that a mirror? It's pretty dim, but I suppose it'll do.\"")
        util.mem_set("in_cave", false)
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { event = "lured_out" }
        }))
    end

    if not in_cave and roll < 2 then
        util.log(self.name .. " retreats into the Celestial Cave. \"I'm not hiding! I'm making a STATEMENT!\"")
        util.log(self.name .. " sets own mood to 'hiding'. The world dims slightly.")
        util.mem_set("in_cave", true)
        util.set_mood("hiding", 100)
        table.insert(events, util.event("mood", {
            source = self.id,
            data = { mood = "hiding", event = "cave_retreat" }
        }))
        return events
    end

    if not in_cave then
        local mortal_locs = world.find_mortal_locations()
        if not mortal_locs or #mortal_locs == 0 then return events end
        local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
        if not world.move_to(dest) then return events end

        if roll < 5 then
            util.log(self.name .. " reluctantly illuminates " .. dest .. ". The light is dim and flickery, like a lantern running low on oil.")
            local nearby = world.nearby_entities()
            if nearby then
                for _, eid in ipairs(nearby) do
                    if eid ~= self.id then
                        local info = world.entity_info(eid)
                        if info and info.alive then
                            util.log(info.name .. " gets a mild sunburn from " .. self.name .. "'s presence. You're welcome.")
                        end
                    end
                end
            end
            table.insert(events, util.event("ambient", {
                source = self.id,
                data = { location = dest, event = "dim_illumination" }
            }))
        elseif roll < 8 then
            util.log(self.name .. " checks " .. dest .. " for shiny objects. Finds nothing interesting. \"Mortals have NO taste.\"")
            table.insert(events, util.event("ambient", {
                source = self.id,
                data = { event = "shiny_scouting" }
            }))
        else
            util.log(self.name .. " stands at " .. dest .. " looking vaguely luminous and deeply uninterested.")
            util.log("A mortal child asks " .. self.name .. " to light their candle. " .. self.name .. " sighs for three full minutes before complying.")
        end

        world.move_to(self.home)
    end

    try_divine_conception(events)
    return events
end

return do_tick()
