-- Lizardfolk AI - Amphibious Swamp Dweller
-- Territorial defender with ambush predator tactics
-- Hunts weaker prey opportunistically
-- Retreats to home swamp when injured, conserves energy at night
-- Gains pack bonus when kin are nearby

local ATTACK_CHANCE = 65
local AMBUSH_BONUS = 20
local FLEE_HP_THRESHOLD = 0.3
local HUNT_INTERVAL = 12
local WANDER_INTERVAL = 25

local function count_kin()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return 0
    end
    local count = 0
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.species == self.species then
                count = count + 1
            end
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
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and world.is_hostile(self.id, eid) then
            return eid, info
        end
        ::continue::
    end
    return nil
end

local function find_prey()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return nil
    end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and info.species ~= self.species then
            if not world.is_hostile(self.id, eid) then
                if info.hp and info.max_hp and info.hp < self.hp then
                    return eid, info
                end
            end
        end
        ::continue::
    end
    return nil
end

local function is_outnumbered()
    local kin = count_kin()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return false
    end
    local threats = 0
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and world.is_hostile(self.id, eid) then
                threats = threats + 1
            end
        end
    end
    return threats > kin + 1
end

local function should_flee()
    local hp_ratio = self.hp / self.max_hp
    if hp_ratio < FLEE_HP_THRESHOLD then
        return true
    end
    return is_outnumbered()
end

local function do_attack(target_id, target_info)
    local chance = ATTACK_CHANCE
    local kin = count_kin()
    if kin > 0 then
        chance = chance + AMBUSH_BONUS
        chance = chance + 5 * kin
    end

    if util.rand_int(100) < chance then
        local hit = world.attack(self.id, target_id)
        if hit then
            util.log(self.name .. " savagely bit " .. (target_info and target_info.name or target_id))
        end
        return true
    end
    return false
end

local function flee_home()
    if self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
        util.log(self.name .. " retreated to its swamp territory")
        return true
    end
    local exits = world.exits_from(self.loc_id)
    if exits and #exits > 0 then
        local dest = exits[util.rand_int(#exits) + 1]
        world.move_to(dest)
        util.log(self.name .. " slithered away from danger")
        return true
    end
    return false
end

function do_tick()
    local phase = world.phase

    if phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            return {util.event("move", {})}
        end
        return {}
    end

    if should_flee() then
        flee_home()
        if self.home and self.loc_id == self.home and self.hp < self.max_hp then
            world.heal(self.id, 3)
            util.log(self.name .. " submerged in swamp waters to recuperate")
        end
        return {util.event("flee", {})}
    end

    local hostile_id, hostile_info = find_hostile()
    if hostile_id then
        do_attack(hostile_id, hostile_info)
        return {util.event("attack", {})}
    end

    if world.tick % HUNT_INTERVAL == 0 then
        local prey_id, prey_info = find_prey()
        if prey_id then
            do_attack(prey_id, prey_info)
            return {util.event("hunt", {})}
        end
    end

    if self.home and self.loc_id == self.home and self.hp < self.max_hp and world.tick % 10 == 0 then
        world.heal(self.id, 3)
        util.log(self.name .. " rested in the swamp shallows")
        return {util.event("heal", {})}
    end

    if world.tick % WANDER_INTERVAL == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 and util.rand_int(100) < 30 then
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