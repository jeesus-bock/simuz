-- Beggar AI
-- A destitute wanderer who asks for coins and scraps,
-- follows the generous and avoids the hostile.

local BEG_CHANCE = 40
local FOLLOW_CHANCE = 30

local function find_generous()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and (info.faction == "civilian" or info.faction == "merchant" or info.faction == "innkeeper") then
            return eid, info
        end
        ::continue::
    end
    return nil
end

local function find_hostile()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and info.faction == "bandit" or info.faction == "thief" or info.faction == "rogue" then
            return eid, info
        end
        ::continue::
    end
    return nil
end

local function beg_for_coins()
    local generous_id, generous_info = find_generous()
    if not generous_id then return false end
    local result = world.try_beg(generous_id)
    if result and result.done then
        util.set_mood("happy")
        util.log(self.name .. " begged coins from " .. generous_info.name)
        return true
    end
    return false
end

local function flee_from_hostile()
    local hostile_id, _ = find_hostile()
    if not hostile_id then return false end
    local exits = world.exits_from(self.loc_id)
    if exits and #exits > 0 then
        local dest = exits[util.rand_int(#exits) + 1]
        if dest ~= self.loc_id then
            world.move_to(dest)
            util.log(self.name .. " fled from danger")
            return true
        end
    end
    return false
end

function do_tick()
    local tick = world.tick

    if world.defend_self and world.defend_self() then
        return {util.event("profession_action", {profession = "beggar"})}
    end

    if tick % 6 == 0 then
        flee_from_hostile()
    end

    if tick % 10 == 0 then
        beg_for_coins()
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