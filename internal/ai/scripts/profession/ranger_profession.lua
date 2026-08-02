-- Ranger profession script: hunts prey, tracks targets via relationships,
-- gathers resources, and avoids unnecessary combat. Demonstrates profession-based
-- behavior and relationship tracking for hunting targets.

local function do_tick()
    local myProfession = self.profession
    local myFaction = self.faction

    -- Ensure we have the ranger profession
    if myProfession ~= "ranger" then
        self.set_profession("ranger")
        myProfession = "ranger"
    end

    -- Find nearby entities
    local nearby = world.nearby_entities()
    local prey = {}
    local hostiles = {}
    local neutral = {}

    for _, id in ipairs(nearby) do
        if id ~= self.id then
            local info = world.entity_info(id)
            if info and info.alive and info.conscious then
                local theirFaction = info.faction
                local theirProf = info.profession

                -- Hostile factions or aggressive professions = threats
                if world.is_hostile(myFaction, theirFaction) or theirProf == "bandit" then
                    table.insert(hostiles, id)
                -- Non-hostile, non-civilian = potential prey or neutral
                elseif theirFaction ~= "civilian" and theirFaction ~= "merchant" and theirFaction ~= "" then
                    table.insert(prey, id)
                else
                    table.insert(neutral, id)
                end
            end
        end
    end

    -- Priority 1: Deal with hostiles (defend)
    if #hostiles > 0 then
        local target = hostiles[1]
        for _, id in ipairs(hostiles) do
            local info = world.entity_info(id)
            local targetInfo = world.entity_info(target)
            if info and targetInfo and info.hp < targetInfo.hp then
                target = id
            end
        end

        -- Check if we have a relationship with this hostile
        if self.has_relationship(target) then
            local relType = self.get_relationship_type(target)
            if relType == "love" then
                -- We love this entity, try to flee instead
                util.log(self.name .. " loves " .. world.entity_name(target) .. " — fleeing")
                world.avoid_combat()
                return {}
            elseif relType == "rival" then
                util.log(self.name .. " is attacking rival " .. world.entity_name(target))
            end
        end

        -- Attack the hostile
        world.attack(self.id, target)
        -- Track the hostile as a rival
        if not self.has_relationship(target) then
            self.add_relationship(target, "rival", world.tick)
            util.log(self.name .. " marked " .. world.entity_name(target) .. " as rival")
        end
        return {}
    end

    -- Priority 2: Hunt prey
    if #prey > 0 then
        local target = prey[1]
        -- Prefer targets we haven't hunted yet
        for _, id in ipairs(prey) do
            if not self.has_relationship(id) then
                target = id
                break
            end
        end

        -- Track the prey via relationship
        if not self.has_relationship(target) then
            self.add_relationship(target, "rival", world.tick)
            util.log(self.name .. " is tracking " .. world.entity_name(target) .. " — added rival")
        end

        -- Attack the prey
        world.attack(self.id, target)

        -- If we killed the prey, check for loot
        local targetInfo = world.entity_info(target)
        if targetInfo and not targetInfo.alive then
            util.log(self.name .. " hunted " .. world.entity_name(target) .. " — prey down")
            -- Remove the rival relationship since the target is dead
            self.remove_relationship(target)
        end
        return {}
    end

    -- Priority 3: Gather resources if nothing else to do
    if #neutral > 0 then
        -- Try to add a useful item to inventory
        local gathered = world.add_item("herb")
        if gathered then
            util.log(self.name .. " gathered herbs")
        end
    end

    -- Avoid unnecessary combat
    world.avoid_combat()
    return {}
end

return do_tick()
