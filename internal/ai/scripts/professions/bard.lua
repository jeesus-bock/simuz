-- Bard AI
-- Looks for inns to play and gather money, bread, or wine.
-- Travels between towns at night, sometimes sleeps in an inn,
-- but prefers resting under the sun of a green field.

local function find_inn()
    local nearby = world.nearby_entities()
    if not nearby then return nil end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.profession == "innkeeper" then
            return eid, info
        end
    end
    return nil
end

local function play_for_reward()
    -- Simple "play" action: try to get money, bread, or wine.
    -- In a real game this would involve dialogue or a mini-game;
    -- here we just log the attempt and assume a reward.
    util.log(self.name .. " plays for money, bread, or wine")
    -- Example reward: add some gold (or other items) to the entity.
    if world.add_item then
        world.add_item(self.id, "gold", 5)
    end
end

local function travel_to_next_town()
    local exits = world.exits_from(self.loc_id)
    if not exits or #exits == 0 then return end
    local dest = exits[util.rand_int(#exits) + 1]
    world.move_to(dest)
    util.log(self.name .. " travels to a new town")
end

local function rest()
    local phase = world.phase
    if phase == "night" then
        -- Sometimes sleep in an inn, otherwise just rest on the field.
        if util.rand_int(100) < 30 and find_inn() then
            util.log(self.name .. " sleeps in an inn")
        else
            util.log(self.name .. " rests under the sun of a green field")
        end
    end
end

local function do_tick()
    local phase = world.phase

    -- Night: travel to next town.
    if phase == "night" then
        travel_to_next_town()
        rest()
        return {}
    end

    -- Day: look for an inn and play.
    local inn_id, inn_info = find_inn()
    if inn_id then
        play_for_reward()
    else
        -- No inn nearby – just rest on the field.
        rest()
    end

    return {}
end

return do_tick()
