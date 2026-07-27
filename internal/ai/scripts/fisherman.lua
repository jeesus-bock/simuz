-- Fisherman AI
-- Lives by the pond at Stillwater, fishes during the day,
-- sells catch at the market, buys bait, sleeps at night.

local FISH_TYPES = {
    "fish_trout",
    "fish_salmon",
    "fish_catfish",
}

local FISH_CHANCE = 15
local MIN_FISH_TO_SELL = 3
local BAIT_PRICE = 5
local DODGE_TICKS = 3

local traded_this_tick = {}
local fish_caught_total = 0

local function is_at_pond()
    return self.loc_id == "stillwater_pond"
end

local function is_at_market()
    return self.loc_id == "stillwater_market" or self.loc_id == "stillwater_temple"
end

local function num_fish_in_inventory()
    local count = 0
    for _, item_id in ipairs(self.inventory) do
        for _, fish_id in ipairs(FISH_TYPES) do
            if item_id == fish_id then
                count = count + 1
                break
            end
        end
    end
    return count
end

local function sell_all_fish(target_id)
    if traded_this_tick[target_id] then
        return
    end

    local target_items = world.entity_items(target_id)
    if target_items then
        for _, target_item in ipairs(target_items) do
            local is_buyable_fish = false
            for _, fish_id in ipairs(FISH_TYPES) do
                if target_item == fish_id then
                    is_buyable_fish = true
                    break
                end
            end
            if not is_buyable_fish then
                goto continue
            end

            for _, my_item in ipairs(self.inventory) do
                if my_item == target_item then
                    local result = world.try_sell(target_id, target_item)
                    if result and result.done then
                        util.log("sold " .. target_item .. " to " .. target_id .. " for " .. result.price)
                        traded_this_tick[target_id] = true
                        return
                    end
                end
            end
            ::continue::
        end
    end
end

local function buy_bait(target_id)
    if traded_this_tick[target_id] then
        return
    end

    local result = world.try_buy(target_id, "bait")
    if result and result.done then
        util.log("bought bait from " .. target_id .. " for " .. result.price)
        traded_this_tick[target_id] = true
    end
end

local function try_trade_at_market()
    local targets = world.nearby_entities()
    if not targets or #targets == 0 then
        return
    end
    traded_this_tick = {}

    local fish_count = num_fish_in_inventory()
    for _, target_id in ipairs(targets) do
        if target_id ~= self.name then
            if fish_count > 0 then
                sell_all_fish(target_id)
            end
            if #self.inventory == 0 or (num_fish_in_inventory() == 0 and fish_caught_total > 0) then
                buy_bait(target_id)
            end
        end
    end
end

local function do_fish()
    if world.phase ~= "day" then
        return
    end
    if world.tick % DODGE_TICKS ~= 0 then
        return
    end
    if #self.inventory > 0 and #self.inventory >= num_fish_in_inventory() + 2 then
        return
    end

    local roll = util.rand_int(100)
    if roll < FISH_CHANCE then
        local fish = FISH_TYPES[util.rand_int(#FISH_TYPES) + 1]
        local ok = world.add_item(fish)
        if ok then
            util.log("Oswin caught " .. fish .. " at the pond")
            fish_caught_total = fish_caught_total + 1
        else
            util.log("Oswin failed to add " .. fish .. " to inventory")
        end
    end
end

local function weather_blocks_fishing()
    local w = world.weather()
    if not w then return false end
    if w.stormy then return true end
    if w.type == "blizzard" or w.type == "thunderstorm" or w.type == "heavy_rain" then
        return true
    end
    return false
end

local function do_tick()
    local tick = world.tick
    local phase = world.phase

    if world.is_traveling() then
        return
    end

    if phase == "night" then
        if not is_at_pond() then
            world.move_to("stillwater_pond")
        end
        return
    end

    if phase == "dusk" then
        if not is_at_pond() then
            world.move_to("stillwater_pond")
        end
        return
    end

    if phase == "dawn" then
        if self.loc_id ~= "stillwater_pond" then
            world.move_to("stillwater_pond")
        end
        return
    end

    if is_at_pond() then
        if weather_blocks_fishing() then
            if world.tick % 30 == 0 then
                util.log("Oswin waits out the storm at the pond")
            end
            return
        end
        do_fish()
        return
    end

    if is_at_market() then
        try_trade_at_market()

        local fish_count = num_fish_in_inventory()
        if fish_count < MIN_FISH_TO_SELL and #self.inventory > 0 then
            return
        end
        if fish_count >= MIN_FISH_TO_SELL then
            world.move_to("stillwater_pond")
            return
        end
        return
    end

    if world.tick % 20 == 0 then
        world.move_to("stillwater_market")
    end
end

do_tick()