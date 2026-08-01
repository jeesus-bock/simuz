-- groan_yin.lua
-- Groan-Yin the Sighing: goddess of passive-aggressive pity.
-- A garbled parody of Guanyin.

local function try_divine_conception(events)
    if world.tick % 190 ~= 95 then return end
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

    if world.tick % 190 ~= 0 then
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        util.log(self.name .. " sighs. No mortals to pity. The silence is its own punishment.")
        return events
    end

    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return events end
    util.log(self.name .. " manifests at " .. dest .. " with a thousand arms. All of them are sighing.")

    local nearby = world.nearby_entities()
    if not nearby then
        util.log(self.name .. " sighs at the emptiness. \"Fine. Nobody needs me. That's fine.\"")
        world.move_to(self.home)
        return events
    end

    local roll = util.rand_int(10)
    local helped_anyone = false

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                if roll < 5 then
                    if info.hp < info.max_hp then
                        local heal_amt = math.min(25, info.max_hp - info.hp)
                        world.heal(self.id, eid, heal_amt)
                        util.log(self.name .. " heals " .. info.name .. " with " .. heal_amt .. " HP. \"You're welcome. Not that anyone asked.\"")
                        util.set_mood("pitied", 30)
                        helped_anyone = true
                        table.insert(events, util.event("combat", {
                            source = self.id,
                            target = eid,
                            data = { event = "passive_aggressive_healing", amount = heal_amt }
                        }))
                    else
                        util.log(self.name .. " looks at " .. info.name .. ". \"Oh, you're FINE. You don't even need me. Nobody does.\"")
                    end
                elseif roll < 8 then
                    if info.hp < info.max_hp / 2 then
                        local heal_amt = math.min(40, info.max_hp - info.hp)
                        world.heal(self.id, eid, heal_amt)
                        util.log(self.name .. " dramatically heals " .. info.name .. " with all thousand arms at once. \"I GUESS I'll do EVERYTHING myself, as usual.\"")
                        helped_anyone = true
                        table.insert(events, util.event("combat", {
                            source = self.id,
                            target = eid,
                            data = { event = "dramatic_healing", amount = heal_amt }
                        }))
                    else
                        util.log(self.name .. " pats " .. info.name .. " on the head. \"There, there. You're doing your best. Which isn't great, but still.\"")
                        util.set_mood("pitied", 20)
                    end
                else
                    util.log(self.name .. " sighs at " .. info.name .. ". \"I COULD help you, but would it even matter?\"")
                    util.log(self.name .. " stares into the middle distance. A single tear rolls down cheek #347.")
                end
            end
        end
    end

    if not helped_anyone then
        util.log(self.name .. " didn't heal anyone. \"No one ever appreciates what I do. Typical.\"")
    end

    util.log(self.name .. " dissipates into a cloud of resigned compassion")
    world.move_to(self.home)
    try_divine_conception(events)
    return events
end

return do_tick()
