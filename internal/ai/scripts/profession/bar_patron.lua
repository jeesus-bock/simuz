-- Bar Patron AI
-- A regular tavern-goer who listens to bards, buys drinks, and occasionally tips
-- performers. Sits in the common room during the evening, returns home at night.

local DRINKS = {
    "beer", "ale", "wine", "mead"
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

local function find_bard()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive then
            local items = world.entity_items(eid)
            if items then
                for _, item_id in ipairs(items) do
                    if item_id == "lute" or item_id == "flute" then
                        return eid, info
                    end
                end
            end
        end
    end
    return nil
end

local function buy_drink()
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, target_id in ipairs(nearby) do
        local info = world.entity_info(target_id)
        if info and info.alive then
            local items = world.entity_items(target_id)
            if items then
                for _, item_id in ipairs(items) do
                    for _, drink in ipairs(DRINKS) do
                        if item_id == drink then
                            local result = world.try_buy(target_id, drink)
                            if result and result.done then
                                util.log(self.name .. " bought a " .. drink)
                                return true
                            end
                        end
                    end
                end
            end
        end
    end
    return false
end

local function tip_bard(bard_id)
    if util.rand_int(100) < 30 then
        for i = 1, 2 do
            world.add_item("cp")
        end
        local result = world.try_sell(bard_id, "cp")
        if result and result.done then
            util.log(self.name .. " tipped the bard")
            util.set_mood("happy")
            return true
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
            util.log(self.name .. " went home from the tavern")
            return true
        end
        return false
    end

    if tick % 10 == 0 then
        buy_drink()
        if has_item("beer") or has_item("ale") or has_item("wine") or has_item("mead") then
            for _, drink in ipairs(DRINKS) do
                if has_item(drink) then
                    world.use_item(drink)
                    util.log(self.name .. " sipped " .. drink)
                    break
                end
            end
        end
    end

    if tick % 15 == 0 then
        local bard_id, bard_info = find_bard()
        if bard_id then
            tip_bard(bard_id)
        end
    end

    if tick % 60 == 0 and self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
    end

    if tick % 35 == 0 then
        local roll = util.rand_int(100)
        if roll < 50 then
            util.set_mood("happy")
        elseif roll < 80 then
            util.set_mood("relaxed")
        end
    end

    return false
end

return do_tick()
