-- tribute_collector.lua
-- Aggressive merchant alternative that collects regional resource taxes by threat.

local function extort_tribute()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            -- Extort non-orc characters or civilian commoners
            if info and info.alive and info.faction ~= self.faction then
                util.log(self.name .. " demands taxes from " .. info.name .. ". 'Pay your tribute to the clan or bleed!'")

                local tax = world.try_buy(eid, "iron_ore") or world.try_buy(eid, "coal")
                if tax and tax.done then
                    util.log(self.name .. " successfully collected the clan's dues.")
                else
                    util.log(self.name .. " screams: 'Defiance means death!'")
                    world.attack(self.id, eid)
                end
                return true
            end
        end
    end
    return false
end

function do_tick()
    local acted = false
    if world.tick % 8 == 0 then acted = extort_tribute() end
    if world.tick % 40 == 0 and util.rand_int(100) < 40 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            world.move_to(exits[util.rand_int(#exits) + 1])
            acted = true
        end
    end
    if acted then return { util.event("profession_action", { profession = "tribute_collector" }) } end
    return {}
end

return do_tick()
