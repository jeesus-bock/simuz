-- Berzerker profession AI script
-- Permanent +5-10 STR/DEX bonus (applied at entity creation)
-- Can attack friendlies when in psychotic rage

local RAGE_DURATION = 300
local RAGE_HP_THRESHOLD = 0.3

local function is_raging(eid)
    local rage_until = util.mem_get("rage_until")
    if rage_until == nil then return false end
    return world.tick < rage_until
end

local function enter_rage(eid)
    local duration = RAGE_DURATION + util.rand_int(101) - 50
    if duration < 100 then duration = 100 end
    util.mem_set("rage_until", world.tick + duration)
end

local function find_target(eid)
    local nearby = world.nearby_entities()
    if nearby == nil then return nil end

    local target = nil
    if is_raging(eid) then
        -- In rage, attack any living entity including friendlies
        for _, other_id in ipairs(nearby) do
            if other_id ~= eid then
                local other = world.entity_info(other_id)
                if other ~= nil and other.alive then
                    target = other_id
                    break
                end
            end
        end
    else
        -- Attack hostiles only
        for _, other_id in ipairs(nearby) do
            if other_id ~= eid then
                local other = world.entity_info(other_id)
                if other ~= nil and other.alive and world.is_hostile(eid, other_id) then
                    target = other_id
                    break
                end
            end
        end
    end

    return target
end

local function coordinate_attack(eid)
    local entity = world.entity_info(eid)
    if entity == nil then return end

    -- Enter rage when HP is low
    if not is_raging(eid) and entity.hp < entity.max_hp * RAGE_HP_THRESHOLD then
        enter_rage(eid)
    end

    local target = find_target(eid)
    if target then
        world.attack(eid, target)
    end
end

return coordinate_attack
