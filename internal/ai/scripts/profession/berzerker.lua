-- Berzerker profession AI script
-- Permanent +5-10 STR/DEX bonus (applied at entity creation via ProfessionBonus)
-- Can attack friendlies when in psychotic rage
-- Rage mechanics: enters rage at low HP, deals bonus damage, attacks any living target
-- After rage ends, enters a brief vulnerable cooldown period

local RAGE_DURATION_MIN = 200
local RAGE_DURATION_MAX = 400
local RAGE_HP_THRESHOLD = 0.35
local RAGE_DAMAGE_BONUS = 0.25
local RAGE_COOLDOWN = 60
local POST_RAGE_VULN = 30

-- State tracking keys
local RAGE_UNTIL_KEY = "rage_until"
local RAGE_COOLDOWN_KEY = "rage_cooldown"
local POST_RAGE_VULN_KEY = "post_rage_vuln"
local KILLS_IN_RAGE_KEY = "kills_in_rage"

local function is_raging(eid)
    local rage_until = util.mem_get(RAGE_UNTIL_KEY)
    if rage_until == nil then return false end
    return world.tick < rage_until
end

local function is_cooldown(eid)
    local cooldown_until = util.mem_get(RAGE_COOLDOWN_KEY)
    if cooldown_until == nil then return false end
    return world.tick < cooldown_until
end

local function is_post_rage_vulnerable(eid)
    local vuln_until = util.mem_get(POST_RAGE_VULN_KEY)
    if vuln_until == nil then return false end
    return world.tick < vuln_until
end

local function enter_rage(eid)
    local duration = RAGE_DURATION_MIN + util.rand_int(RAGE_DURATION_MAX - RAGE_DURATION_MIN + 1)
    util.mem_set(RAGE_UNTIL_KEY, world.tick + duration)
    util.mem_set(KILLS_IN_RAGE_KEY, 0)
    util.log(self.name .. " enters a psychotic rage! Eyes bloodshot, muscles bulging.")
    util.set_mood("furious", duration)
end

local function exit_rage(eid)
    util.mem_set(RAGE_UNTIL_KEY, nil)
    -- Enter cooldown so rage doesn't retrigger immediately
    util.mem_set(RAGE_COOLDOWN_KEY, world.tick + RAGE_COOLDOWN)
    -- Post-rage vulnerability: reduced defense for a short time
    util.mem_set(POST_RAGE_VULN_KEY, world.tick + POST_RAGE_VULN)
    local kills = tonumber(util.mem_get(KILLS_IN_RAGE_KEY) or "0")
    util.log(self.name .. " calms down from rage. " .. kills .. " kills in this episode.")
    if kills >= 3 then
        util.log(self.name .. " is exhilarated by the slaughter! Adrenaline surge.")
        util.set_mood("ecstatic", 20)
    else
        util.set_mood("tired", 15)
    end
end

local function get_rage_damage_bonus(eid)
    if not is_raging(eid) then return 0 end
    return RAGE_DAMAGE_BONUS
end

