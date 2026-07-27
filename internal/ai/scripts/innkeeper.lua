-- Innkeeper AI
-- Tends the common room, serves drinks to patrons, buys tankards from salesmen,
-- and keeps a small stock of ale, wine, and liquor. Returns to the inn at night.

local DRINKS = {
    "beer", "ale", "wine", "liquor", "mead", "brandy"
}

local DRINK_PRICES = {
    beer = 3, ale = 2, wine = 5, liquor = 8, mead = 4, brandy = 7
}

local RESTOCK_INTERVAL = 90
local TRADE_INTERVAL = 5

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

local function low_stock()
    for _, drink in ipairs(DRINKS) do
        if count_item(drink) < 3 then
            return true
        end
    end
    return false
end

local function restock_drinks()
    if not low_stock() then return false end
    for _, drink in ipairs(DRINKS) do
        if count_item(drink) < 5 then
            world.add_item(drink)
            world.add_item(drink)
            util.log(self.name .. " restocked " .. drink)
        end
    end
    return true
end

local function sell_drinks()
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, target_id in ipairs(nearby) do
        if target_id == self.id then goto continue end
        local info = world.entity_info(target_id)
        if info and info.alive and (info.faction == "civilian" or info.faction == "merchant" or info.faction == "bandit" or info.faction == "thief") then
            for _, drink in ipairs(DRINKS) do
                if has_item(drink) then
                    local result = world.try_sell(target_id, drink)
                    if result and result.done then
                        util.log(self.name .. " served " .. drink .. " to " .. info.name)
                        return true
                    end
                end
            end
        end
        ::continue::
    end
    return false
end

local function buy_tankards()
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, target_id in ipairs(nearby) do
        local info = world.entity_info(target_id)
        if info and info.alive and info.faction == "merchant" then
            local items = world.entity_items(target_id)
            if items then
                for _, item_id in ipairs(items) do
                    if item_id == "tankard" then
                        local result = world.try_buy(target_id, "tankard")
                        if result and result.done then
                            util.log(self.name .. " bought a tankard")
                            return true
                        end
                    end
                end
            end
        end
    end
    return false
end

local function clean_up()
    if world.tick % 50 == 0 then
        util.log(self.name .. " wiped down the bar")
        util.set_mood("happy")
    end
end

local function do_tick()
    local phase = world.phase
    local tick = world.tick

    if phase == "night" or phase == "dusk" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            util.log(self.name .. " closed the inn for the night")
        end
        return
    end

    if tick % RESTOCK_INTERVAL == 0 then
        restock_drinks()
    end

    if tick % TRADE_INTERVAL == 0 then
        sell_drinks()
        buy_tankards()
    end

    if tick % 60 == 0 and self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
    end

    clean_up()
end

do_tick()
