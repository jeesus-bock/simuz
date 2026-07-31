-- froyda_the_thistle.lua
-- Froyda the Thistle-Queen: goddess of obsessive infatuation and stinging betrayals.
-- A garbled parody of Freya.

function do_tick()
    local events = {}

    if world.tick % 170 ~= 0 then
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then
        return events
    end

    local roll = util.rand_int(10)

    if roll < 2 then
        util.log(self.name .. " weeps golden tears onto the divine floor. They tarnish immediately. Typical.")
        util.set_mood("melancholy", 50)
        table.insert(events, util.event("mood", {
            source = self.id,
            data = { mood = "melancholy", event = "golden_tears" }
        }))
        return events
    end

    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return events end
    util.log(self.name .. " arrives at " .. dest .. " trailing thistle petals and an overwhelming perfume")

    local nearby = world.nearby_entities()
    if not nearby then
        util.log(self.name .. " finds nobody to adore her. Flips a table.")
        world.move_to(self.home)
        return events
    end

    if roll < 5 then
        local targets = {}
        for _, eid in ipairs(nearby) do
            if eid ~= self.id then
                local info = world.entity_info(eid)
                if info and info.alive then
                    table.insert(targets, {id = eid, info = info})
                end
            end
        end
        if #targets > 0 then
            local t = targets[util.rand_int(#targets) + 1]
            util.log(self.name .. " bestows obsessive infatuation upon " .. t.info.name .. ". They will think of nothing else for weeks.")
            util.set_mood("infatuated", 80)
            table.insert(events, util.event("mood", {
                source = self.id,
                target = t.id,
                data = { mood = "infatuated", event = "love_curse" }
            }))
        end
    elseif roll < 7 then
        local has_dead = false
        for _, eid in ipairs(nearby) do
            if eid ~= self.id then
                local info = world.entity_info(eid)
                if info and not info.alive then
                    util.log(self.name .. " collects the soul of " .. info.name .. " for her personal army of the dead. Gorgeous.")
                    has_dead = true
                    table.insert(events, util.event("ambient", {
                        source = self.id,
                        target = eid,
                        data = { event = "soul_collection" }
                    }))
                end
            end
        end
        if not has_dead then
            util.log(self.name .. " finds no dead to collect. Disappointing.")
        end
    else
        local targets = {}
        for _, eid in ipairs(nearby) do
            if eid ~= self.id then
                local info = world.entity_info(eid)
                if info and info.alive then
                    table.insert(targets, {id = eid, info = info})
                end
            end
        end
        if #targets > 0 then
            local t = targets[util.rand_int(#targets) + 1]
            if t.info.worship and t.info.worship < 5 then
                util.log(self.name .. " stings " .. t.info.name .. " with a thistle whip for ignoring her! \"LOOK AT ME!\"")
                world.attack(self.id, t.id)
                table.insert(events, util.event("combat", {
                    source = self.id,
                    target = t.id,
                    data = { event = "thistle_sting" }
                }))
            else
                util.log(self.name .. " blows a kiss at " .. t.info.name .. ". They feel vaguely itchy.")
                util.set_mood("infatuated", 30)
            end
        end
    end

    util.log(self.name .. " vanishes in a cloud of purple thistle pollen")
    world.move_to(self.home)
    return events
end

return do_tick()
