-- Raiding AI
-- Seeks out hostile locations and loots them.
-- Runs after aggressive.lua (which handles immediate combat).
-- Focuses on raiding behavior: moving toward populated areas, looting kills.

local RAID_INTERVAL = 15
local LOOT_CHANCE = 40

local function find_lootable()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return nil
    end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and not info.alive then
            local items = world.entity_items(eid)
            if items and #items > 0 then
                return eid, info
            end
        end
    end
    return nil
end

local function do_loot(target_id)
    if util.rand_int(100) >= LOOT_CHANCE then
        return false
    end
    local items = world.entity_items(target_id)
    if not items or #items == 0 then
        return false
    end
    local item = items[util.rand_int(#items) + 1]
    world.add_item(item)
    util.log(self.name .. " looted " .. item .. " from the carnage")
    return true
end

local function find_raid_target()
    local exits = world.exits_from(self.loc_id)
    if not exits or #exits == 0 then
        return nil
    end
    for _, dest in ipairs(exits) do
        local entities = world.entities_at(dest)
        if entities and #entities > 0 then
            return dest
        end
    end
    return nil
end

local function do_tick()
    local tick = world.tick

    if tick % RAID_INTERVAL ~= 0 then
        return false
    end

    local corpse_id, corpse_info = find_lootable()
    if corpse_id then
        do_loot(corpse_id)
        return true
    end

    local raid_dest = find_raid_target()
    if raid_dest then
        world.move_to(raid_dest)
        util.log(self.name .. " raided toward " .. raid_dest)
        return true
    end

    if tick % 30 == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            if dest ~= self.loc_id then
                world.move_to(dest)
                return true
            end
        end
    end
    return false
end

return do_tick()
