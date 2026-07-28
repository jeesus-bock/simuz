-- Merchant profession script: trades with others, builds friendly relationships
-- with customers, and avoids combat. Demonstrates profession-based behavior
-- and relationship building for trade.

local function do_tick()
    local myProfession = self.profession
    local myFaction = self.faction

    -- Ensure we have the merchant profession
    if myProfession ~= "merchant" then
        self.set_profession("merchant")
        myProfession = "merchant"
    end

    -- Ensure we have a merchant-friendly faction
    if myFaction ~= "merchant" and myFaction ~= "civilian" then
        self.set_faction("merchant")
        myFaction = "merchant"
    end

    -- Find nearby entities
    local nearby = world.nearby_entities()
    local customers = {}
    local threats = {}

    for _, id in ipairs(nearby) do
        if id ~= self.id then
            local info = world.entity_info(id)
            if info and info.alive and info.conscious then
                local theirFaction = info.faction
                local theirProf = info.profession

                -- Other merchants and civilians = customers
                if theirFaction == "merchant" or theirFaction == "civilian" or theirProf == "merchant" then
                    table.insert(customers, id)
                -- Hostile entities = threats
                elseif world.is_hostile(myFaction, theirFaction) then
                    table.insert(threats, id)
                end
            end
        end
    end

    -- Priority 1: Deal with threats — avoid combat
    if #threats > 0 then
        local target = threats[1]
        -- Try to flee rather than fight
        if not world.avoid_combat() then
            -- If we can't avoid, try to talk our way out
            world.talk_to(target)
        end
        return
    end

    -- Priority 2: Trade with customers
    if #customers > 0 then
        local target = customers[1]
        -- Prefer customers we already have a relationship with
        for _, id in ipairs(customers) do
            if self.has_relationship(id) then
                target = id
                break
            end
        end

        -- Try to buy from the customer (they might have goods to sell)
        local buySuccess = world.try_buy(target, "herb")
        if not buySuccess then
            -- Try to sell to the customer
            local sellSuccess = world.try_sell(target, "herb")
            if sellSuccess then
                util.log(self.name .. " sold herbs to " .. world.entity_name(target))
            end
        end

        -- Build friendly relationship with the customer
        if not self.has_relationship(target) then
            self.add_relationship(target, "friend", world.tick)
            util.log(self.name .. " met " .. world.entity_name(target) .. " — added friend")
        else
            -- Strengthen existing friendship
            local relType = self.get_relationship_type(target)
            if relType == "friend" then
                -- Already friends, maybe upgrade to mate-like bond for loyal customers
                -- (keeping as friend for now)
            end
        end

        -- Talk to the customer to build rapport
        world.talk_to(target)
        return
    end

    -- Priority 3: Find anyone to trade with or befriend
    if #nearby > 0 then
        for _, id in ipairs(nearby) do
            local info = world.entity_info(id)
            if info and info.alive and info.conscious and info.faction ~= myFaction then
                -- Try to befriend non-hostile entities
                if not self.has_relationship(id) then
                    self.add_relationship(id, "friend", world.tick)
                    util.log(self.name .. " is befriending " .. world.entity_name(id))
                end
                break
            end
        end
    end

    -- Avoid combat
    world.avoid_combat()
end

do_tick()
