-- Aggressive AI
-- Attacks hostiles in same location on sight.
-- Roams between adjacent locations when no target.
-- Rate-limited per-faction.

local function faction_hash()
    local h = 0
    for i = 1, #self.faction do
        h = h + string.byte(self.faction, i)
    end
    return h
end

local function get_attack_rate()
    return 5 + (faction_hash() % 5)
end

local function find_hostile_target()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return nil
    end
    -- Prefer conscious hostiles; finish knocked-out foes only when no active threat remains.
    local downed_id, downed_info = nil, nil
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.species ~= "deity" and world.is_hostile(self.faction, info.faction) then
            if info.conscious ~= false then
                return eid, info
            end
            if not downed_id then
                downed_id, downed_info = eid, info
            end
        end
    end
    return downed_id, downed_info
end

local function wander()
    -- Prefer real exits (camps, dens, sibling buildings) before falling back to children of parent
    local exits = world.travel_exits(self.loc_id)
    local targets = {}
    if exits then
        for _, e in ipairs(exits) do
            local tid = e.target_id
            if tid and tid ~= self.loc_id then
                table.insert(targets, tid)
            end
        end
    end
    if #targets == 0 then
        local parent_id = world.parent_location(self.loc_id)
        if parent_id and parent_id ~= "" then
            local children = world.exits_from(parent_id)
            if children then
                for _, sid in ipairs(children) do
                    if sid ~= self.loc_id then
                        table.insert(targets, sid)
                    end
                end
            end
        end
    end
    if #targets == 0 then
        return
    end
    local dest = targets[util.rand_int(#targets) + 1]
    local ok = world.move_to(dest)
    if ok then
        util.log(self.name .. " prowled to " .. dest)
    end
end

local function do_tick()
    local tick = world.tick
    local rate = get_attack_rate()

    if tick % rate ~= 0 then
        return
    end

    local target, info = find_hostile_target()
    if target then
        local ok = world.attack(self.id, target)
        if ok and info and info.hp <= 0 then
            util.log(self.name .. " killed " .. (info.name or target))
        end
        return
    end

    if tick % 30 == 0 and util.rand_int(100) < 40 then
        wander()
    end
end

do_tick()
