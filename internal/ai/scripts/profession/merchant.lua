-- Merchant AI
-- Runs a shop in a settlement, buying and selling goods.
-- Stays near the market during the day, returns home at night.

local WARES_TO_SELL = {
    iron_sword = true,
    short_sword = true,
    iron_axe = true,
    iron_shield = true,
    leather_armor = true,
    leather_helmet = true,
    leather_boots = true,
    leather_gloves = true,
    iron_ore = true,
    whetstone = true,
    bandage = true,
    herb = true,
    beer = true,
    ale = true,
    bread = true,
}

local WARES_TO_BUY = {
    iron_ore = true,
    herb = true,
    leather = true,
    iron = true,
}

local TRADE_INTERVAL = 5
local RESTOCK_INTERVAL = 100

local function restock()
    for item_id, _ in pairs(WARES_TO_SELL) do
        local count = 0
        for _, inv_id in ipairs(self.inventory) do
            if inv_id == item_id then count = count + 1 end
        end
        if count < 3 then
            world.add_item(item_id)
            world.add_item(item_id)
            util.log(self.name .. " restocked " .. item_id)
        end
    end
end

local function do_trade()
    local targets = world.nearby_entities()
    if not targets or #targets == 0 then return end

    for _, target_id in ipairs(targets) do
        if target_id ~= self.id then
            local info = world.entity_info(target_id)
            if info and info.alive and info.faction == "civilian" then
                for item_id, _ in pairs(WARES_TO_BUY) do
                    local target_items = world.entity_items(target_id)
                    if target_items then
                        for _, tid in ipairs(target_items) do
                            if tid == item_id then
                                local result = world.try_buy(target_id, item_id)
                                if result and result.done then
                                    util.log(self.name .. " bought " .. item_id .. " from " .. info.name)
                                    return
                                end
                            end
                        end
                    end
                end

                for _, item_id in ipairs(self.inventory) do
                    if WARES_TO_SELL[item_id] then
                        local result = world.try_sell(target_id, item_id)
                        if result and result.done then
                            util.log(self.name .. " sold " .. item_id .. " to " .. info.name)
                            return
                        end
                    end
                end
            end
        end
    end
end

function do_tick()
    local tick = world.tick
    local phase = world.phase

    if world.defend_self and world.defend_self() then
        return {util.event("profession_action", {profession = "merchant"})}
    end
    if world.avoid_combat and world.avoid_combat() then
        return {util.event("profession_action", {profession = "merchant"})}
    end

    if phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
        end
        return {}
    end

    if tick % RESTOCK_INTERVAL == 0 then
        restock()
    end

    if tick % TRADE_INTERVAL == 0 then
        do_trade()
    end

    return {}
end

return do_tick()