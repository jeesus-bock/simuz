-- flesh_carver.lua
-- Tribal orc battlefield surgeon and tissue harvester.
-- Uses aggressive medicine to stitch allies and harvests materials from raw corpses.

local function execute_triage()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.hp < info.max_hp / 2 then
                local healAmt = 5 + util.rand_int(10)
                world.heal(eid, healAmt)
                util.log(self.name .. " brutally stitches " .. info.name .. "'s wounds. \"Hold still or I'll take a finger.\"")
                world.set_mood_target(eid, "stressed", 10)
                return true
            end
        end
    end
    return false
end

local function harvest_corpses()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and not info.alive then
                if world.steal(self.id, eid) then
                    util.log(self.name .. " carves useful parts from " .. info.name .. "'s remains.")
                    if world.has_material("leather") or util.rand_int(100) < 40 then
                        world.add_item(self.id, "raw_meat", 1)
                        util.log(self.name .. " strips a slab of raw meat from the corpse.")
                    end
                    return true
                end
            end
        end
    end
    return false
end

function do_tick()
    if world.tick % 7 ~= 0 then return {} end

    if self.hp and self.hp < 30 then
        util.log(self.name .. " tends to own wounds first. \"Can't carve if I'm dead.\"")
        world.heal(self.id, 3 + util.rand_int(5))
        return {util.event("profession_action", {profession = "flesh_carver"})}
    end

    if execute_triage() then
        return {util.event("profession_action", {profession = "flesh_carver"})}
    end

    if harvest_corpses() then
        return {util.event("profession_action", {profession = "flesh_carver"})}
    end

    if world.tick % 35 == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            world.move_to(exits[util.rand_int(#exits) + 1])
        end
    end

    return {}
end

return do_tick()
