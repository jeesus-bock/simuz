-- Thief AI
-- Hides in the shadows, pickpockets civilians and merchants, sells stolen goods,
-- and flees when threatened or when guards are near.

local STEAL_CHANCE = 40
local FLEE_HP_THRESHOLD = 0.4
local SELL_INTERVAL = 25

local VALUABLES = {
    "tankard", "holy_symbol", "leather_helmet", "leather_boots", "leather_gloves",
    "dagger", "short_sword", "common_clothes", "fine_clothes", "simple_robe", "work_tunic",
    "cp", "sp", "gp"
}

local function count_item(def_id)
    local n = 0
    for _, id in ipairs(self.inventory) do
        if id == def_id then n = n + 1 end
    end
    return n
end

local function has_item(def_id)
    return count_item(def_id) > 0
end

local function find_victim()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and info.faction ~= self.faction and info.faction ~= "deity" then
            if info.faction == "civilian" or info.faction == "merchant" or info.faction == "guard" or info.faction == "ranger" or info.faction == "bard" then
                return eid, info
            end
        end
        ::continue::
    end
    return nil
end

local function find_guard()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.faction == "guard" then
            return eid, info
        end
    end
    return nil
end

local function pick_best_item(items)
    if not items or #items == 0 then return nil end
    for _, valuable in ipairs(VALUABLES) do
        for _, item_id in ipairs(items) do
            if item_id == valuable then
                return item_id
            end
        end
    end
    return items[util.rand_int(#items) + 1]
end

local function do_steal()
    if util.rand_int(100) >= STEAL_CHANCE then return false end
    local victim_id, victim_info = find_victim()
    if not victim_id then return false end
    local items = world.entity_items(victim_id)
    local target_item = pick_best_item(items)
    if not target_item then return false end
    local result = world.steal(victim_id, target_item)
    if result and result.done then
        util.set_mood("happy")
        return true
    end
    return false
end

local function sell_stolen_goods()
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, target_id in ipairs(nearby) do
        local info = world.entity_info(target_id)
        if info and info.alive and (info.faction == "merchant" or info.faction == "civilian" or info.faction == "bandit") then
            for _, valuable in ipairs(VALUABLES) do
                if has_item(valuable) then
                    local result = world.try_sell(target_id, valuable)
                    if result and result.done then
                        util.log(self.name .. " fenced " .. valuable .. " to " .. info.name)
                        return true
                    end
                end
            end
        end
    end
    return false
end

local function should_flee()
    local hp_ratio = self.hp / self.max_hp
    if hp_ratio < FLEE_HP_THRESHOLD then
        return true
    end
    local guard_id = find_guard()
    if guard_id then
        return true
    end
    return false
end

local function flee()
    local exits = world.exits_from(self.loc_id)
    if exits and #exits > 0 then
        local dest = exits[util.rand_int(#exits) + 1]
        if dest ~= self.loc_id then
            world.move_to(dest)
            util.log(self.name .. " slipped away into the shadows")
        end
    end
end

local function do_tick()
    local tick = world.tick

    if should_flee() then
        flee()
        util.set_mood("stressed")
        return
    end

    if tick % 10 == 0 then
        do_steal()
    end

    if tick % SELL_INTERVAL == 0 then
        sell_stolen_goods()
    end

    if tick % 50 == 0 and util.rand_int(100) < 30 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            if dest ~= self.loc_id then
                world.move_to(dest)
            end
        end
    end

    if tick % 60 == 0 then
        local roll = util.rand_int(100)
        if roll < 40 then
            util.set_mood("neutral")
        end
    end
end

do_tick()
