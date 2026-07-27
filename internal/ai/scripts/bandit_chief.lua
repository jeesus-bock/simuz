-- Bandit Chief AI
-- A hardened leader who commands nearby bandits, raids towns, attacks merchants and
-- guards, and retreats to a camp when wounded. Can declare hostility shifts.

local ATTACK_INTERVAL = 3
local RAID_INTERVAL = 120
local FLEE_THRESHOLD = 0.3
local ATTACK_CHANCE = 80

local function count_bandits()
    local nearby = world.nearby_entities()
    if not nearby then return 0 end
    local n = 0
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.faction == "bandit" then
            n = n + 1
        end
    end
    return n
end

local function find_target()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and info.species ~= "deity" then
            if info.faction == "merchant" or info.faction == "civilian" or info.faction == "guard" or world.is_hostile(self.faction, info.faction) then
                return eid, info
            end
        end
        ::continue::
    end
    return nil
end

local function do_attack()
    local target_id, target_info = find_target()
    if not target_id then return false end
    if util.rand_int(100) < ATTACK_CHANCE then
        local hit = world.attack(self.id, target_id)
        if hit then
            util.log(self.name .. " led the raid against " .. target_info.name)
        end
        return hit
    end
    return false
end

local function should_flee()
    local hp_ratio = self.hp / self.max_hp
    if hp_ratio < FLEE_THRESHOLD then
        return true
    end
    local guards = 0
    local nearby = world.nearby_entities()
    if not nearby then return false end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.faction == "guard" then
            guards = guards + 1
        end
    end
    return guards >= 2
end

local function retreat_to_camp()
    if self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
        util.log(self.name .. " retreated to the bandit camp")
    end
end

local function do_tick()
    local tick = world.tick

    if should_flee() then
        retreat_to_camp()
        return
    end

    if tick % ATTACK_INTERVAL == 0 then
        do_attack()
    end

    if tick % RAID_INTERVAL == 0 then
        local town = world.parent_location(self.loc_id)
        if town then
            local buildings = world.exits_from(town)
            if buildings and #buildings > 0 then
                local dest = buildings[util.rand_int(#buildings) + 1]
                world.move_to(dest)
                util.log(self.name .. " raided " .. dest)
            end
        end
    end

    if tick % 60 == 0 then
        local bandits = count_bandits()
        if bandits >= 1 then
            util.set_mood("angry")
        end
    end
end

do_tick()
