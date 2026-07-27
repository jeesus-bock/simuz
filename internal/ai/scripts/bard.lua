-- Bard AI
-- Traveling musicians who perform in taverns for tips
-- Quality tiers affect earnings: lousy, mediocre, great
-- Can inspire allies and calm hostiles during performances
-- Some bards drink during performances for inspiration

local route = {
    "frosthold_inn_common",
    "stillwater_inn_common",
    "golden_gate_inn_common",
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

local QUALITY = {
    lousy = { tip_chance = 15, tip_min = 1, tip_max = 3, drink_chance = 5 },
    mediocre = { tip_chance = 30, tip_min = 2, tip_max = 8, drink_chance = 15 },
    great = { tip_chance = 50, tip_min = 5, tip_max = 15, drink_chance = 25 },
}

local quality = "mediocre"
local cha = 10
local function determine_quality()
    local effective_cha = cha
    if self.mood == "happy" then
        effective_cha = effective_cha + 2
    elseif self.mood == "stressed" then
        effective_cha = effective_cha - 2
    end
    if effective_cha < 10 then
        quality = "lousy"
    elseif effective_cha > 15 then
        quality = "great"
    else
        quality = "mediocre"
    end
end

local function is_at_tavern()
    for _, loc in ipairs(route) do
        if self.loc_id == loc then
            return true
        end
    end
    return false
end

local function is_weekend()
    return world.day_of_week == 6 or world.day_of_week == 7
end

local performance_tick = 0
local last_performance = 0

local function do_perform()
    if not is_at_tavern() then
        return false
    end
    
    if world.phase ~= "day" then
        return false
    end
    
    if world.tick - last_performance < 5 then
        return false
    end
    
    performance_tick = performance_tick + 1
    last_performance = world.tick
    
    local q = QUALITY[quality]
    local rolled = util.rand_int(100)
    
    if rolled < q.tip_chance then
        local tip = q.tip_min + util.rand_int(q.tip_max - q.tip_min + 1)
        local tip_type = util.rand_int(3)
        
        if tip_type == 0 then
            for i = 1, tip do
                world.add_item("cp")
            end
            util.log(self.name .. " earned " .. tip .. " copper pieces from performing")
        elseif tip_type == 1 then
            local drinks = {"beer", "wine", "ale"}
            local drink = drinks[util.rand_int(#drinks) + 1]
            world.add_item(drink)
            util.log(self.name .. " earned a " .. drink .. " from performing")
        else
            for i = 1, math.floor(tip / 2) do
                world.add_item("sp")
            end
            util.log(self.name .. " earned " .. math.floor(tip / 2) .. " silver pieces from performing")
        end
        
        return true
    end
    
    return false
end

local function do_inspire()
    if world.phase ~= "day" then
        return
    end
    
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return
    end
    
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.faction == "civilian" then
            util.mem_set(eid .. "_inspired", true)
            util.log(self.name .. " inspired " .. info.name)
        end
    end
end

local function do_calm_hostiles()
    if world.phase ~= "day" then
        return
    end
    
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return
    end
    
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.faction ~= "civilian" and info.faction ~= "merchant" then
            if world.is_hostile("civilian", info.faction) then
                util.mem_set(eid .. "_calmed", true)
                util.log(self.name .. " calmed hostile " .. info.name)
            end
        end
    end
end

local function maybe_drink()
    if quality == "great" and util.rand_int(100) < QUALITY.great.drink_chance then
        local drinks = {"wine", "mead", "brandy"}
        local drink = drinks[util.rand_int(#drinks) + 1]
        local result = world.use_item(drink)
        if result then
            util.log(self.name .. " drank " .. drink .. " for inspiration")
            util.set_mood("inspired")
        end
    elseif quality == "mediocre" and util.rand_int(100) < QUALITY.mediocre.drink_chance then
        local drinks = {"beer", "ale"}
        local drink = drinks[util.rand_int(#drinks) + 1]
        local result = world.use_item(drink)
        if result then
            util.log(self.name .. " drank " .. drink .. " for inspiration")
            util.set_mood("inspired")
        end
    end
end

local function do_tick()
    local tick = world.tick
    local phase = world.phase

    if world.is_traveling() then
        return
    end

    -- resync index if we are at a known stop
    for i, v in ipairs(route) do
        if v == self.loc_id then
            current_index = i
            break
        end
    end
    
    determine_quality()
    
    if phase == "night" then
        if not is_at_tavern() then
            local dest = pick_next_destination()
            if dest and dest ~= self.loc_id then
                local ok = world.move_to(dest)
                if ok then
                    current_index = (current_index % #route) + 1
                end
            end
        end
        util.set_mood("sleepy")
        return
    end
    
    if phase == "day" then
        if is_at_tavern() then
            if tick % 3 == 0 then
                local performed = do_perform()
                if performed then
                    do_inspire()
                    do_calm_hostiles()
                    maybe_drink()
                end
            end
        else
            if tick % 15 == 0 then
                local dest = pick_next_destination()
                if dest and dest ~= self.loc_id then
                    local ok = world.move_to(dest)
                    if ok then
                        current_index = (current_index % #route) + 1
                    end
                end
            end
        end
    end
    
    if tick % 20 == 0 then
        local roll = util.rand_int(100)
        if roll < 30 then
            util.set_mood("happy")
        elseif roll < 60 then
            util.set_mood("neutral")
        else
            util.set_mood("stressed")
        end
    end
end

do_tick()
