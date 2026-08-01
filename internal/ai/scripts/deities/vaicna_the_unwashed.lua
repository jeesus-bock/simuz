-- vaicna_the_unwashed.lua
-- Vaicna the Unwashed: lich of petty secrets and mildew scrolls.
-- A garbled parody of Vecna.

local function try_divine_conception(events)
    if world.tick % 230 ~= 115 then return end
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

    if world.tick % 230 ~= 0 then
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        return events
    end

    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return events end
    util.log(self.name .. " manifests at " .. dest .. " trailing cobwebs and the smell of wet parchment")

    local roll = util.rand_int(10)

    if roll < 3 then
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive then
                        local secrets = {
                            "I know what you did with that missing goat.",
                            "Your grandmother's secret recipe is terrible. I've read it.",
                            "You owe three coppers to someone you've already forgotten.",
                            "The shadow behind you has been following you for six days.",
                            "I have seen your future. It involves a lot of tripping."
                        }
                        local secret = secrets[util.rand_int(#secrets) + 1]
                        util.log(self.name .. " whispers to " .. info.name .. ": \"" .. secret .. "\"")
                        util.set_mood("paranoid", 40)
                        table.insert(events, util.event("mood", {
                            source = self.id,
                            target = eid,
                            data = { mood = "paranoid", event = "petty_secret" }
                        }))
                        break
                    end
                end
            end
        end
    elseif roll < 6 then
        util.log(self.name .. " pulls a moldy scroll from inside his ribcage and reads it aloud")
        util.log("The scroll crumbles to dust halfway through. \"Typical. This one was ACTUALLY important.\"")
        util.log(self.name .. " searches for another scroll. Finds only lint and a small beetle.")
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { event = "mildew_scroll" }
        }))
    elseif roll < 8 then
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive and info.species ~= "divine" then
                        util.log(self.name .. " offers " .. info.name .. " forbidden knowledge in exchange for... their left sock.")
                        util.log(info.name .. " is deeply confused but mildly tempted.")
                        table.insert(events, util.event("mood", {
                            source = self.id,
                            target = eid,
                            data = { event = "knowledge_trade" }
                        }))
                        break
                    end
                end
            end
        end
    else
        util.log(self.name .. " attempts to cast a terrifying curse at " .. dest)
        util.log("Nothing happens. " .. self.name .. " coughs up dust. \"The spell needs more mildew.\"")
        world.damage_location(self.id, 3)
        table.insert(events, util.event("world", {
            source = self.id,
            data = { location = dest, event = "failed_curse" }
        }))
    end

    util.log(self.name .. " dissolves into a cloud of spores and drifting parchment dust")
    world.move_to(self.home)
    try_divine_conception(events)
    return events
end

return do_tick()
