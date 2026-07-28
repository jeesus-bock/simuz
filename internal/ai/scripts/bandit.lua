-- Bandit profession script: steals, raids, and tracks victims via relationships.
-- Uses self.profession to confirm role, world.get_profession/world.set_profession
-- for manipulating others, and relationships to track rivals and stolen goods.

local function is_hostile(factionA, factionB)
    return world.is_hostile(factionA, factionB)
end

local function get_relation(factionA, factionB)
    return world.get_relation(factionA, factionB)
end

-- Main tick logic
local function do_tick()
    local myFaction = self.faction
    local myProfession = self.profession

    -- Confirm we are a bandit
    if myProfession ~= "bandit" then
        -- Try to set our profession if we are a human in bandit territory
        local loc = world.location_name(self.loc_id)
        if loc and string.find(loc, "bandit") then
            self.set_profession("bandit")
            myProfession = "bandit"
        end
    end

    -- Find nearby entities
    local nearby = world.nearby_entities()
    local hostiles = {}
    local victims = {}
    local allies = {}

    for _, id in ipairs(nearby) do
        local info = world.entity_info(id)
        if info and info.alive and info.conscious then
            local theirFaction = info.faction
            local theirProf = info.profession

            -- Bandits are hostile to civilians and merchants
            if theirFaction == "civilian" or theirFaction == "merchant" or theirProf == "merchant" then
                table.insert(victims, id)
            elseif is_hostile(myFaction, theirFaction) or theirProf == "bandit" then
                -- Other bandits are rivals; hostile factions are enemies
                if theirProf == "bandit" then
                    table.insert(allies, id)
                else
                    table.insert(hostiles, id)
                end
            end
        end
    end

    -- If we have hostiles, attack the weakest one
    if #hostiles > 0 then
        local target = hostiles[1]
        -- Pick the one with the lowest HP
        for _, id in ipairs(hostiles) do
            local info = world.entity_info(id)
            local targetInfo = world.entity_info(target)
            if info and targetInfo and info.hp < targetInfo.hp then
                target = id
            end
        end

        -- Check if we are related to the target (rival relationship)
        local relType = self.get_relationship_type(target)
        if relType == "rival" then
            util.log(self.name .. " is attacking rival " .. world.entity_name(target))
        end

        -- Attack
        local hit = world.attack(self.id, target)
        if hit then
            -- Add a rival relationship with whoever we hit
            self.add_relationship(target, "rival", world.tick)
            util.log(self.name .. " hit " .. world.entity_name(target) .. " — marked as rival")
        end

        -- If we killed them, loot and update faction relations
        local targetInfo = world.entity_info(target)
        if targetInfo and not targetInfo.alive then
            -- Shift faction relation toward the victim's faction
            if targetInfo.faction and targetInfo.faction ~= myFaction then
                world.set_relation(myFaction, targetInfo.faction, "hostile")
            end
            -- Add a rival relationship with the deceased
            self.add_relationship(target, "rival", world.tick)
            util.log(self.name .. " killed " .. world.entity_name(target) .. " — added rival")
        end

        return
    end

    -- If we have victims, try to steal from them
    if #victims > 0 then
        local target = victims[1]
        -- Prefer targets we already have a relationship with (repeat victims)
        for _, id in ipairs(victims) do
            if self.has_relationship(id) then
                target = id
                break
            end
        end

        -- Try to steal a valuable item
        local targetInfo = world.entity_info(target)
        if targetInfo then
            local theirItems = world.entity_items(target)
            if theirItems and #theirItems > 0 then
                -- Pick the first non-currency item
                local itemToSteal = nil
                for _, itemDefID in ipairs(theirItems) do
                    if itemDefID ~= "gp" then
                        itemToSteal = itemDefID
                        break
                    end
                end

                if itemToSteal then
                    local success = world.steal(target, itemToSteal)
                    if success then
                        -- Track the victim via a relationship
                        if not self.has_relationship(target) then
                            self.add_relationship(target, "rival", world.tick)
                            util.log(self.name .. " stole from " .. world.entity_name(target) .. " — added rival")
                        end

                        -- Update faction relations
                        if targetInfo.faction and targetInfo.faction ~= myFaction then
                            world.set_relation(myFaction, targetInfo.faction, "hostile")
                        end
                        return
                    end
                end
            end
        end

        -- If steal failed, try to attack instead
        world.attack(self.id, target)
        return
    end

    -- If we have allies (other bandits), coordinate
    if #allies > 0 then
        -- Check if any ally is in combat nearby
        for _, allyID in ipairs(allies) do
            local allyInfo = world.entity_info(allyID)
            if allyInfo and allyInfo.hp < allyInfo.max_hp then
                -- Ally is hurt, help them
                util.log(self.name .. " helping ally " .. world.entity_name(allyID))
                -- Move toward the ally
                world.move_to(allyInfo.location_id)
                return
            end
        end
    end

    -- If nothing to do, wander or avoid combat
    if not world.avoid_combat() then
        -- Try to find a target to provoke
        for _, id in ipairs(nearby) do
            local info = world.entity_info(id)
            if info and info.alive and info.conscious and info.faction ~= myFaction then
                -- Provoke by attacking
                world.attack(self.id, id)
                return
            end
        end
    end
end

do_tick()
