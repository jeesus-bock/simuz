-- fisherman.lua
-- Maritime extraction routine gathering fish commodities and tracking storm weather.

local function execute_fishing()
    local weather = world.weather(self.loc_id)
    local catch_chance = 30

    -- Fish bite significantly better during rainy overcast weather
    if weather == "rain" or weather == "fog" then
        catch_chance = 70
        util.set_mood("focused", 15)
    elseif weather == "thunderstorm" then
        util.log(self.name .. " reeling line in—waves are too violent for fishing.")
        util.set_mood("annoyed", 10)
        return
    end

    if util.rand_int(100) < catch_chance then
        world.add_item(self.id, "raw_fish", 1)
        util.log(self.name .. " successfully pulled a heavy silver fish out of the water.")

        -- Progress survival/cooking tasks if tracked
        world.quest_progress("feed_the_docks_q", "catch_fish", 1)
    end

    -- Trade catch to passing cooking professionals
    local nearby = world.nearby_entities()
    if nearby then
        for _, eid in ipairs(nearby) do
            local info = world.entity_info(eid)
            if info and info.alive and info.profession == "cooking" then
                world.talk_to(eid)
                local sale = world.try_sell(eid, "raw_fish")
                if sale and sale.done then
                    util.log(self.name .. " sold fresh catch tokens to cook: " .. info.name)
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
            return {util.event("profession_action", {profession = "fisherman"})}
        end
        return {}
    end

    if world.tick % 18 == 0 then
        execute_fishing()
    end

    return {}
end

return do_tick()
