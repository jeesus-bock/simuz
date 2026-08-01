-- raijin_the_rattler.lua
-- Raijin the Rattler: demon of cracked drums and tinnitus.
-- A garbled parody of Raijin.

local function try_divine_conception(events)
    if world.tick % 140 ~= 70 then return end
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

    if world.tick % 140 ~= 0 then
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        return events
    end

    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return events end
    util.log(self.name .. " arrives at " .. dest .. " dragging a set of very cracked drums")

    local roll = util.rand_int(10)

    if roll < 4 then
        util.log(self.name .. " attempts to play a thunderous drumroll...")
        util.log("...but the drumskin is torn. It makes a sad flapping noise.")
        util.log(self.name .. " tries again. FLAP FLAP FLAP. Not very intimidating.")
        util.log("A nearby crow caws louder. " .. self.name .. " is humiliated.")
        world.damage_location(self.id, 3)
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { location = dest, event = "failed_drumroll" }
        }))
    elseif roll < 6 then
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive then
                        util.log(self.name .. " jumps out at " .. info.name .. " and yells \"BOO!\" while shaking a broken rattle")
                        if info.level and info.level < 5 then
                            util.log(info.name .. " is mildly startled. " .. self.name .. " considers this a major victory.")
                            util.set_mood("frightened", 15)
                        else
                            util.log(info.name .. " doesn't even flinch. " .. self.name .. " tries a scarier face. It doesn't help.")
                        end
                        table.insert(events, util.event("mood", {
                            source = self.id,
                            target = eid,
                            data = { mood = "frightened", event = "startle_attempt" }
                        }))
                        break
                    end
                end
            end
        end
    elseif roll < 8 then
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive and info.level and info.level <= 3 then
                        util.log(self.name .. " tries to steal " .. info.name .. "'s belly button! Wait, that's not how it works. Steals their lunch instead.")
                        world.steal(self.id, eid)
                        table.insert(events, util.event("combat", {
                            source = self.id,
                            target = eid,
                            data = { event = "belly_button_theft" }
                        }))
                        break
                    end
                end
            end
        end
    else
        util.log(self.name .. " attempts a mighty thunder-clap with his drums!")
        util.log("BANG! ...well, more of a 'pfft.' The drums rattle pathetically.")
        world.damage_location(self.id, 5)
        util.log("A small crack appears in a nearby pot. " .. self.name .. " claims victory.")
        table.insert(events, util.event("world", {
            source = self.id,
            data = { location = dest, event = "weak_thunder" }
        }))
    end

    util.log(self.name .. " shuffles away, drums clanking sadly behind him")
    world.move_to(self.home)
    try_divine_conception(events)
    return events
end

return do_tick()
