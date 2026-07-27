-- Guard AI
-- Patrols between guardhouse and town locations.
-- Attacks hostiles (thieves, bandits) on sight.
-- Returns to guardhouse at dusk.

local patrol_targets = nil
local patrol_index = 1
local patrolling = false

local function get_patrol_route()
    if patrol_targets then return patrol_targets end
    local town = world.parent_location(self.home)
    if not town then
        patrol_targets = {}
        return patrol_targets
    end
    local town_buildings = world.exits_from(town)
    if not town_buildings or #town_buildings == 0 then
        patrol_targets = {}
        return patrol_targets
    end
    patrol_targets = {}
    for _, bid in ipairs(town_buildings) do
        if bid ~= self.home then
            table.insert(patrol_targets, bid)
        end
    end
    table.insert(patrol_targets, self.home)
    return patrol_targets
end

local function should_return_home()
    local phase = world.phase
    return phase == "dusk" or phase == "night"
end

local function attack_hostiles()
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive then
            if world.is_hostile(self.faction, info.faction) then
                if not world.attack then return false end
                local hit = world.attack(self.id, eid)
                if hit then
                    util.log(self.name .. " attacked " .. (info.name or eid))
                end
                return true
            end
        end
    end
    return false
end

local function do_patrol()
    if world.tick % 15 ~= 0 then return end
    if patrolling then
        local route = get_patrol_route()
        if #route == 0 then return end
        patrol_index = patrol_index + 1
        if patrol_index > #route then
            patrol_index = 1
        end
        local target = route[patrol_index]
        if target ~= self.loc_id then
            world.move_to(target)
            util.log(self.name .. " patrolling to " .. target)
        end
    end
end

local function maybe_drink()
    if world.phase ~= "day" then
        return
    end
    
    local is_weekend = world.day_of_week == 6 or world.day_of_week == 7
    
    local drink_chance = 5
    if is_weekend then
        drink_chance = 15
    end
    
    if util.rand_int(100) < drink_chance then
        local drinks = {"beer", "ale", "wine"}
        local drink = drinks[util.rand_int(#drinks) + 1]
        local result = world.use_item(drink)
        if result then
            util.log(self.name .. " drank " .. drink .. " after patrol")
            util.set_mood("relaxed")
        end
    end
end

local function update_mood()
    if self.mood == "drunk" or self.mood == "relaxed" then
        return
    end
    
    if world.phase == "night" then
        util.set_mood("tired")
        return
    end

    local w = world.weather()
    if w and w.type == "fog" then
        util.set_mood("stressed")
        return
    end
    
    local roll = util.rand_int(100)
    if roll < 20 then
        util.set_mood("stressed")
    elseif roll < 40 then
        util.set_mood("neutral")
    elseif roll < 60 then
        util.set_mood("alert")
    end
end

local function do_tick()
    if should_return_home() then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
        end
        patrolling = false
        return
    end

    if attack_hostiles() then
        return
    end

    -- Prefer contested buildings when territory is hostile-controlled
    if world.phase == "day" and world.tick % 20 == 0 then
        local town = world.parent_location(self.home)
        if town then
            local buildings = world.exits_from(town)
            for _, bid in ipairs(buildings or {}) do
                local ctl = world.location_control(bid)
                if ctl and ctl.faction ~= "" and ctl.faction ~= "civilian" and ctl.faction ~= "guard" then
                    world.move_to(bid)
                    util.log(self.name .. " moves to contested " .. bid .. " (" .. ctl.faction .. ")")
                    return
                end
            end
        end
    end

    if world.phase == "day" then
        patrolling = true
        do_patrol()
        if world.tick % 10 == 0 then
            maybe_drink()
        end
    end

    if world.phase == "dawn" and self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
        patrolling = true
    end
    
    if world.tick % 30 == 0 then
        update_mood()
    end
end

do_tick()
