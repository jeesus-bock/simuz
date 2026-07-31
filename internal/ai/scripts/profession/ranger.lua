-- Ranger AI
-- A wilderness hunter who tracks and kills hostile beasts, gathers pelts and meat,
-- sells them to merchants or civilians, and camps at night. Stays in forests and fields.

local HUNT_INTERVAL = 5
local GATHER_INTERVAL = 25
local SELL_INTERVAL = 40

local HUNTED_FACTIONS = {
    beast = true, vermin = true
}

local PELTS_AND_MEAT = {
    "raw_chicken", "raw_pork", "raw_beef", "raw_mutton", "raw_goat", "leather", "wool", "feather"
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

local function find_prey()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and HUNTED_FACTIONS[info.faction] then
            return eid, info
        end
        ::continue::
    end
    return nil
end

local function do_hunt()
    local prey_id, prey_info = find_prey()
    if not prey_id then return false end
    local hit = world.attack(self.id, prey_id)
    if hit then
        util.log(self.name .. " hunted " .. prey_info.name)
        if not prey_info.alive then
            local drops = {"leather", "feather", "raw_chicken", "raw_pork"}
            local drop = drops[util.rand_int(#drops) + 1]
            world.add_item(drop)
        end
    end
    return hit
end

local function sell_goods()
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, target_id in ipairs(nearby) do
        local info = world.entity_info(target_id)
        if info and info.alive and (info.faction == "merchant" or info.faction == "civilian") then
            for _, good in ipairs(PELTS_AND_MEAT) do
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

local function make_camp()
    if self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
        util.log(self.name .. " made camp")
    end
end

local function do_tick()
    local phase = world.phase
    local tick = world.tick

    if phase == "night" then
        make_camp()
        return true
    end

    if tick % HUNT_INTERVAL == 0 then
        do_hunt()
    end

    if tick % SELL_INTERVAL == 0 then
        sell_goods()
    end

    if tick % 60 == 0 and util.rand_int(100) < 40 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            if dest ~= self.loc_id then
                world.move_to(dest)
            end
        end
    end

    if tick % 45 == 0 then
        local roll = util.rand_int(100)
        if roll < 30 then
            util.set_mood("focused")
        elseif roll < 60 then
            util.set_mood("relaxed")
        end
    end

    return false
end

return do_tick()
