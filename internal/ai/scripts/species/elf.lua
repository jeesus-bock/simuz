-- elf.lua
-- Elven city guard AI: superior traders and guards protecting elven commercial interests.
-- Elves are capitalist in nature — individual wealth accumulation, trade dominance,
-- and mercenary-style defense for profit. Their long lifespan (1500 years) makes
-- them experienced, disciplined merchants and guards who dominate trade routes.
--
-- Guards block all hostile entry into elven cities to protect trade.
-- Elves fight with superior skill and precision, motivated by profit and commerce.

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

local function find_weakest_ally()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return nil
    end
    local weakest = nil
    local weakest_hp_pct = 1.0
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.faction == self.faction then
                local hp_pct = info.hp / info.max_hp
                if hp_pct < weakest_hp_pct then
                    weakest_hp_pct = hp_pct
                    weakest = eid
                end
            end
        end
    end
    return weakest
end

local function find_strongest_ally()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return nil
    end
    local strongest = nil
    local strongest_hp_pct = 0.0
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.faction == self.faction then
                local hp_pct = info.hp / info.max_hp
                if hp_pct > strongest_hp_pct then
                    strongest_hp_pct = hp_pct
                    strongest = eid
                end
            end
        end
    end
    return strongest
end

local function assess_trade_threat()
    -- Elves evaluate threats to trade routes and commercial interests
    local nearby = world.nearby_entities()
    if not nearby then return nil end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and world.is_hostile(self.faction, info.faction) then
                return eid, info
            end
        end
    end
    return nil
end

local function protect_trade_route()
    -- Elves guard trade routes for profit and commercial dominance
    local exits = world.exits_from(self.loc_id)
    if exits and #exits > 0 then
        local dest = exits[util.rand_int(#exits) + 1]
        if dest ~= self.loc_id then
            world.move_to(dest)
            util.log(self.name .. " patrols the elven trade route")
            return true
        end
    end
    return false
end

local function mercenary_defense()
    -- Elves defend for profit — they are paid to protect elven commerce
    local target_id, target_info = find_hostile()
    if not target_id then return false end

    local hit = world.attack(self.id, target_id)
    if hit then
        util.log(self.name .. " strikes " .. (target_info.name or target_id) .. " with elven precision — for coin and commerce")
    end
    return true
end

local function collect_tribute()
    -- Elves demand tribute from weaker factions — capitalist dominance
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.faction ~= self.faction then
                if info.hp < info.max_hp * 0.3 then
                    util.log(self.name .. " demands tribute from " .. (info.name or eid))
                    util.set_mood("greedy", 20)
                    return true
                end
            end
        end
    end
    return false
end

local function hoard_wealth()
    -- Elves accumulate wealth and valuable items for personal gain
    local items = world.entity_items(self.id)
    if items and #items > 0 then
        local valuable = items[util.rand_int(#items) + 1]
        util.log(self.name .. " secures " .. valuable .. " for personal wealth")
        util.set_mood("avaricious", 15)
        return true
    end
    return false
end

local function patrol_elven_commerce()
    -- Elves patrol their commercial territories with grace and precision
    local exits = world.exits_from(self.loc_id)
    if exits and #exits > 0 then
        local dest = exits[util.rand_int(#exits) + 1]
        if dest ~= self.loc_id then
            world.move_to(dest)
            util.log(self.name .. " patrols the elven commercial borders")
            return true
        end
    end
    return false
end

function do_tick()
    local phase = world.phase
    local tick = world.tick
    local events = {}

    -- Night: return to the elven stronghold to count profits
    if phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            return {util.event("move", {})}
        end
        -- Hoard wealth at night
        if tick % 10 == 0 then
            hoard_wealth()
        end
        return {}
    end

    -- Day: elven city defense and trade protection
    -- Priority 1: Protect trade routes from hostiles
    if tick % 5 == 0 then
        local threat = assess_trade_threat()
        if threat then
            if mercenary_defense() then
                table.insert(events, util.event("attack", {}))
                return events
            end
        end
    end

    -- Priority 2: Demand tribute from weaker factions
    if tick % 15 == 0 then
        if collect_tribute() then
            table.insert(events, util.event("trade", {}))
        end
    end

    -- Priority 3: Hoard wealth and valuable items
    if tick % 10 == 0 then
        hoard_wealth()
    end

    -- Priority 4: Patrol elven commercial territory
    if tick % 25 == 0 and util.rand_int(100) < 40 then
        if patrol_elven_commerce() then
            table.insert(events, util.event("move", {}))
        end
    end

    -- Priority 5: Mercenary defense on hostiles
    if tick % 8 == 0 then
        if mercenary_defense() then
            table.insert(events, util.event("attack", {}))
        end
    end

    return events
end

return do_tick()
