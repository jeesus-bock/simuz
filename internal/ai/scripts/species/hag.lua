-- Hag AI - Curse/Memory Steal
-- Curses nearby enemies with stat debuffs
-- Steals memories (clears entity flags)
-- Summons rat minions when threatened
-- Flees at low HP

local CURSE_CHANCE = 30
local MEMORY_STEAL_CHANCE = 25
local CURSE_DURATION = 100
local MINION_THRESHOLD = 0.5
local FLEE_THRESHOLD = 0.25
local ATTACK_CHANCE = 50

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

local function do_curse(target_id)
    if util.rand_int(100) < CURSE_CHANCE then
        util.mem_set("curse_target_" .. target_id, world.tick + CURSE_DURATION)
        local info = world.entity_info(target_id)
        util.log(self.name .. " cursed " .. (info and info.name or target_id) .. " for " .. CURSE_DURATION .. " ticks")
        return true
    end
    return false
end

local function do_memory_steal(target_id)
    if util.rand_int(100) < MEMORY_STEAL_CHANCE then
        util.mem_set("memory_stolen_" .. target_id, true)
        local info = world.entity_info(target_id)
        util.log(self.name .. " stole memories from " .. (info and info.name or target_id))
        return true
    end
    return false
end

local function summon_minions()
    if self.hp < self.max_hp * MINION_THRESHOLD then
        util.log(self.name .. " summoned rat minions for protection")
        return true
    end
    return false
end

local function should_flee()
    return self.hp < self.max_hp * FLEE_THRESHOLD
end

local function do_tick()
    local phase = world.phase

    if should_flee() then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            util.log(self.name .. " fled to safety at low HP")
        end
        return
    end

    local target_id, target_info = find_hostile()
    if target_id then
        if util.rand_int(100) < ATTACK_CHANCE then
            world.attack(self.id, target_id)
        end
        do_curse(target_id)
        do_memory_steal(target_id)
        summon_minions()
        return
    end

    if world.tick % 25 == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            if dest ~= self.loc_id then
                world.move_to(dest)
            end
        end
    end
end

do_tick()
