-- Traveling Salesman AI
-- Wanders between towns, trades goods, haggles with creatures
-- Sleeps at inns at night, trades during day

local route = {
    "frosthold_inn_common",
    "frosthold_blacksmith",
    "frosthold_guardhouse",
    "stillwater_inn_common",
    "stillwater_blacksmith",
    "stillwater_market",
    "golden_gate_inn_common",
    "golden_gate_market",
    "golden_gate_guardhouse",
}

local current_index = 1
for i, v in ipairs(route) do
    if v == self.loc_id then
        current_index = i
        break
    end
end

local function pick_next_destination()
    if #route == 0 then
        return nil
    end
    local next = (current_index % #route) + 1
    return route[next]
end

local wares_to_sell = {
    fine_clothes = true,
    simple_robe = true,
    work_tunic = true,
    dagger = true,
    short_sword = true,
    cudgel = true,
    leather_helmet = true,
    leather_boots = true,
    leather_gloves = true,
    wooden_shield = true,
}

local wares_to_buy = {
    tankard = true,
    holy_symbol = true,
}

local traded_this_tick = {}

local function try_trade(target_id)
    if traded_this_tick[target_id] then
        return
    end

    local target_items = world.entity_items(target_id)
    if target_items then
        for _, item_id in ipairs(target_items) do
            if wares_to_buy[item_id] then
                local has = false
                for _, inv_id in ipairs(self.inventory) do
                    if inv_id == item_id then has = true break end
                end
                if not has then
                    local result = world.try_buy(target_id, item_id)
                    if result and result.done then
                        util.log("bought " .. item_id .. " from " .. target_id .. " for " .. result.price)
                        traded_this_tick[target_id] = true
                        return
                    end
                end
            end
        end
    end

    for _, item_id in ipairs(self.inventory) do
        if wares_to_sell[item_id] then
            local result = world.try_sell(target_id, item_id)
            if result and result.done then
                util.log("sold " .. item_id .. " to " .. target_id .. " for " .. result.price)
                traded_this_tick[target_id] = true
                return
            end
        end
    end
end

local function do_trade()
    if #self.inventory == 0 then
        return
    end
    local targets = world.nearby_entities()
    if not targets or #targets == 0 then
        return
    end
    traded_this_tick = {}
    for _, target_id in ipairs(targets) do
        if target_id ~= self.name then
            try_trade(target_id)
        end
    end
end

local function do_tick()
    local tick = world.tick
    local phase = world.phase

    if world.is_traveling() then
        return
    end

    -- Sync index if we landed on a route stop (after travel or load)
    for i, v in ipairs(route) do
        if v == self.loc_id then
            current_index = i
            break
        end
    end

    if phase == "night" then
        return
    end

    if tick % 5 == 0 and phase ~= "dawn" and phase ~= "dusk" then
        do_trade()
    end

    if tick % 10 == 0 then
        local dest = pick_next_destination()
        if dest and dest ~= self.loc_id then
            local ok = world.move_to(dest)
            if ok then
                current_index = (current_index % #route) + 1
            end
        end
    end
end

do_tick()
