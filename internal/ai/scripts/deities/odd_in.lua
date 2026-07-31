-- odd_in.lua
-- Odd-In the Near-Sighted: god of conspiracy theories and damp scrolls.
-- A garbled parody of Odin.

function do_tick()
    local events = {}

    if world.tick % 200 ~= 0 then
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        return events
    end

    local roll = util.rand_int(10)

    if roll < 3 then
        util.log(self.name .. " dispatches his two ravens, Bluggin and Dullinn, to scout the mortal realm")
        local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
        local nearby = world.entities_at(dest)
        if nearby then
            for _, eid in ipairs(nearby) do
                local info = world.entity_info(eid)
                if info and info.alive then
                    util.log("Bluggin reports: " .. info.name .. " is " .. info.species .. ", level " .. info.level .. ". Suspicious.")
                end
            end
        end
        util.log("Dullinn reports: found a damp scroll. It's unreadable. Classic.")
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { event = "raven_scouting" }
        }))
    elseif roll < 6 then
        local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
        if not world.move_to(dest) then return events end
        util.log(self.name .. " arrives at " .. dest .. " peering through a monocle at everyone")

        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and info.alive then
                        util.log(self.name .. " offers " .. info.name .. " forbidden knowledge in exchange for... wait, what was the deal again?")
                        util.set_mood("wise", 20)
                        break
                    end
                end
            end
        end
        world.move_to(self.home)
    elseif roll < 8 then
        util.log(self.name .. " sacrifices 20 HP hanging from the Cosmic Coat Rack for wisdom!")
        if self.hp > 20 then
            world.damage_location(self.id, 0)
            util.set_mood("wise", 60)
            util.log(self.name .. " gains... a headache. But a WISE headache.")
            table.insert(events, util.event("mood", {
                source = self.id,
                data = { mood = "wise", event = "self_sacrifice" }
            }))
        else
            util.log(self.name .. " is too weak to hang from anything right now. Lies down instead.")
        end
    else
        util.log(self.name .. " goes to collect the honored dead")
        local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
        if not world.move_to(dest) then return events end
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                if eid ~= self.id then
                    local info = world.entity_info(eid)
                    if info and not info.alive then
                        util.log(self.name .. " inspects the corpse of " .. info.name .. ". \"You'll do for my collection.\"")
                        table.insert(events, util.event("ambient", {
                            source = self.id,
                            target = eid,
                            data = { event = "dead_collection" }
                        }))
                    end
                end
            end
        end
        world.move_to(self.home)
    end

    util.log(self.name .. " squints at the horizon and wanders home")
    return events
end

return do_tick()
