-- Rogue AI
-- A skilled assassin and infiltrator who strikes from the shadows,
-- poisons targets, and escapes before guards can respond.

local POISON_CHANCE = 25
local ASSASSINATE_CHANCE = 15
local FLEE_HP_THRESHOLD = 0.3

local function find_high_value_target()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and info.faction ~= self.faction and info.faction ~= "deity" then
            if info.faction == "civilian" or info.faction == "merchant" or info.faction == "bandit" then
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

local function poison_target(target_id)
    if util.rand_int(100) >= POISON_CHANCE then return false end
    local result = world.apply_effect(target_id, "poisoned")
    if result and result.done then
        util.log(self.name .. " poisoned " .. target_id)
        return true
    end
    return false
end

local function assassinate(target_id)
    if util.rand_int(100) >= ASSASSINATE_CHANCE then return false end
    local result = world.attack(target_id)
    if result and result.done then
        util.set_mood("cold")
        util.log(self.name .. " assassinated " .. target_id)
        return true
    end
    return false
end

local function steal_and_flee()
    local target_id, info = find_high_value_target()
    if not target_id then return false end
    local items = world.entity_items(target_id)
    if not items or #items == 0 then return false end
    local stolen = items[util.rand_int(#items) + 1]
    local result = world.steal(target_id, stolen)
    if result and result.done then
        util.set_mood("sneaky")
        util.log(self.name .. " stole " .. stolen .. " from " .. info.name)
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
            util.log(self.name .. " fled into the shadows")
        end
    end
end

function do_tick()
    local tick = world.tick

    if should_flee() then
        flee()
        util.set_mood("stressed")
        return {util.event("profession_action", {profession = "rogue"})}
    end

    if tick % 8 == 0 then
        steal_and_flee()
    end

    if tick % 12 == 0 then
        local target_id, _ = find_high_value_target()
        if target_id then
            poison_target(target_id)
            assassinate(target_id)
        end
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