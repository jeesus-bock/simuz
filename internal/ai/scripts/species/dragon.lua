-- Dragon AI
-- A slumbering boss who guards its hoard in the dragon lair. When intruders enter,
-- it unleashes a fiery breath attack that damages all nearby non-allies, then attacks
-- the strongest target. It heals slowly while in the lair and returns there if it
-- leaves to chase prey.

local BREATH_INTERVAL = 8
local ATTACK_INTERVAL = 4
local HEAL_INTERVAL = 15
local HEAL_AMOUNT = 10
local FLEE_THRESHOLD = 0.2

local function find_intruder()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and info.species ~= "deity" and info.faction ~= "dragon" then
            return eid, info
        end
        ::continue::
    end
    return nil
end

local function do_breath_attack()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then return false end
    local damage = math.floor(self.level * 2 + 5)
    local result = world.damage_location(self.id, damage)
    if result and result.targets > 0 then
        util.log(self.name .. " breathed fire for " .. damage .. " damage on " .. result.targets .. " targets")
    end
    return true
end

local function do_heal()
    if self.hp < self.max_hp then
        world.heal(self.id, HEAL_AMOUNT)
    end
end

local function return_to_lair()
    if self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
        util.log(self.name .. " returned to its hoard")
    end
end

local function do_tick()
    local tick = world.tick

    if self.hp / self.max_hp < FLEE_THRESHOLD then
        return_to_lair()
        do_heal()
        return true
    end

    local intruder, info = find_intruder()
    if intruder then
        if tick % BREATH_INTERVAL == 0 then
            do_breath_attack()
        end
        if tick % ATTACK_INTERVAL == 0 then
            local hit = world.attack(self.id, intruder)
            if hit then
                util.log(self.name .. " clawed at " .. info.name)
            end
        end
        return true
    end

    if self.home and self.loc_id ~= self.home then
        return_to_lair()
        return true
    end

    if tick % HEAL_INTERVAL == 0 then
        do_heal()
    end

    if tick % 100 == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 and util.rand_int(100) < 20 then
            local dest = exits[util.rand_int(#exits) + 1]
            world.move_to(dest)
            util.log(self.name .. " stretched its wings and left the lair")
            return true
        end
    end
    return false
end

return do_tick()
