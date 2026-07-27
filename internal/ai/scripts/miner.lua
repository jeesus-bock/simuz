-- Miner AI
-- Digs ore and coal at the mine during the day, sells raw materials to blacksmiths,
-- and returns to town to drink and sleep at night.

local MINE_INTERVAL = 15
local SELL_INTERVAL = 25
local HOME_INTERVAL = 80

local MINED_GOODS = {
    "iron_ore", "coal", "iron_ore", "iron_ore", "coal"
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

local function mine()
    local pick = "iron_ore"
    if util.rand_int(100) < 40 then
        pick = "coal"
    end
    world.add_item(pick)
    util.log(self.name .. " mined " .. pick)
end

local function sell_to_blacksmith()
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, target_id in ipairs(nearby) do
        local info = world.entity_info(target_id)
        if info and info.alive and (info.faction == "civilian" or info.faction == "merchant") then
            for _, good in ipairs(MINED_GOODS) do
                if has_item(good) then
                    local result = world.try_sell(target_id, good)
                    if result and result.done then
                        util.log(self.name .. " sold " .. good .. " to " .. info.name)
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

    if phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            util.log(self.name .. " left the mine for the night")
        end
        if tick % 30 == 0 then
            local drinks = {"beer", "ale"}
            local drink = drinks[util.rand_int(#drinks) + 1]
            if has_item(drink) then
                world.use_item(drink)
                util.log(self.name .. " enjoyed a " .. drink .. " after work")
            end
        end
        return
    end

    if tick % MINE_INTERVAL == 0 then
        mine()
    end

    if tick % SELL_INTERVAL == 0 then
        sell_to_blacksmith()
    end

    if tick % HOME_INTERVAL == 0 and self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
    end

    if tick % 50 == 0 then
        util.set_mood("tired")
    end
end

do_tick()
