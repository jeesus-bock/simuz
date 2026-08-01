-- ooh_huang.lua
-- Ooh-Huang the Clerk: supreme ruler of divine red tape and stagnation.
-- A garbled parody of the Jade Emperor.

local function try_divine_conception(events)
    if world.tick % 300 ~= 150 then return end
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

    if world.tick % 300 ~= 0 then
        return events
    end

    local roll = util.rand_int(10)

    if roll < 5 then
        local forms = {
            "stamps a divine form 27-B in triplicate",
            "reviews a complaint about celestial plumbing (denied)",
            "reorganizes the cloud filing cabinet for the 9000th time",
            "drafts a strongly-worded memo about divine overtime policies",
            "signs a decree requiring all mortals to use proper punctuation in prayers",
            "processes a permit request for a new constellation (pending review since the last age)",
            "denies a weather modification request on a technicality"
        }
        local action = forms[util.rand_int(#forms) + 1]
        util.log(self.name .. " " .. action)
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { event = "divine_bureaucracy" }
        }))
        return events
    end

    if roll < 8 then
        local mortal_locs = world.find_mortal_locations()
        if not mortal_locs or #mortal_locs == 0 then return events end
        local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
        if not world.move_to(dest) then return events end
        util.log(self.name .. " descends to " .. dest .. " to conduct a surprise audit")

        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive then
                        util.log(self.name .. " inspects " .. info.name .. ": Level " .. info.level .. ", HP " .. info.hp .. "/" .. info.max_hp .. ". Scribbles notes.")
                        if info.hp > info.max_hp / 2 then
                            util.log(self.name .. " issues " .. info.name .. " a citation for being too healthy. \"Must be hiding something.\"")
                            util.set_mood("annoyed", 20)
                        else
                            util.log(self.name .. " issues " .. info.name .. " a citation for being too unhealthy. \"Not a good look for the realm.\"")
                        end
                    end
                end
            end
        end

        util.log(self.name .. " files the audit report and returns to heaven. Another productive day of doing nothing.")
        world.move_to(self.home)
    else
        util.log(self.name .. " issues a divine decree: ALL MORTALS SHALL HENCEFORTH QUEUE IN AN ORDERLY FASHION")
        util.log(self.name .. " nobody hears this. The decree is posted on a cloud. It is already damp.")
        table.insert(events, util.event("world", {
            source = self.id,
            data = { event = "divine_decree", decree = "orderly_queuing" }
        }))
    end

    try_divine_conception(events)
    return events
end

return do_tick()
