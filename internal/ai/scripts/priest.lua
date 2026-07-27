-- Priest AI
-- Heals wounded civilians. Returns to home temple at dusk.
-- Wanders to nearby buildings during the day.

local function find_wounded()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.hp < info.max_hp and info.faction == "civilian" then
            return eid, info
        end
    end
    return nil
end

local function heal_step()
    if not world.divine_intervention then return end
    local target, info = find_wounded()
    if target then
        local result = world.divine_intervention(self.id, target, "heal")
        if result and result.done then
            util.log(self.name .. " healed " .. (info.name or target) .. " for " .. result.amount)
        end
    end
end

local function maybe_drink()
    if world.phase ~= "day" then
        return
    end
    
    local is_weekend = world.day_of_week == 6 or world.day_of_week == 7
    
    local drink_chance = 3
    if is_weekend then
        drink_chance = 10
    end
    
    if util.rand_int(100) < drink_chance then
        local drinks = {"wine", "mead"}
        local drink = drinks[util.rand_int(#drinks) + 1]
        local result = world.use_item(drink)
        if result then
            util.log(self.name .. " drank " .. drink .. " for divine communion")
            util.set_mood("contemplative")
        end
    end
end

local function update_mood()
    if self.mood == "contemplative" then
        return
    end
    
    if world.phase == "night" then
        util.set_mood("prayerful")
        return
    end
    
    local roll = util.rand_int(100)
    if roll < 15 then
        util.set_mood("stressed")
    elseif roll < 35 then
        util.set_mood("neutral")
    elseif roll < 55 then
        util.set_mood("serene")
    end
end

local function do_tick()
    local phase = world.phase

    if phase == "dusk" or phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
        end
        return
    end

    if world.tick % 10 == 0 then
        heal_step()
    end

    if world.tick % 20 == 0 then
        maybe_drink()
    end

    if world.tick % 60 == 0 and self.home then
        if self.loc_id == self.home then
            local town = world.parent_location(self.home)
            if town then
                local buildings = world.exits_from(town)
                if buildings then
                    for _, bid in ipairs(buildings) do
                        if bid ~= self.home then
                            world.move_to(bid)
                            util.log(self.name .. " visiting " .. bid)
                            break
                        end
                    end
                end
            end
        else
            world.move_to(self.home)
            util.log(self.name .. " returned to temple")
        end
    end
    
    if world.tick % 30 == 0 then
        update_mood()
    end
end

do_tick()
