-- Werewolf AI
-- A cursed human who appears normal during the day but transforms into a savage
-- beast at dusk/night, attacking any non-cultist/non-civilian (prey) nearby.
-- Returns to a safe lair at dawn.

local function is_transformed()
    return world.phase == "night" or world.phase == "dusk"
end

local function find_prey()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and info.species ~= "deity" then
            if info.faction ~= "cultist" and info.faction ~= "undead" then
                return eid, info
            end
        end
        ::continue::
    end
    return nil
end

local function find_hostile()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        if eid == self.id then goto continue end
        local info = world.entity_info(eid)
        if info and info.alive and world.is_hostile(self.faction, info.faction) then
            return eid, info
        end
        ::continue::
    end
    return nil
end

local function return_to_lair()
    if self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
        util.log(self.name .. " retreated to the lair")
    end
end

local function do_tick()
    local phase = world.phase
    local tick = world.tick

    if phase == "dawn" then
        return_to_lair()
        util.set_mood("tired")
        return
    end

    if is_transformed() then
        if self.home and self.loc_id == self.home then
            -- leave the lair to hunt
            local exits = world.exits_from(self.loc_id)
            if exits and #exits > 0 then
                local dest = exits[util.rand_int(#exits) + 1]
                world.move_to(dest)
            end
        end

        local target, info = find_hostile()
        if not target then
            target, info = find_prey()
        end
        if target and tick % 3 == 0 then
            local hit = world.attack(self.id, target)
            if hit then
                util.log(self.name .. " savaged " .. info.name)
            end
        end
        return
    end

    -- Daytime: act like a wounded or tired civilian
    if tick % 30 == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 and util.rand_int(100) < 30 then
            local dest = exits[util.rand_int(#exits) + 1]
            world.move_to(dest)
        end
    end

    if tick % 60 == 0 then
        local roll = util.rand_int(100)
        if roll < 50 then
            util.set_mood("neutral")
        else
            util.set_mood("tired")
        end
    end
end

do_tick()
