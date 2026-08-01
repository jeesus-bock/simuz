-- oriz_the_bloodshot.lua
-- Oriz the Bloodshot: god of unwarranted brawls and bruised shins.
-- A garbled parody of Ares.

local function try_divine_conception(events)
    if world.tick % 120 ~= 60 then return end
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

    if world.tick % 120 ~= 0 then
        return events
    end

    if self.hp < self.max_hp / 3 then
        util.log(self.name .. " flees to the divine realm nursing a black eye!")
        world.move_to(self.home)
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        return events
    end

    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return events end
    util.log(self.name .. " crashes into " .. dest .. " swinging wildly!")

    local nearby = world.nearby_entities()
    if not nearby or #nearby < 2 then
        util.log(self.name .. " finds nobody worth punching. Kicks a rock instead.")
        world.move_to(self.home)
        return events
    end

    local targets = {}
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.conscious then
                table.insert(targets, eid)
            end
        end
    end

    if #targets == 0 then
        util.log(self.name .. " everyone here is already unconscious. BORING!")
        world.move_to(self.home)
        return events
    end

    local num_punches = math.min(3, #targets)
    for i = 1, num_punches do
        local tidx = util.rand_int(#targets) + 1
        local target_id = targets[tidx]
        local tinfo = world.entity_info(target_id)
        util.log(self.name .. " headbutts " .. (tinfo.name or target_id) .. " for absolutely no reason!")
        world.attack(self.id, target_id)
        table.remove(targets, tidx)
        table.insert(events, util.event("combat", {
            source = self.id,
            target = target_id,
            data = { event = "unwarranted_brawl" }
        }))
        if #targets == 0 then break end
    end

    if self.hp < self.max_hp / 2 then
        util.log(self.name .. " takes a hit and immediately regrets everything! RETREAT!")
        world.move_to(self.home)
        return events
    end

    util.log(self.name .. " cackles and disappears before anyone can retaliate")
    world.move_to(self.home)
    try_divine_conception(events)
    return events
end

return do_tick()
