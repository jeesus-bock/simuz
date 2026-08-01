-- flesh_carver.lua
-- Tribal orc battlefield surgeon and tissue harvester.
-- Uses aggressive medicine to stitch allies and harvests materials from raw corpses.

local function execute_triage()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    -- 1. Intimidating self-preservation
    if self.hp and self.hp  0 then
            world.move_to(exits[util.rand_int(#exits) + 1])
            acted = true
        end
    end

    if acted then
        return {util.event("profession_action", {profession = "flesh_carver"})}
    end
    return {}
end

return do_tick()
