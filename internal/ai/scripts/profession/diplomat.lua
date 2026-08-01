-- diplomat.lua
-- Machiavellian Statecraft Script: Realpolitik negotiation,
-- calculated alliance shifts, and strategic destabilization.
-- Diplomats carry diplomatic immunity and seek out politicians.

local function find_politicians()
    local nearby = world.nearby_entities()
    if not nearby then return {} end
    local politicians = {}
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.profession == "politician" then
                table.insert(politicians, info)
            end
        end
    end
    return politicians
end

local function negotiate_with_politician(pol)
    if world.can_communicate(pol.id) then
        world.say_to(pol.id, "I bring propositions from distant lands. Shall we discuss terms?")
        util.log(self.name .. " engages politician " .. pol.name .. " in formal diplomacy.")
        util.set_mood("diplomatic", 25)
        return true
    else
        util.log(self.name .. " presents written credentials to " .. pol.name .. ".")
        util.set_mood("formal", 15)
        return true
    end
end

local function practice_statecraft()
    local nearby = world.nearby_entities()
    if not nearby then return nil, nil end

    -- Realpolitik: "He who adapts his policy to the times prospers."
    -- If critically wounded, retreat immediately. Preservation of the state (self) is paramount.
    if self.hp and self.hp < 45 then
        util.log(self.name .. " calculates poor odds. 'A prince must know how to escape danger when necessary.'")
        util.set_mood("calculating", 15)
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then world.move_to(exits[util.rand_int(#exits) + 1]) end
        return nil, nil
    end

    local strongest_entity = nil
    local weakest_entity = nil
    local max_hp = -1
    local min_hp = 9999

    -- 1. Scan the room to map the local balance of power
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                -- Skip diplomats and politicians for power assessment
                if info.profession ~= "diplomat" and info.profession ~= "politician" then
                    if info.hp > max_hp then
                        max_hp = info.hp
                        strongest_entity = info
                    end
                    if info.hp < min_hp then
                        min_hp = info.hp
                        weakest_entity = info
                    end
                end
            end
        end
    end

    return strongest_entity, weakest_entity
end

local function negotiate(strongest, weakest)
    if strongest and strongest.id ~= self.id then
        -- Flattery: Defer to the strongest entity present
        if world.can_communicate(strongest.id) then
            world.say_to(strongest.id, "Your strength precedes you. Perhaps we might find mutual benefit.")
        end
        util.log(self.name .. " addresses " .. strongest.name .. " with calculated deference.")
        util.set_mood("diplomatic", 20)
        return true
    end

    if weakest and weakest.id ~= self.id and weakest.faction ~= self.faction then
        -- Destabilization: Seed distrust in the weakest non-aligned entity
        if world.can_communicate(weakest.id) then
            world.say_to(weakest.id, "I hear troubling whispers about your rivals...")
        end
        util.log(self.name .. " whispers veiled threats to " .. weakest.name .. ".")
        util.set_mood("scheming", 20)
        return true
    end

    return false
end

function do_tick()
    local acted = false

    if world.tick % 20 == 0 then
        -- Priority 1: Seek out politicians for formal diplomacy
        local politicians = find_politicians()
        if #politicians > 0 then
            acted = negotiate_with_politician(politicians[1])
        end

        -- Priority 2: General statecraft with non-politician entities
        if not acted then
            local strongest, weakest = practice_statecraft()
            if strongest and type(strongest) == "table" then
                acted = negotiate(strongest, weakest)
            end
        end

        if not acted then
            -- If no targets, travel to a new location to scout
            local exits = world.exits_from(self.loc_id)
            if exits and #exits > 0 then
                local dest = exits[util.rand_int(#exits) + 1]
                if dest ~= self.loc_id then
                    util.log(self.name .. " leaves to sow regional stability—or profitable discord.")
                    world.move_to(dest)
                    acted = true
                end
            end
        end
    end

    if acted then
        return {util.event("profession_action", {profession = "diplomat"})}
    end
    return {}
end

return do_tick()
