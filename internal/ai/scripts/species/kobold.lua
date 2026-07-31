-- Kobold AI - Pack Tactics
-- Gains bonus when 2+ allies nearby
-- Sets traps that damage attackers
-- Flees when outnumbered or low HP

local ATTACK_CHANCE = 60
local TRAP_CHANCE = 25
local FLEE_HP_THRESHOLD = 0.5
local WANDER_INTERVAL = 20

local function count_allies()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return 0
    end
    local count = 0
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.faction == self.faction then
            count = count + 1
        end
    end
    return count
end

local function find_hostile()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return nil
    end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and world.is_hostile(self.faction, info.faction) then
            return eid, info
        end
    end
    return nil
end

local function is_outnumbered()
    local allies = count_allies()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return false
    end
    local enemies = 0
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and world.is_hostile(self.faction, info.faction) then
            enemies = enemies + 1
        end
    end
    return enemies > allies
end

local function should_flee()
    local hp_ratio = self.hp / self.max_hp
    return hp_ratio < FLEE_HP_THRESHOLD or is_outnumbered()
end

local function do_attack(eid)
    local allies = count_allies()
    local attack_chance = ATTACK_CHANCE
    if allies >= 2 then
        attack_chance = attack_chance + 20
    end

    if util.rand_int(100) < attack_chance then
        local hit = world.attack(self.id, eid)
        if hit then
            local info = world.entity_info(eid)
            util.log(self.name .. " pack-attacked " .. (info and info.name or eid) .. " with " .. allies .. " allies nearby")
        end
        return true
    end
    return false
end

local function do_trap_damage(attacker_id)
    if util.rand_int(100) < TRAP_CHANCE then
        local trap_dmg = util.rand_int(6) + 1
        util.log(self.name .. "'s trap dealt " .. trap_dmg .. " damage to attacker")
        return true
    end
    return false
end

function do_tick()
    local phase = world.phase

    if phase == "night" or phase == "dusk" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            return {util.event("move", {})}
        end
        return {}
    end

    if should_flee() then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            world.move_to(dest)
            util.log(self.name .. " fled from overwhelming odds")
            return {util.event("flee", {})}
        end
        return {}
    end

    local target_id, target_info = find_hostile()
    if target_id then
        do_attack(target_id)
        return {util.event("attack", {})}
    end

    if world.tick % WANDER_INTERVAL == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            if dest ~= self.loc_id then
                world.move_to(dest)
                return {util.event("move", {})}
            end
        end
    end
    return {}
end

return do_tick()
