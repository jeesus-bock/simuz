-- elf.lua
-- Elven city guard AI: superior fighters defending elven strongholds.
-- Elves are communist in nature — collective defense, shared resources,
-- and coordinated group tactics. Their long lifespan (1500 years) makes
-- them experienced, disciplined warriors who form an impenetrable defensive line.
--
-- Guards block all hostile entry into elven cities.
-- Elves fight with superior skill, coordination, and magical support.

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
    -- Elves share healing and resources with wounded allies
    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.faction == self.faction and info.hp < info.max_hp then
                if world.has_material("herb") then
                    world.give_item(eid, "herb")
                    util.log(self.name .. " shares healing herbs with " .. info.name .. " (collective care)")
                    return
                end
            end
        end
    end
end

local function form_defensive_line()
    -- Elves form an impenetrable defensive line — the strongest hold the center
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
        util.log(self.name .. " forms the elven defensive line with " .. #allies .. " allies")
        util.set_mood("authoritative", 25)
        return true
    end
    return false
end

local function collective_attack()
    -- Elves attack with superior coordination and bonus damage
    local target_id, target_info = find_hostile()
    if not target_id then return false end

    local allies = count_allies()
    local damage_bonus = 1.0
    -- Superior fighters: higher group bonus
    if allies >= 2 then
        damage_bonus = 1.0 + (allies * 0.2)
        util.log(self.name .. " commands the elven assault! +" .. math.floor(damage_bonus * 100 - 100) .. "% group damage")
    end

    local hit = world.attack(self.id, target_id)
    if hit then
        util.log(self.name .. " strikes " .. (target_info.name or target_id) .. " with elven precision!")
        -- Distribute any loot among the collective
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
    -- Protect the weakest elf — elven guards never abandon their own
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

local function magical_support()
    -- Elves provide magical buffs to nearby allies
    local nearby = world.nearby_entities()
    if not nearby then return false end

    local allies_nearby = 0
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.faction == self.faction then
                allies_nearby = allies_nearby + 1
            end
        end
    end

    if allies_nearby >= 2 then
        util.log(self.name .. " weaves a protective enchantment over the elven stronghold")
        util.set_mood("focused", 30)
        return true
    end
    return false
end

local function patrol_elven_territory()
    -- Elves patrol their territory with grace and precision
    local exits = world.exits_from(self.loc_id)
    if exits and #exits > 0 then
        local dest = exits[util.rand_int(#exits) + 1]
        if dest ~= self.loc_id then
            world.move_to(dest)
            util.log(self.name .. " patrols the elven borders")
            return true
        end
    end
    return false
end

function do_tick()
    local phase = world.phase
    local tick = world.tick
    local events = {}

    -- Night: return to the elven stronghold
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

    -- Day: elven city defense and patrol
    -- Priority 1: Defend wounded allies
    if defend_collective() then
        table.insert(events, util.event("attack", {}))
        return events
    end

    -- Priority 2: Share resources with the collective
    if tick % 15 == 0 then
        share_resources()
    end

    -- Priority 3: Form defensive line
    if tick % 10 == 0 then
        form_defensive_line()
    end

    -- Priority 4: Magical support for the stronghold
    if tick % 20 == 0 then
        magical_support()
    end

    -- Priority 5: Collective attack on hostiles
    if tick % 5 == 0 then
        if collective_attack() then
            table.insert(events, util.event("attack", {}))
        end
    end

    -- Priority 6: Patrol elven territory
    if tick % 25 == 0 and util.rand_int(100) < 40 then
        if patrol_elven_territory() then
            table.insert(events, util.event("move", {}))
        end
    end

    return events
end

return do_tick()
