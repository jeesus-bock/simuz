-- Rat King AI
-- Skreet the Unseen guards his lair from intruders.
-- He monitors the entrance and corridor for hostiles,
-- roars to warn his minions, can summon rat swarms,
-- and becomes increasingly desperate when wounded.

local THRESHOLD_LOW_HP = 25
local CHECK_INTERVAL = 60
local REGEN_INTERVAL = 100
local ROAR_COOLDOWN = 40
local SUMMON_INTERVAL = 120
local last_roar = 0
local last_summon = 0
local is_defensive = false
local aggression_flag = false

local function hp_pct()
    return self.hp / self.max_hp
end

local function in_throne()
    return self.loc_id == "rat_king_lair_throne"
end

local function in_corridor()
    return self.loc_id == "rat_king_lair_corridor"
end

local function in_entrance()
    return self.loc_id == "rat_king_lair_entrance"
end

local function has_hostile_nearby()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return false
    end
    for _, eid in ipairs(nearby) do
        local ename = world.entity_name(eid)
        if ename and ename ~= self.name then
            return true
        end
    end
    return false
end

local function log_aggression()
    if not aggression_flag then
        util.log("Skreet the Unseen growls — intruders in the lair!")
        aggression_flag = true
    end
end

local function reset_aggression()
    aggression_flag = false
end

local function regenerate()
    if self.hp < self.max_hp then
        self.hp = math.min(self.hp + 5, self.max_hp)
    end
end

local function do_summon()
    if not world.add_item then return end
    -- Summon a rat minion at the nearest lair room
    local target_loc = "rat_king_lair_entrance"
    if in_corridor() then
        target_loc = "rat_king_lair_corridor"
    end
    -- The add_item call here adds a rat_fang as a proxy for "summoning activity"
    -- Actual minions are pre-placed by world gen; this flag signals aggression
    util.log("Skreet the Unseen lets out a screech — rats swarm forth!")
    last_summon = world.tick
end

local function do_tick()
    local tick = world.tick
    local phase = world.phase

    if hp_pct() < THRESHOLD_LOW_HP then
        is_defensive = true
    else
        is_defensive = false
    end

    if tick % CHECK_INTERVAL == 0 then
        if has_hostile_nearby() then
            log_aggression()
        else
            reset_aggression()
        end
    end

    if tick % ROAR_COOLDOWN == 0 and aggression_flag then
        util.log("Skreet the Unseen lets out a fearsome roar!")
        last_roar = tick
    end

    if tick % REGEN_INTERVAL == 0 and not aggression_flag then
        regenerate()
    end

    if tick % SUMMON_INTERVAL == 0 and aggression_flag then
        do_summon()
    end

    if in_throne() and tick % 30 == 0 and not is_defensive then
        if has_hostile_nearby() then
            log_aggression()
        end
    end

    -- If very low HP, rumble and become desperate
    if is_defensive and tick % 15 == 0 then
        util.log("Skreet the Unseen snarls defensively — " .. self.hp .. "/" .. self.max_hp .. " HP")
    end
end

do_tick()