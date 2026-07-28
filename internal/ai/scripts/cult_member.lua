-- Cult faction member script: loyal to cult, converts others, uses relationships
-- to track converts and enemies. Demonstrates faction-based behavior and
-- relationship management for a voluntary group membership.

local function do_tick()
    local myFaction = self.faction
    local myProfession = self.profession

    -- Ensure we have the cult faction set
    if myFaction ~= "cult" then
        self.set_faction("cult")
        myFaction = "cult"
    end

    -- Set profession if not already set
    if myProfession == "" then
        self.set_profession("cultist")
        myProfession = "cultist"
    end

    -- Find nearby entities
    local nearby = world.nearby_entities()
    local cultists = {}
    local potentialConverts = {}
    local enemies = {}

    for _, id in ipairs(nearby) do
        if id ~= self.id then
            local info = world.entity_info(id)
            if info and info.alive and info.conscious then
                local theirFaction = info.faction
                local theirProf = info.profession

                -- Same faction = ally
                if theirFaction == "cult" then
                    table.insert(cultists, id)
                -- Non-hostile, non-cult = potential convert
                elseif theirFaction ~= "" and not world.is_hostile(myFaction, theirFaction) then
                    table.insert(potentialConverts, id)
                -- Hostile = enemy
                else
                    table.insert(enemies, id)
                end
            end
        end
    end

    -- Priority 1: Defend against enemies
    if #enemies > 0 then
        local target = enemies[1]
        -- Pick the weakest enemy
        for _, id in ipairs(enemies) do
            local info = world.entity_info(id)
            local targetInfo = world.entity_info(target)
            if info and targetInfo and info.hp < targetInfo.hp then
                target = id
            end
        end

        -- Check relationship: if we have a rival or love relationship, adjust behavior
        local relType = self.get_relationship_type(target)
        if relType == "love" then
            -- We love this entity, don't attack
            util.log(self.name .. " refuses to attack loved one " .. world.entity_name(target))
            return
        elseif relType == "rival" then
            util.log(self.name .. " is attacking rival " .. world.entity_name(target))
        end

        -- Attack the enemy
        world.attack(self.id, target)
        return
    end

    -- Priority 2: Convert potential converts
    if #potentialConverts > 0 then
        local target = potentialConverts[1]
        -- Prefer targets we already have a relationship with
        for _, id in ipairs(potentialConverts) do
            if self.has_relationship(id) then
                target = id
                break
            end
        end

        -- Try to talk to the target first
        world.talk_to(target)

        -- Give them a quest to test their loyalty
        world.give_quest(target, "cult_initiation")

        -- Add a relationship tracking the conversion attempt
        if not self.has_relationship(target) then
            self.add_relationship(target, "friend", world.tick)
            util.log(self.name .. " is trying to convert " .. world.entity_name(target) .. " — added friend")
        end

        -- If the target is already a friend, try to upgrade to cult member
        local relType = self.get_relationship_type(target)
        if relType == "friend" then
            -- Set their faction to cult
            world.set_faction(target, "cult")
            world.set_profession(target, "cultist")
            -- Upgrade relationship to mate-like bond (cult brotherhood)
            self.add_relationship(target, "mate", world.tick)
            util.log(self.name .. " converted " .. world.entity_name(target) .. " to the cult")
        end

        return
    end

    -- Priority 3: Support fellow cultists
    if #cultists > 0 then
        -- Check if any cultist is in trouble
        for _, id in ipairs(cultists) do
            local info = world.entity_info(id)
            if info and info.hp < info.max_hp * 0.5 then
                -- Heal them
                world.heal(id, 5)
                util.log(self.name .. " healed cultist " .. world.entity_name(id))
                return
            end
        end

        -- If all cultists are healthy, strengthen bonds
        for _, id in ipairs(cultists) do
            if not self.has_relationship(id) then
                self.add_relationship(id, "friend", world.tick)
            end
        end
    end

    -- Priority 4: Spread the word — find non-hostile entities to convert
    if #nearby > 0 then
        for _, id in ipairs(nearby) do
            local info = world.entity_info(id)
            if info and info.alive and info.conscious and info.faction ~= "cult" and info.faction ~= "" then
                -- Try to convert anyone we haven't interacted with
                if not self.has_relationship(id) then
                    self.add_relationship(id, "friend", world.tick)
                    util.log(self.name .. " is watching " .. world.entity_name(id) .. " for conversion opportunity")
                end
                break
            end
        end
    end

    -- If nothing else to do, avoid combat and stay near cult members
    world.avoid_combat()
end

do_tick()
