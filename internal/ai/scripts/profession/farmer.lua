-- farmer.lua
-- Routine for agricultural production and regional trade.

local function tend_crops()
    local weather = world.weather(self.loc_id)

    -- Farmers thrive in clear/sunny weather, but rest during harsh thunderstorms
    if weather == "thunderstorm" then
        util.set_mood("annoyed", 10)
        util.log(self.name .. " hunkers down to wait out the violent storm.")
        return
    end

    -- Process harvesting loop
    util.set_mood("focused", 15)
    world.add_item(self.id, "raw_grain", 1)
    util.log(self.name .. " harvested a bundle of fresh grain.")

    -- Look for economic trading vectors nearby
    local nearby = world.nearby_entities()
    if nearby then
        for _, eid in ipairs(nearby) do
            local info = world.entity_info(eid)
            if info and info.alive and info.profession == "baker" then
                world.talk_to(eid)
                local sale = world.try_sell(eid, "raw_grain")
                if sale and sale.done then
                    util.log(self.name .. " traded grain directly to baker: " .. info.name)
                    util.set_mood("happy", 20)
                    break
                end
            end
        end
    end
end

function do_tick()
    if world.phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            return {util.event("profession_action", {profession = "farmer"})}
        end
        return {}
    end

    if world.tick % 15 == 0 then
        tend_crops()
    end

    return {}
end

return do_tick()
