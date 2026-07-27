-- Cultist AI
-- A fanatic who performs rituals at night, attacks non-cultists, and attempts to
-- summon shadow-touched strength. Flee at low health and try to regroup at home.

local ATTACK_INTERVAL = 4
local RITUAL_INTERVAL = 30
local FLEE_THRESHOLD = 0.35
local ATTACK_CHANCE = 75

local function find_victim()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and not world.is_hostile(self.faction, self.faction) then
            if info.faction ~= "cultist" and info.species ~= "deity" then
                return eid, info
            end
        end
        ::continue::
    end
    return nil
end

local function do_ritual()
    if util.rand_int(100) < 50 then
        util.mem_set("blessed_by_shadow", world.tick + 40)
        util.log(self.name .. " performed a blood ritual")
        util.set_mood("inspired")
        return true
    end
    return false
end

local function should_flee()
    return self.hp / self.max_hp < FLEE_THRESHOLD
end

local function do_attack()
    local victim_id, victim_info = find_victim()
    if not victim_id then return false end
    if util.rand_int(100) < ATTACK_CHANCE then
        local hit = world.attack(self.id, victim_id)
        if hit then
            util.log(self.name .. " struck " .. victim_info.name .. " for the cult")
        end
        return hit
    end
    return false
end

local function do_tick()
    local phase = world.phase
    local tick = world.tick

    if should_flee() then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            util.log(self.name .. " fled to regroup")
        end
        return
    end

    if phase == "night" then
        if tick % RITUAL_INTERVAL == 0 then
            do_ritual()
        end
        if tick % ATTACK_INTERVAL == 0 then
            do_attack()
        end
        return
    end

    if phase == "dusk" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            util.log(self.name .. " returned to the cult camp")
        end
    end

    if tick % ATTACK_INTERVAL == 0 then
        do_attack()
    end

    if tick % 50 == 0 then
        util.set_mood("stressed")
    end
end

do_tick()
