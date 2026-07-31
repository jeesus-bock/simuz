-- cultist.lua
-- Nocturnal heretical script processing ritual chants and hunting infidels.

local function perform_black_mass()
    util.set_mood("focused", 30)

    local devotion = util.mem_get("dark_devotion") or 0
    devotion = devotion + 1
    util.mem_set("dark_devotion", devotion)
    util.log(self.name .. " chants dark vows in the shadows. Devotion: " .. devotion)

    if devotion % 5 == 0 then
        world.add_item(self.id, "ritual_energy", 1)
        util.log(self.name .. " condensed a drop of raw unholy energy.")
    end
end

local function scan_for_sacrifices()
    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            -- Aggressively target priests, holy paladins, or rival gods (like Zeus)
            if info and info.alive and (info.profession == "priest" or info.faction == "worshippers_of_zeus") then
                util.log(self.name .. " shrieks: 'A sacrifice for the void!' Attacking " .. info.name)
                util.set_mood("furious", 40)
                world.attack(self.id, eid)
                break
            end
        end
    end
end

local function do_tick()
    -- Cultists sleep during the day and run their operations in the night phase
    if world.phase == "day" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            return true
        end
        return false
    end

    if world.tick % 10 == 0 then
        scan_for_sacrifices()
    end

    if world.tick % 20 == 0 then
        perform_black_mass()
    end

    return false
end

return do_tick()
