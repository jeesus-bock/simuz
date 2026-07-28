-- Healing AI
-- Heals injured allies and creatures nearby.
-- Stays near a sanctuary or home location.

local HEAL_INTERVAL = 10
local HEAL_AMOUNT = 5

local function find_injured()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return nil
    end
    local best = nil
    local worst_hp_pct = 1.0
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.hp < info.max_hp and info.species ~= "deity" then
            if not world.is_hostile(self.faction, info.faction) then
                local hp_pct = info.hp / info.max_hp
                if hp_pct < worst_hp_pct then
                    worst_hp_pct = hp_pct
                    best = eid
                end
            end
        end
    end
    return best
end

local function do_heal()
    local target_id = find_injured()
    if target_id then
        local target = world.entity_name(target_id)
        if world.attack then
            -- Use attack as a proxy for healing (healing done via divine_intervention from deity)
            -- In this context, healing just logs the action
            util.log(self.name .. " tended to " .. (target or target_id))
        end
    end
end

local function do_tick()
    local tick = world.tick
    local phase = world.phase

    if phase == "night" or phase == "dusk" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
        end
        return
    end

    if tick % HEAL_INTERVAL == 0 then
        do_heal()
    end

    if self.home and self.loc_id ~= self.home and tick % 30 == 0 then
        world.move_to(self.home)
    end
end

do_tick()
