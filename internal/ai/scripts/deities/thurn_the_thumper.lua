-- thurn_the_thumper.lua
-- Thurn the Thumper: god of loud noises and shattered handles.
-- A garbled parody of Thor.

function do_tick()
    local events = {}

    if world.tick % 160 ~= 0 then
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        return events
    end

    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return events end
    util.log(self.name .. " CRASHES down from the heavens at " .. dest .. "! The ground cracks!")

    local nearby = world.nearby_entities()
    if not nearby then
        util.log(self.name .. " smashes the ground anyway. Nothing to hit. Hammer handle snaps off. Typical.")
        world.damage_location(self.id, 10)
        world.move_to(self.home)
        return events
    end

    local roll = util.rand_int(10)

    if roll < 5 then
        local hit_count = 0
        for _, eid in ipairs(nearby) do
            if eid ~= self.id then
                local info = world.entity_info(eid)
                if info and info.alive then
                    if info.faction and info.faction ~= "civilian" then
                        util.log(self.name .. " brings MJOLNIR down on " .. info.name .. "! THUNDEROUS SMASH!")
                        world.attack(self.id, eid)
                        hit_count = hit_count + 1
                        table.insert(events, util.event("combat", {
                            source = self.id,
                            target = eid,
                            data = { event = "mjolnir_smash" }
                        }))
                    end
                end
            end
        end
        if hit_count == 0 then
            util.log(self.name .. " finds no enemies! Smashes a table in frustration!")
            world.damage_location(self.id, 5)
        end
    elseif roll < 8 then
        util.log(self.name .. " summons a THUNDERSTORM! Lightning crashes everywhere!")
        world.damage_location(self.id, 25)
        util.log("Everything at " .. dest .. " is now very crispy and smells of ozone.")
        table.insert(events, util.event("world", {
            source = self.id,
            data = { location = dest, event = "thunderstorm" }
        }))
    else
        util.log(self.name .. " looks around for worshippers to protect")
        for _, eid in ipairs(nearby) do
            if eid ~= self.id then
                local info = world.entity_info(eid)
                if info and info.alive and info.worship and info.worship > 15 then
                    if info.hp < info.max_hp / 2 then
                        util.log(self.name .. " shields " .. info.name .. " with his mighty frame! \"YOU ARE SAFE NOW, TINY MORTAL!\"")
                        world.heal(self.id, eid, 20)
                        table.insert(events, util.event("combat", {
                            source = self.id,
                            target = eid,
                            data = { event = "divine_protection" }
                        }))
                    end
                end
            end
        end
    end

    util.log(self.name .. " punches a cloud on the way home. The cloud loses.")
    world.move_to(self.home)
    return events
end

return do_tick()
