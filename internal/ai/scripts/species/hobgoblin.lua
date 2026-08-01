-- hobgoblin.lua
-- Communist hobgoblin AI: disciplined collective warfare,
-- chain of command, shared resources, and organized defense.
-- Hobgoblins are the most organized of the communist greenskins.

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

local function count_allies()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return 0
    end
    local count = 0
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.faction == self.faction then
                count = count + 1
            end
        end
    end
    return count
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

local function share_resources()
    -- Redistribute food and materials to wounded allies
    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.faction == self.faction and info.hp < info.max_hp then
                if world.has_material("raw_meat") then
                    world.give_item(eid, "raw_meat")
                    util.log(self.name .. " shares provisions with " .. info.name .. " (collective ration)")
                    return
                end
            end
        end
    end
end

local function coordinate_line()
    -- Form a defensive line with the strongest allies
    local nearby = world.nearby_entities()
    if not nearby then return false end

    local allies = {}
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.faction == self.faction then
                table.insert(allies, eid)
            end
        end
    end

    if #allies >= 2 then
        util.log(self.name .. " organizes the collective line — " .. #allies .. " allies in formation")
        util.set_mood("authoritative", 20)
        return true
    end
    return false
end

local function collective_attack()
    local target_id, target_info = find_hostile()
    if not target_id then return false end

    -- Disciplined group attack with bonus for numbers
    local allies = count_allies()
    local damage_bonus = 1.0
    if allies >= 2 then
        damage_bonus = 1.0 + (allies * 0.12)
        util.log(self.name .. " commands the collective assault! +" .. math.floor(damage_bonus * 100 - 100) .. "% group bonus")
    end

    local hit = world.attack(self.id, target_id)
    if hit then
        util.log(self.name .. " strikes " .. (target_info.name or target_id) .. " for the collective!")
        -- Distribute loot among allies
        if target_info and not target_info.alive then
            local items = world.entity_items(target_id)
            if items and #items > 0 then
                for _, ally_id in ipairs(nearby) do
                    if ally_id ~= self.id then
                        local ally_info = world.entity_info(ally_id)
                        if ally_info and ally_info.alive and ally_info.faction == self.faction then
                            local item = items[util.rand_int(#items) + 1]
                            world.give_item(ally_id, item)
                            util.log(self.name .. " distributes " .. item .. " to " .. ally_info.name .. " (collective property)")
                            break
                        end
                    end
                end
            end
        end
    end
    return true
end

local function defend_collective()
    -- Protect the weakest hobgoblin first, then the strongest holds the line
    local weakest = find_weakest_ally()
    if weakest then
        local info = world.entity_info(weakest)
        if info and info.hp < info.max_hp * 0.5 then
            local nearby = world.nearby_entities()
            if nearby then
                for _, eid in ipairs(nearby) do
                    if eid ~= self.id then
                        local enemy = world.entity_info(eid)
                        if enemy and enemy.alive and world.is_hostile(self.faction, enemy.faction) then
                            util.log(self.name .. " shields " .. info.name .. " from " .. (enemy.name or eid))
                            world.attack(self.id, eid)
                            return true
                        end
                    end
                end
            end
        end
    end
    return false
end

function do_tick()
    local phase = world.phase
    local tick = world.tick
    local events = {}

    -- Night: return to the fortified camp
    if phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            return {util.event("move", {})}
        end
        -- Share resources and rest with the collective
        if tick % 10 == 0 then
            share_resources()
        end
        return {}
    end

    -- Day: disciplined collective operations
    -- Priority 1: Defend wounded allies
    if defend_collective() then
        table.insert(events, util.event("attack", {}))
        return events
    end

    -- Priority 2: Share resources with the collective
    if tick % 15 == 0 then
        share_resources()
    end

    -- Priority 3: Coordinate defensive line
    if tick % 20 == 0 then
        coordinate_line()
    end

    -- Priority 4: Disciplined collective attack
    if tick % 5 == 0 then
        if collective_attack() then
            table.insert(events, util.event("attack", {}))
        end
    end

    -- Priority 5: Patrol in formation
    if tick % 25 == 0 and util.rand_int(100) < 40 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            if dest ~= self.loc_id then
                world.move_to(dest)
                table.insert(events, util.event("move", {}))
            end
        end
    end

    return events
end

return do_tick()
