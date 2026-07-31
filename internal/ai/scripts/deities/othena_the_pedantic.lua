-- othena_the_pedantic.lua
-- Othena the Pedantic: goddess of grammatical heresy and unwinnable debates.
-- A garbled parody of Athena.

function do_tick()
    local events = {}

    if world.tick % 180 ~= 0 then
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        return events
    end

    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return events end
    util.log(self.name .. " descends to " .. dest .. ", scroll of corrections in hand")

    local nearby = world.nearby_entities()
    if not nearby then
        world.move_to(self.home)
        return events
    end

    local roll = util.rand_int(10)

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                if roll < 4 then
                    util.log(self.name .. " corrects " .. info.name .. "'s grammar: \"It's 'whom,' you absolute walnut!\"")
                    util.set_mood("annoyed", 30)
                    table.insert(events, util.event("mood", {
                        source = self.id,
                        target = eid,
                        data = { mood = "annoyed", event = "grammar_correction" }
                    }))
                elseif roll < 7 then
                    if info.worship and info.worship > 20 then
                        util.log(self.name .. " grants " .. info.name .. " a moment of dubious wisdom")
                        util.set_mood("inspired", 50)
                        table.insert(events, util.event("mood", {
                            source = self.id,
                            target = eid,
                            data = { mood = "inspired", event = "divine_wisdom" }
                        }))
                    else
                        util.log(self.name .. " refuses to enlighten " .. info.name .. ". \"You haven't earned it, plebeian.\"")
                    end
                elseif roll < 9 then
                    if not info.worship or info.worship < 5 then
                        util.log(self.name .. " declares " .. info.name .. " guilty of intellectual hubris! TRANSFORMATION!")
                        world.polymorph(eid, "spider")
                        util.log(info.name .. " is now a spider. \"Perhaps eight legs will help you think better.\"")
                        table.insert(events, util.event("world", {
                            source = self.id,
                            target = eid,
                            data = { event = "arachne_punishment" }
                        }))
                    end
                else
                    util.log(self.name .. " launches into an unwinnable debate about the nature of virtue with " .. info.name)
                    util.set_mood("confused", 40)
                    table.insert(events, util.event("mood", {
                        source = self.id,
                        target = eid,
                        data = { mood = "confused", event = "philosophical_debate" }
                    }))
                end
            end
        end
    end

    util.log(self.name .. " returns to the divine realm, muttering about the decline of rhetoric")
    world.move_to(self.home)
    return events
end

return do_tick()