local function find_target(eid)
    local nearby = world.nearby_entities()
    if nearby == nil then return nil end

    local target = nil

    if is_raging(eid) then
        -- In rage, attack any living entity including friendlies
        -- Prioritize: enemies first (most dangerous), then neutrals, then friendlies
        local enemies = {}
        local neutrals = {}
        local friendlies = {}

        for _, other_id in ipairs(nearby) do
            if other_id ~= eid then
                local other = world.entity_info(other_id)
                if other ~= nil and other.alive then
                    if world.is_hostile(eid, other_id) then
                        table.insert(enemies, other_id)
                    elseif other.faction == "" or other.faction == "civilian" then
                        table.insert(neutrals, other_id)
                    else
                        table.insert(friendlies, other_id)
                    end
                end
            end
        end

        -- Prefer enemies, then neutrals, then friendlies
        local candidates = enemies
        if #candidates == 0 then candidates = neutrals end
        if #candidates == 0 then candidates = friendlies end

        if #candidates > 0 then
            -- Pick the weakest target (lowest HP%) to maximize rage efficiency
            local weakest = candidates[1]
            local weakest_hp_pct = 1.0
            for _, cid in ipairs(candidates) do
                local cinfo = world.entity_info(cid)
                if cinfo and cinfo.max_hp > 0 then
                    local hp_pct = cinfo.hp / cinfo.max_hp
                    if hp_pct < weakest_hp_pct then
                        weakest_hp_pct = hp_pct
                        weakest = cid
                    end
                end
            end
            target = weakest
        end
    else
        -- Outside rage: attack hostiles only, prefer weakest
        for _, other_id in ipairs(nearby) do
            if other_id ~= eid then
                local other = world.entity_info(other_id)
                if other ~= nil and other.alive and world.is_hostile(eid, other_id) then
                    if target == nil then
                        target = other_id
                    else
                        -- Prefer the target with lowest HP%
                        local tinfo = world.entity_info(target)
                        local oinfo = world.entity_info(other_id)
                        if tinfo and oinfo and tinfo.max_hp > 0 and oinfo.max_hp > 0 then
                            if (oinfo.hp / oinfo.max_hp) < (tinfo.hp / tinfo.max_hp) then
                                target = other_id
                            end
                        end
                    end
                end
            end
        end
    end

    return target
end

local function apply_rage_bonus(eid, target_id)
    if not is_raging(eid) then return false end

    local entity = world.entity_info(eid)
    local target = world.entity_info(target_id)
    if entity == nil or target == nil then return false end

    -- Bonus damage: 25% extra during rage
    local bonus_dmg = math.floor(entity.max_hp * RAGE_DAMAGE_BONUS)
    if bonus_dmg < 5 then bonus_dmg = 5 end

    -- Apply the bonus damage via a second attack call
    -- The first attack is handled by world.attack, so we log the bonus here
    util.log(self.name .. " attacks " .. (target.name or target_id) .. " with rage-fueled fury! +" .. bonus_dmg .. " bonus damage.")

    return true
end

local function track_rage_kill(eid)
    if not is_raging(eid) then return end
    local kills = tonumber(util.mem_get(KILLS_IN_RAGE_KEY) or "0")
    kills = kills + 1
    util.mem_set(KILLS_IN_RAGE_KEY, kills)
end

local function coordinate_attack(eid)
    local entity = world.entity_info(eid)
    if entity == nil then return end

    -- Check if rage has expired
    if is_raging(eid) and not is_cooldown(eid) then
        local rage_until = tonumber(util.mem_get(RAGE_UNTIL_KEY) or "0")
        if world.tick >= rage_until then
            exit_rage(eid)
        end
    end

    -- Check if cooldown has expired and post-rage vulnerability is over
    if is_cooldown(eid) then
        local cooldown_until = tonumber(util.mem_get(RAGE_COOLDOWN_KEY) or "0")
        if world.tick >= cooldown_until then
            util.mem_set(RAGE_COOLDOWN_KEY, nil)
            util.mem_set(POST_RAGE_VULN_KEY, nil)
            util.log(self.name .. " has recovered from the rage.")
        end
    end

    -- Enter rage when HP is low and not already raging or cooling down
    if not is_raging(eid) and not is_cooldown(eid) and entity.hp < entity.max_hp * RAGE_HP_THRESHOLD then
        enter_rage(eid)
    end

    -- If in post-rage vulnerability, reduce attack frequency
    local attack_chance = 100
    if is_post_rage_vulnerable(eid) then
        attack_chance = 50 -- Only attack half the time when vulnerable
    end

    local target = find_target(eid)
    if target then
        if util.rand_int(100) < attack_chance then
            local hit = world.attack(eid, target)
            if hit then
                apply_rage_bonus(eid, target)
                track_rage_kill(eid)
            end
        end
    end
end

return coordinate_attack
