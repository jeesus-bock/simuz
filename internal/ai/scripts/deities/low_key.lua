-- low_key.lua
-- Low-Key the Fumbler: god of bad advice and spilled grease.
-- A garbled parody of Loki.

function do_tick()
    local events = {}

    if world.tick % 100 ~= 0 then
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
        local forms = {"rat", "horse", "spider", "dog", "goblin", "orc", "elf", "chicken"}
        local form = forms[util.rand_int(#forms) + 1]
        util.log(self.name .. " shape-shifts into a " .. form .. "! ...wait, something went wrong.")
        world.polymorph(self.id, form)
        util.log(self.name .. " is stuck as a " .. form .. " and knocking things over!")
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { event = "botched_polymorph", form = form }
        }))
        world.revert_polymorph(self.id)
        util.log(self.name .. " reverts, but one ear is still a " .. form .. "'s ear. Close enough.")
    elseif roll < 5 then
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive then
                        util.log(self.name .. " steals from " .. info.name .. " while pretending to help them look for their keys")
                        world.steal(self.id, eid)
                        table.insert(events, util.event("combat", {
                            source = self.id,
                            target = eid,
                            data = { event = "larceny" }
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
                    if info and info.alive then
                        local bad_advice = {
                            "You should definitely fight that bear.",
                            "Invest all your gold in magic beans.",
                            "Trust me, the bridge is totally safe.",
                            "Have you tried turning yourself in to the guard? They love honesty.",
                            "That mushroom is definitely not poisonous."
                        }
                        local advice = bad_advice[util.rand_int(#bad_advice) + 1]
                        util.log(self.name .. " whispers to " .. info.name .. ": \"" .. advice .. "\"")
                        util.set_mood("confused", 30)
                        table.insert(events, util.event("mood", {
                            source = self.id,
                            target = eid,
                            data = { mood = "confused", event = "bad_advice" }
                        }))
                        break
                    end
                end
            end
        end
    elseif roll < 9 then
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive and info.faction and info.faction ~= "" then
                        util.log(self.name .. " tells " .. info.name .. " that their best friend said something terrible about them")
                        util.set_mood("angry", 40)
                        table.insert(events, util.event("mood", {
                            source = self.id,
                            target = eid,
                            data = { mood = "angry", event = "betrayal_lie" }
                        }))
                        break
                    end
                end
            end
        end
    else
        util.log(self.name .. " spills a bucket of grease on the floor at " .. dest)
        util.log(self.name .. " watches from a corner, giggling, as everyone slips around")
        world.damage_location(self.id, 5)
        table.insert(events, util.event("world", {
            source = self.id,
            data = { location = dest, event = "grease_trap" }
        }))
    end

    util.log(self.name .. " disappears in a puff of suspicious-smelling smoke")
    world.move_to(self.home)
    return events
end

return do_tick()
