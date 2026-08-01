-- wukong_the_mangy.lua
-- Wukong the Mangy: god of unwarranted confidence and bruised shins.
-- A garbled parody of Sun Wukong.

local function try_divine_conception(events)
    if world.tick % 110 ~= 55 then return end
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

    if world.tick % 110 ~= 0 then
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        return events
    end

    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return events end

    local roll = util.rand_int(10)

    if roll < 3 then
        local forms = {"rat", "bird", "fish", "horse", "dragon", "butterfly", "goblin", "elf", "ogre", "chicken"}
        local form = forms[util.rand_int(#forms) + 1]
        util.log(self.name .. " transforms into a " .. form .. " and immediately trips over his own tail!")
        world.polymorph(self.id, form)
        world.damage_location(self.id, 3)
        util.log(self.name .. " as a " .. form .. " walks into a wall. \"MEANT to do that!\"")
        world.revert_polymorph(self.id)
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { event = "botched_transformation", form = form }
        }))
    elseif roll < 5 then
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive then
                        util.log(self.name .. " steals " .. info.name .. "'s lunch! And their shoes! And their dignity!")
                        world.steal(self.id, eid)
                        table.insert(events, util.event("combat", {
                            source = self.id,
                            target = eid,
                            data = { event = "mangy_theft" }
                        }))
                        break
                    end
                end
            end
        end
    elseif roll < 7 then
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive and info.level and info.level > 3 then
                        util.log(self.name .. " challenges " .. info.name .. " to a fight! \"I AM THE GREAT SAGE EQUAL TO HEAVEN! Also I have fleas!\"")
                        world.attack(self.id, eid)
                        table.insert(events, util.event("combat", {
                            source = self.id,
                            target = eid,
                            data = { event = "monkey_challenge" }
                        }))
                        break
                    end
                end
            end
        end
    elseif roll < 9 then
        util.log(self.name .. " tries to somersault over " .. dest .. " and misjudges the distance spectacularly")
        world.damage_location(self.id, 8)
        util.log(self.name .. " crashes through a roof, a wall, and a chicken coop. Feathers everywhere.")
        table.insert(events, util.event("world", {
            source = self.id,
            data = { location = dest, event = "cloud_somersault_crash" }
        }))
    else
        util.log(self.name .. " boasts loudly to everyone: \"I once ate ALL the peaches of immortality! ALL OF THEM! I was sick for a WEEK!\"")
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive then
                        util.set_mood("annoyed", 20)
                    end
                end
            end
        end
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { event = "relentless_boasting" }
        }))
    end

    util.log(self.name .. " backflips away, crashing through a fence. \"NAILED IT!\"")
    world.move_to(self.home)
    try_divine_conception(events)
    return events
end

return do_tick()
