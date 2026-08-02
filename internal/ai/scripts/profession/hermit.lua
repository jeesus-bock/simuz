-- Hermit AI
-- A reclusive wanderer who lives on the edge of civilization,
-- foraging for food and avoiding contact with others.

local FORAGE_CHANCE = 30
local AVOID_CHANCE = 50

local function find_forage_spot()
    local exits = world.exits_from(self.loc_id)
    if not exits or #exits == 0 then return nil end
    for _, dest in ipairs(exits) do
        if dest ~= self.loc_id then
            return dest
        end
    end
    return nil
end

local function find_nearby()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive then
            return eid, info
        end
        ::continue::
    end
    return nil
end

local function forage()
    local spot = find_forage_spot()
    if spot then
        world.move_to(spot)
        util.log(self.name .. " foraged in the wilderness")
        if util.rand_int(100) < 50 then
            util.set_mood("neutral")
        end
        return true
    end
    return false
end

local function avoid_crowds()
    local nearby = find_nearby()
    if not nearby then return false end
    if util.rand_int(100) < AVOID_CHANCE then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            if dest ~= self.loc_id then
                world.move_to(dest)
                util.log(self.name .. " retreated into solitude")
                return true
            end
        end
    end
    return false
end

function do_tick()
    local tick = world.tick

    if world.defend_self and world.defend_self() then
        return {util.event("profession_action", {profession = "hermit"})}
    end

    if tick % 8 == 0 then
        avoid_crowds()
    end

    if tick % 15 == 0 then
        forage()
    end

    if tick % 50 == 0 then
        util.set_mood("contemplative")
    end

    return {}
end

return do_tick()