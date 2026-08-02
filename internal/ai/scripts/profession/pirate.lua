-- Pirate AI
-- A seafaring raider who plunders coastal settlements,
-- raids ships, and sells stolen cargo on the black market.

local RAID_CHANCE = 45
local FLEE_HP_THRESHOLD = 0.2

local function find_coastal_target()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and info.faction ~= self.faction and info.faction ~= "deity" then
            if info.faction == "civilian" or info.faction == "merchant" then
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

local function raid_target(target_id)
    if util.rand_int(100) >= RAID_CHANCE then return false end
    local result = world.attack(target_id)
    if result and result.done then
        util.set_mood("aggressive")
        util.log(self.name .. " raided " .. target_id)
        return true
    end
    return false
end

local function sell_stolen_cargo()
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, target_id in ipairs(nearby) do
        local info = world.entity_info(target_id)
        if info and info.alive and (info.faction == "merchant" or info.faction == "civilian" or info.faction == "bandit") then
            for _, item_id in ipairs(self.inventory) do
                local result = world.try_sell(target_id, item_id)
                if result and result.done then
                    util.log(self.name .. " sold stolen cargo to " .. info.name)
                    return true
                end
            end
        end
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
            util.log(self.name .. " sailed away")
        end
    end
end

function do_tick()
    local tick = world.tick

    if world.defend_self and world.defend_self() then
        return {util.event("profession_action", {profession = "pirate"})}
    end
    if world.avoid_combat and world.avoid_combat() then
        return {util.event("profession_action", {profession = "pirate"})}
    end

    if should_flee() then
        flee()
        util.set_mood("stressed")
        return {util.event("profession_action", {profession = "pirate"})}
    end

    if tick % 6 == 0 then
        local target_id, _ = find_coastal_target()
        if target_id then
            raid_target(target_id)
        end
    end

    if tick % 15 == 0 then
        sell_stolen_cargo()
    end

    if tick % 40 == 0 then
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