-- Brigand AI
-- A highwayman who roams roads and ambushes travelers,
-- demanding tribute or fighting when refused.

local ROBBERY_CHANCE = 50
local FLEE_HP_THRESHOLD = 0.25

local function find_traveler()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and info.faction ~= self.faction and info.faction ~= "deity" then
            if info.faction == "civilian" or info.faction == "merchant" or info.faction == "traveler" then
                return eid, info
            end
        end
        ::continue::
    end
    return nil
end

local function find_guard()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.faction == "guard" then
            return eid, info
        end
    end
    return nil
end

local function rob_target(target_id)
    if util.rand_int(100) >= ROBBERY_CHANCE then return false end
    local items = world.entity_items(target_id)
    if not items or #items == 0 then return false end
    local target_item = items[util.rand_int(#items) + 1]
    local result = world.steal(target_id, target_item)
    if result and result.done then
        util.set_mood("aggressive")
        util.log(self.name .. " robbed " .. target_item .. " from " .. target_id)
        return true
    end
    return false
end

local function demand_tribute()
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and info.faction == "civilian" then
            local result = world.try_sell(eid, "gold")
            if result and result.done then
                util.log(self.name .. " demanded tribute from " .. info.name)
                return true
            end
        end
        ::continue::
    end
    return false
end

local function should_flee()
    local hp_ratio = self.hp / self.max_hp
    if hp_ratio < FLEE_HP_THRESHOLD then return true end
    if find_guard() then return true end
    return false
end

local function flee()
    local exits = world.exits_from(self.loc_id)
    if exits and #exits > 0 then
        local dest = exits[util.rand_int(#exits) + 1]
        if dest ~= self.loc_id then
            world.move_to(dest)
            util.log(self.name .. " fled from the law")
        end
    end
end

function do_tick()
    local tick = world.tick

    if world.defend_self and world.defend_self() then
        return {util.event("profession_action", {profession = "brigand"})}
    end
    if world.avoid_combat and world.avoid_combat() then
        return {util.event("profession_action", {profession = "brigand"})}
    end

    if should_flee() then
        flee()
        util.set_mood("stressed")
        return {util.event("profession_action", {profession = "brigand"})}
    end

    if tick % 6 == 0 then
        rob_target(find_traveler())
    end

    if tick % 10 == 0 then
        demand_tribute()
    end

    if tick % 30 == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            if dest ~= self.loc_id then
                world.move_to(dest)
            end
        end
    end

    return {}
end

return do_tick()