-- Herbalist AI
-- Gathers herbs in the wild or garden during the day, crafts poultices and salves
-- at a campfire or cauldron, sells remedies to priests and wounded civilians,
-- and returns home at night.

local GATHER_INTERVAL = 12
local CRAFT_INTERVAL = 35
local SELL_INTERVAL = 20

local REMEDIES = {
    "herbal_poultice", "healing_salve", "bandage"
}

local RECIPE_ORDER = {
    "cook_poultice", "refine_salve", "craft_bandage"
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

local function gather_herbs()
    world.add_item("herb")
    if util.rand_int(100) < 25 then
        world.add_item("herb")
    end
    util.log(self.name .. " gathered herbs")
end

local function craft_remedies()
    for _, recipe_id in ipairs(RECIPE_ORDER) do
        local ok = world.craft(recipe_id)
        if ok then
            return true
        end
    end
    return false
end

local function find_wounded_buyer()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, target_id in ipairs(nearby) do
        local info = world.entity_info(target_id)
        if info and info.alive and info.hp < info.max_hp and info.faction ~= "undead" and info.faction ~= "hag" then
            return target_id, info
        end
    end
    return nil
end

local function sell_remedies()
    local target_id, info = find_wounded_buyer()
    if target_id then
        for _, remedy in ipairs(REMEDIES) do
            if has_item(remedy) then
                local result = world.try_sell(target_id, remedy)
                if result and result.done then
                    util.log(self.name .. " sold " .. remedy .. " to " .. info.name)
                    return true
                end
            end
        end
    end

    -- Sell to any civilian/merchant if no wounded around
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, other_id in ipairs(nearby) do
        local other_info = world.entity_info(other_id)
        if other_info and other_info.alive and (other_info.faction == "civilian" or other_info.faction == "merchant" or other_info.faction == "priest") then
            for _, remedy in ipairs(REMEDIES) do
                if has_item(remedy) then
                    local result = world.try_sell(other_id, remedy)
                    if result and result.done then
                        util.log(self.name .. " sold " .. remedy .. " to " .. other_info.name)
                        return true
                    end
                end
            end
        end
    end
    return false
end

local function do_tick()
    local phase = world.phase
    local tick = world.tick

    if world.defend_self and world.defend_self() then
        return
    end

    if phase == "night" or phase == "dusk" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            util.log(self.name .. " returned to the hut for the night")
        end
        return
    end

    if tick % GATHER_INTERVAL == 0 then
        gather_herbs()
    end

    if tick % CRAFT_INTERVAL == 0 then
        craft_remedies()
    end

    if tick % SELL_INTERVAL == 0 then
        sell_remedies()
    end

    if tick % 60 == 0 and self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
    end

    if tick % 45 == 0 then
        util.set_mood("relaxed")
    end
end

do_tick()
