-- Faction loyalist script: demonstrates faction-based loyalty, relationship
-- management for allies and enemies, and profession switching for faction
-- members. Shows how factions (cults, religions, political movements) work
-- alongside the profession system.

local function do_tick()
    local myFaction = self.faction
    local myProfession = self.profession

    -- If we have no faction, find one nearby or set a default
    if myFaction == "" or myFaction == "civilian" then
        -- Check if there's a cult nearby
        local nearby = world.nearby_entities()
        for _, id in ipairs(nearby) do
            local info = world.entity_info(id)
            if info and info.alive and info.faction == "cult" then
                -- Join the cult
                self.set_faction("cult")
                self.set_profession("cultist")
                myFaction = "cult"
                myProfession = "cultist"
                util.log(self.name .. " joined the cult")
                break
            end
        end
    end

    -- If still no faction, set as civilian
    if myFaction == "" or myFaction == "civilian" then
        self.set_faction("civilian")
        myFaction = "civilian"
    end

    -- Find nearby entities
    local nearby = world.nearby_entities()
    local factionMembers = {}
    local enemies = {}
    local potentialAllies = {}

    for _, id in ipairs(nearby) do
        if id ~= self.id then
            local info = world.entity_info(id)
            if info and info.alive and info.conscious then
                local theirFaction = info.faction

                -- Same faction = ally
                if theirFaction == myFaction then
                    table.insert(factionMembers, id)
                -- Hostile = enemy
                elseif world.is_hostile(myFaction, theirFaction) then
                    table.insert(enemies, id)
                -- Neutral = potential ally
                else
                    table.insert(potentialAllies, id)
                end
            end
        end
    end

    -- Priority 1: Defend against enemies
    if #enemies > 0 then
        local target = enemies[1]
        for _, id in ipairs(enemies) do
            local info = world.entity_info(id)
            local targetInfo = world.entity_info(target)
            if info and targetInfo and info.hp < targetInfo.hp then
                target = id
            end
        end

        -- Check relationship before attacking
        if self.has_relationship(target) then
            local relType = self.get_relationship_type(target)
            if relType == "love" then
                util.log(self.name .. " loves " .. world.entity_name(target) .. " — cannot attack")
                world.avoid_combat()
                return
            end
        end

        world.attack(self.id, target)

        -- Track the enemy
        if not self.has_relationship(target) then
            self.add_relationship(target, "rival", world.tick)
            util.log(self.name .. " marked " .. world.entity_name(target) .. " as rival")
        end
        return
    end

    -- Priority 2: Support faction members
    if #factionMembers > 0 then
        for _, id in ipairs(factionMembers) do
            local info = world.entity_info(id)
            if info and info.hp < info.max_hp * 0.5 then
                world.heal(id, 3)
                util.log(self.name .. " healed faction member " .. world.entity_name(id))
                return
            end
        end

        -- Strengthen bonds with faction members
        for _, id in ipairs(factionMembers) do
            if not self.has_relationship(id) then
                self.add_relationship(id, "friend", world.tick)
            end
        end
    end

    -- Priority 3: Convert potential allies
    if #potentialAllies > 0 then
        local target = potentialAllies[1]
        -- Prefer targets we already have a relationship with
        for _, id in ipairs(potentialAllies) do
            if self.has_relationship(id) then
                target = id
                break
            end
        end

        -- Talk to the potential ally
        world.talk_to(target)

        -- If they're not already in our faction, try to convert them
        local targetInfo = world.entity_info(target)
        if targetInfo and targetInfo.faction ~= myFaction and targetInfo.faction ~= "" then
            -- Shift their faction toward ours
            world.set_relation(myFaction, targetInfo.faction, "friendly")
            util.log(self.name .. " is trying to befriend " .. world.entity_name(target))
        end

        -- Track the relationship
        if not self.has_relationship(target) then
            self.add_relationship(target, "friend", world.tick)
            util.log(self.name .. " is befriending " .. world.entity_name(target))
        end
        return
    end

    -- Priority 4: If we have no relationships at all, start building them
    if #nearby > 0 then
        for _, id in ipairs(nearby) do
            local info = world.entity_info(id)
            if info and info.alive and info.conscious and info.faction ~= "" then
                if not self.has_relationship(id) then
                    self.add_relationship(id, "friend", world.tick)
                    util.log(self.name .. " is observing " .. world.entity_name(id))
                end
                break
            end
        end
    end

    -- Avoid combat when nothing else to do
    world.avoid_combat()
end

do_tick()
