-- groan_yin.lua
-- Groan-Yin the Sighing: goddess of passive-aggressive pity.
-- A garbled parody of Guanyin.

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
    return events
end

return do_tick()
