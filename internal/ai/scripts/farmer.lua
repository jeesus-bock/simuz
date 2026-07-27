local home = self.home
local farm_id = self.loc_id

local SLAUGHTER_AGE = {
    chicken = 900,
    pig = 2400,
    cow = 4800,
    sheep = 3600,
    goat = 3600
}

local MEAT_LOOKUP = {
    chicken = "raw_chicken",
    pig = "raw_pork",
    cow = "raw_beef",
    sheep = "raw_mutton",
    goat = "raw_goat"
}

local leashed_companion_id = nil

local function is_livestock(species)
    return SLAUGHTER_AGE[species] ~= nil
end

local function find_farm_dog()
    local nearby = world.nearby_entities()
    if not nearby then
        return nil
    end
    for _, other_id in ipairs(nearby) do
        local info = world.entity_info(other_id)
        if info and info.alive and info.species == "dog" then
            return other_id, info
        end
    end
    return nil
end

local function leash_companion()
    local dog_id, dog_info = find_farm_dog()
    if not dog_id then
        return false
    end
    local leashed, dragger = world.is_leashed(dog_id)
    if leashed and dragger ~= self.id then
        return false
    end
    local ok = world.drag_entity(dog_id)
    if ok then
        leashed_companion_id = dog_id
        util.log(self.name .. " leashed " .. (dog_info and dog_info.name or dog_id) .. " for the trip")
    end
    return ok
end

local function release_companion()
    if not leashed_companion_id then
        return false
    end
    local ok = world.undrag_entity(leashed_companion_id)
    if ok then
        util.log(self.name .. " released " .. leashed_companion_id .. " back at the farm")
    end
    leashed_companion_id = nil
    return ok
end

local function count_items(def_id)
    local count = 0
    for _, id in ipairs(self.inventory) do
        if id == def_id then
            count = count + 1
        end
    end
    return count
end

local function has_item(def_id)
    return count_items(def_id) > 0
end

local function sell_at_market()
    local market_id = world.parent_location(farm_id)
    if not market_id then
        return
    end
    local market = world.parent_location(market_id)
    if not market then
        market = market_id
    end
    local sell_items = {"raw_chicken", "raw_pork", "raw_beef", "raw_mutton", "raw_goat", "egg", "milk", "wool", "feather", "leather", "grain"}
    for _, item_id in ipairs(sell_items) do
        if has_item(item_id) then
            local nearby = world.entities_at(market)
            for _, buyer_id in ipairs(nearby) do
                local info = world.entity_info(buyer_id)
                if info and info.alive and (info.faction == "merchant" or info.faction == "civilian") then
                    local result = world.try_sell(buyer_id, item_id)
                    if result.done then
                        util.log("Sold " .. item_id .. " for " .. result.price)
                        return
                    end
                end
            end
        end
    end
end

local function go_to_market()
    local market_id = world.parent_location(farm_id)
    if market_id then
        leash_companion()
        world.move_to(market_id)
    end
end

local function return_to_farm()
    world.move_to(farm_id)
    release_companion()
end

local function tend_animals()
    local nearby = world.nearby_entities()
    for _, other_id in ipairs(nearby) do
        local info = world.entity_info(other_id)
        if not info or not info.alive then
            goto continue
        end
        if not is_livestock(info.species) then
            goto continue
        end
        if info.hunger and info.hunger > 0.6 then
            world.feed(other_id)
            util.log("Fed " .. info.name)
        end
        local sla = SLAUGHTER_AGE[info.species]
        if sla and info.age and info.age > sla then
            local hit = world.attack(self.id, other_id)
            if hit then
                local meat = MEAT_LOOKUP[info.species]
                if meat then
                    world.add_item(meat)
                end
                local extras = {"feather", "leather", "wool"}
                if info.species ~= "chicken" then
                    world.add_item("leather")
                else
                    world.add_item("feather")
                end
                if info.species == "sheep" then
                    world.add_item("wool")
                end
                util.log("Slaughtered " .. info.name .. ", got " .. meat)
            end
        end
        ::continue::
    end
end

local phase = world.phase
local tick = world.tick
local wth = world.weather()

if world.defend_self and world.defend_self() then
    return
end

if phase == "dawn" then
    return_to_farm()
elseif phase == "day" then
    if tick % 5 == 0 then
        tend_animals()
    end
    if wth and wth.stormy then
        if tick % 20 == 0 then
            return_to_farm()
            util.log(self.name .. " stays at farm due to storm")
        end
    elseif tick % 40 == 0 then
        go_to_market()
        sell_at_market()
        return_to_farm()
    end
elseif phase == "dusk" then
    return_to_farm()
elseif phase == "night" then
    if self.loc_id ~= farm_id then
        return_to_farm()
    end
end
