-- Goblin Ambush AI - Stealth Attack
-- Hides before combat for first strike
-- Gets bonus damage from hidden
-- Steals loot on kill
-- Flees when outnumbered

local HIDE_CHANCE = 50
local AMBUSH_BONUS = 1.5
local STEAL_CHANCE = 30
local FLEE_HP_THRESHOLD = 0.5
local ATTACK_CHANCE = 65
local HIDE_DURATION = 10

local function find_hostile()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return nil
    end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and world.is_hostile(self.faction, info.faction) then
            return eid, info
        end
    end
    return nil
end

local function count_allies()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return 0
    end
    local count = 0
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.faction == self.faction then
            count = count + 1
        end
    end
    return count
end

local function is_outnumbered()
    local allies = count_allies()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return false
    end
    local enemies = 0
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and world.is_hostile(self.faction, info.faction) then
            enemies = enemies + 1
        end
    end
    return enemies > allies
end

local function should_flee()
    local hp_ratio = self.hp / self.max_hp
    return hp_ratio < FLEE_HP_THRESHOLD or is_outnumbered()
end

local function do_hide()
    if util.rand_int(100) < HIDE_CHANCE then
        util.mem_set("hidden", world.tick + HIDE_DURATION)
        util.log(self.name .. " hid in the shadows")
        return true
    end
    return false
end

local function do_ambush(target_id)
    local hidden = util.mem_get("hidden")
    local is_hidden = hidden and hidden > world.tick

    if is_hidden then
        util.mem_set("hidden", nil)
        util.log(self.name .. " ambushed " .. target_id .. " from hiding!")
        return true
    end
    return false
end

local function do_steal_loot(target_id)
    if util.rand_int(100) < STEAL_CHANCE then
        local items = world.entity_items(target_id)
        if items and #items > 0 then
            local item = items[util.rand_int(#items) + 1]
            world.add_item(item)
            util.log(self.name .. " looted " .. item .. " from " .. target_id)
            return true
        end
    end
    return false
end

function do_tick()
    local phase = world.phase

    if should_flee() then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            world.move_to(dest)
            util.log(self.name .. " fled from danger")
            return {util.event("flee", {})}
        end
        return {}
    end

    local target_id, target_info = find_hostile()
    if target_id then
        local hidden = util.mem_get("hidden")
        local is_hidden = hidden and hidden > world.tick

        if not is_hidden then
            do_hide()
        end

        if util.rand_int(100) < ATTACK_CHANCE then
            local hit = world.attack(self.id, target_id)
            if hit then
                do_steal_loot(target_id)
            end
        end
        return {util.event("attack", {})}
    end

    if world.tick % 20 == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            if dest ~= self.loc_id then
                world.move_to(dest)
                return {util.event("move", {})}
            end
        end
    end
    return {}
end

return do_tick()
