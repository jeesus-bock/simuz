-- Necromancer AI
-- A dark sorcerer who drains life from enemies, stores soul shards, and flees at
-- low health. Raises the dead as a memory token; can empower undead allies nearby.

local ATTACK_INTERVAL = 4
local DRAIN_CHANCE = 50
local FLEE_THRESHOLD = 0.25
local BUFF_INTERVAL = 40

local function count_undead()
    local nearby = world.nearby_entities()
    if not nearby then return 0 end
    local n = 0
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.faction == "undead" then
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
            if info.faction ~= "undead" and info.faction ~= "cultist" then
                return eid, info
            end
        end
        ::continue::
    end
    return nil
end

local function do_life_drain(target_id)
    local hit = world.attack(self.id, target_id)
    if hit then
        local heal = math.floor(self.max_hp * 0.1)
        util.log(self.name .. " drained life and healed " .. heal .. " HP")
    end
    return hit
end

local function collect_soul()
    local souls = util.mem_get("soul_shards") or 0
    util.mem_set("soul_shards", souls + 1)
    util.log(self.name .. " collected a soul shard")
end

local function raise_dead_token()
    local souls = util.mem_get("soul_shards") or 0
    if souls >= 3 then
        util.mem_set("soul_shards", souls - 3)
        util.mem_set("skeleton_count", (util.mem_get("skeleton_count") or 0) + 1)
        util.log(self.name .. " raised a skeleton from the dead")
        return true
    end
    return false
end

local function should_flee()
    return self.hp / self.max_hp < FLEE_THRESHOLD
end

local function do_tick()
    local phase = world.phase
    local tick = world.tick

    if should_flee() then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            util.log(self.name .. " retreated to the graveyard")
        end
        return
    end

    if phase == "night" then
        if tick % 20 == 0 then
            raise_dead_token()
        end
    end

    if tick % BUFF_INTERVAL == 0 then
        local undead = count_undead()
        if undead > 0 then
            util.log(self.name .. " empowered " .. undead .. " undead ally(ies)")
        end
    end

    if tick % ATTACK_INTERVAL == 0 then
        local target_id, target_info = find_target()
        if target_id then
            if util.rand_int(100) < DRAIN_CHANCE then
                do_life_drain(target_id)
                if not target_info.alive then
                    collect_soul()
                end
            end
        end
    end

    if tick % 60 == 0 then
        local roll = util.rand_int(100)
        if roll < 30 then
            util.set_mood("inspired")
        elseif roll < 60 then
            util.set_mood("stressed")
        end
    end
end

do_tick()
