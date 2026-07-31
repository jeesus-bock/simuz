-- tie_o_mat.lua
-- Tie-O-Mat the Flatulent: primordial dragon goddess of hoarded copper and dungeon odors.
-- A garbled parody of Tiamat.

function do_tick()
    local events = {}

    if world.tick % 240 ~= 0 then
        return events
    end

    local roll = util.rand_int(10)

    if roll < 3 then
        local arguments = {
            {"Left Head", "Right Head", "We should eat the village!"},
            {"Right Head", "Left Head", "No, we should eat the FIELDS!"},
            {"Middle Head", "Both", "Can we please just hoard some copper in peace?"},
            {"Left Head", "Middle Head", "Stop agreeing with Right Head!"},
            {"Right Head", "Left Head", "I'm the dominant head! It says so on the form!"},
            {"Middle Head", "Left Head", "You ate the last villager, it's MY turn!"}
        }
        local arg = arguments[util.rand_int(#arguments) + 1]
        util.log(self.name .. " argues with itself:")
        util.log("  " .. arg[1] .. " to " .. arg[2] .. ": \"" .. arg[3] .. "\"")
        util.log("  The argument escalates. One head tries to bite another. It bites itself instead.")
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { event = "internal_argument" }
        }))
        return events
    end

    local mortal_locs = world.find_mortal_locations()
    if not mortal_locs or #mortal_locs == 0 then return events end
    local dest = mortal_locs[util.rand_int(#mortal_locs) + 1]
    if not world.move_to(dest) then return events end
    util.log(self.name .. " CRASHES into " .. dest .. " trailing an overwhelming stench of sulfur and old cheese!")

    if roll < 6 then
        util.log(self.name .. " lets out a PRIMORDIAL BLAST from all five heads simultaneously!")
        util.log("The smell is indescribable. Trees wilt. Birds fall from the sky.")
        world.damage_location(self.id, 20)
        util.log("Several mortals faint from the odor alone. Others consider it.")
        table.insert(events, util.event("world", {
            source = self.id,
            data = { location = dest, event = "primordial_flatulence" }
        }))
    elseif roll < 8 then
        util.log(self.name .. " begins hoarding everything at " .. dest .. ":")
        local loot = {"copper coins", "rusty spoons", "pebbles", "broken pottery", "old socks", "a single grain of wheat"}
        for i = 1, 3 do
            local item = loot[util.rand_int(#loot) + 1]
            util.log("  ...picks up " .. item .. ". \"TREASURE!\"")
        end
        util.log(self.name .. " sits on the hoard. It is not impressive.")
        table.insert(events, util.event("ambient", {
            source = self.id,
            data = { event = "copper_hoarding" }
        }))
    else
        util.log(self.name .. " spawns CHAOS! All five heads breathe different things at once!")
        util.log("  Head 1: fire (it's a candle flame)")
        util.log("  Head 2: ice (a light frost)")
        util.log("  Head 3: acid (basically lemon juice)")
        util.log("  Head 4: lightning (a static shock)")
        util.log("  Head 5: gas (the usual)")
        world.damage_location(self.id, 15)
        util.log("The combined effect is mildly annoying and very smelly.")
        table.insert(events, util.event("world", {
            source = self.id,
            data = { location = dest, event = "chaos_breath" }
        }))
    end

    util.log(self.name .. " flies away. All five heads are still arguing about directions.")
    world.move_to(self.home)
    return events
end

return do_tick()
