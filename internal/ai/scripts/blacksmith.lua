-- Blacksmith AI
-- Works at the forge during the day, crafts arms and armour from raw materials,
-- sells finished goods to nearby civilians and guards, and buys ore/coal from miners.
-- Returns to the smithy at night.

local CRAFT_INTERVAL = 30
local TRADE_INTERVAL = 10
local RESTOCK_INTERVAL = 120

local WARES = {
    "iron_ingot", "iron_sword", "iron_axe", "iron_spear", "dagger",
    "short_sword", "leather_armor", "chainmail", "iron_helmet",
    "iron_boots", "leather_boots", "wooden_shield", "iron_shield", "bandage"
}

local BUY_MATERIALS = {
    "iron_ore", "coal", "cloth", "leather_strips", "leather"
}

local RECIPE_ORDER = {
    "smelt_iron", "craft_bandage"
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

local function is_at_workshop()
    local loc = self.loc_id
    local info = world.recipe_info("smelt_iron")
    if info and info.station then
        -- We can only craft if the current location tag matches the station;
        -- we approximate this by checking whether the recipe would succeed.
        return world.has_material("iron_ore") or world.has_material("coal")
    end
    return false
end

local function craft_goods()
    for _, recipe_id in ipairs(RECIPE_ORDER) do
        local info = world.recipe_info(recipe_id)
        if info then
            local ok = world.craft(recipe_id)
            if ok then
                return true
            end
        end
    end
    return false
end

local function try_buy_materials()
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, target_id in ipairs(nearby) do
        local info = world.entity_info(target_id)
        if info and info.alive then
            local items = world.entity_items(target_id)
            if items then
                for _, item_id in ipairs(items) do
                    for _, mat in ipairs(BUY_MATERIALS) do
                        if item_id == mat then
                            if not has_item(mat) or count_item(mat) < 5 then
                                local result = world.try_buy(target_id, mat)
                                if result and result.done then
                                    util.log(self.name .. " bought " .. mat .. " from " .. info.name)
                                    return true
                                end
                            end
                        end
                    end
                end
            end
        end
    end
    return false
end

local function sell_wares()
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, target_id in ipairs(nearby) do
        local info = world.entity_info(target_id)
        if info and info.alive and (info.faction == "civilian" or info.faction == "guard" or info.faction == "merchant") then
            for _, ware in ipairs(WARES) do
                if has_item(ware) then
                    local result = world.try_sell(target_id, ware)
                    if result and result.done then
                        util.log(self.name .. " sold " .. ware .. " to " .. info.name)
                        return true
                    end
                end
            end
        end
    end
    return false
end

local function restock_materials()
    -- If we have no raw materials at all, quietly "restock" a small amount
    -- so the smithy can keep producing even without a miner supplier.
    local has_materials = false
    for _, mat in ipairs(BUY_MATERIALS) do
        if has_item(mat) then
            has_materials = true
            break
        end
    end
    if not has_materials then
        world.add_item("iron_ore")
        world.add_item("iron_ore")
        world.add_item("coal")
        util.log(self.name .. " restocked raw materials")
    end
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
            util.log(self.name .. " returned to the smithy")
        end
        return
    end

    if tick % RESTOCK_INTERVAL == 0 then
        restock_materials()
    end

    if tick % TRADE_INTERVAL == 0 then
        try_buy_materials()
        sell_wares()
    end

    if tick % CRAFT_INTERVAL == 0 then
        craft_goods()
    end

    if tick % 60 == 0 and self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
    end

    if tick % 40 == 0 then
        local roll = util.rand_int(100)
        if roll < 20 then
            util.set_mood("focused")
        elseif roll < 40 then
            util.set_mood("neutral")
        end
    end
end

do_tick()
