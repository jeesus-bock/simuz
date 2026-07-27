-- Deity AI
-- Active deities watch over the mortal realm.
-- Smite hostiles, bless allies, answer prayers.

local function do_divine_intervention()
    if not world.divine_intervention then return end

    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if not info or not info.alive then goto continue end

        if info.species ~= "deity" and world.is_hostile(self.faction, info.faction) then
            local result = world.divine_intervention(self.id, eid, "smite")
            if result and result.done and result.targets > 0 then
                util.log(self.name .. " smote " .. (info.name or eid))
            end
            goto continue
        end

        if info.faction == "civilian" and info.hp < info.max_hp / 2 then
            local result = world.divine_intervention(self.id, eid, "heal")
            if result and result.done then
                util.log(self.name .. " blessed " .. (info.name or eid) .. " with healing")
            end
            goto continue
        end

        if info.faction == "civilian" and world.tick % 100 == 0 then
            local result = world.divine_intervention(self.id, eid, "bless")
            if result and result.done then
                util.log(self.name .. " blessed " .. (info.name or eid))
            end
        end

        ::continue::
    end
end

local function do_zeus_intervention()
    if self.name ~= "Zeus" then return end
    if world.tick % 200 ~= 0 then return end

    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if not info or not info.alive then goto continue end

        if info.species == "human" and info.faction == "civilian" then
            local result = world.divine_intervention(self.id, eid, "quest", "zeus_crazy_task")
            if result and result.done then
                util.log(self.name .. " compelled " .. (info.name or eid) .. " to perform a crazy task!")
            end
            goto continue
        end

        ::continue::
    end
end

local function do_tick()
    if world.tick % 5 ~= 0 then return end
    do_divine_intervention()
    do_zeus_intervention()
end

do_tick()
