-- Smuggler AI
-- An illegal trader who moves contraband between settlements,
-- avoiding authorities and selling to the highest bidder.

local SELL_CHANCE = 55
local BOUNTY_HUNTER_CHANCE = 20
local FLEE_HP_THRESHOLD = 0.35

local CONTRABAND = {
    "charged_quartz", "herb_pouch", "dagger", "short_sword",
    "beer", "wine", "mead", "ale", "bandage", "iron_ore",
}

local function find_buyer()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and info.profession ~= "guard" and info.profession ~= "deity" then
            if info.profession == "civilian" or info.profession == "merchant" or info.profession == "bandit" then
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
        if info and info.alive and info.profession == "guard" then
            return eid, info
        end
    end
    return nil
end

local function has_contraband()
    for _, item_id in ipairs(self.inventory) do
        for _, contraband_id in ipairs(CONTRABAND) do
            if item_id == contraband_id then return true, item_id end
        end
    end
    return false, nil
end

local function sell_contraband()
    local has_cb, item_id = has_contraband()
    if not has_cb then return false end

    -- Use SELL_CHANCE to determine if the smuggler successfully sells
    if math.random(100) > SELL_CHANCE then
        return false
    end

    local buyer_id, buyer_info = find_buyer()
    if not buyer_id then return false end

    local result = world.try_sell(buyer_id, item_id)
    if result and result.done then
        util.set_mood("sneaky")
        util.log(self.name .. " smuggled " .. item_id .. " to " .. buyer_info.name)
        return true
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
            util.log(self.name .. " slipped away from authorities")
        end
    end
end

function do_tick()
    local tick = world.tick
    local events = {}

    if world.defend_self and world.defend_self() then
        table.insert(events, util.event("profession_action", { profession = "smuggler" }))
        return events
    end
    if world.avoid_combat and world.avoid_combat() then
        table.insert(events, util.event("profession_action", { profession = "smuggler" }))
        return events
    end

    if should_flee() then
        flee()
        util.set_mood("stressed")
        table.insert(events, util.event("profession_action", { profession = "smuggler" }))
        return events
    end

    if tick % 5 == 0 then
        sell_contraband()
    end

    if tick % 20 == 0 then
        if util.rand_int(100) < BOUNTY_HUNTER_CHANCE then
            local exits = world.exits_from(self.loc_id)
            if exits and #exits > 0 then
                local dest = exits[util.rand_int(#exits) + 1]
                if dest ~= self.loc_id then
                    world.move_to(dest)
                end
            end
        end
    end

    return events
end

return do_tick()
