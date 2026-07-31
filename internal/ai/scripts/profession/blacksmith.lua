-- blacksmith.lua
-- Production engine transforming metal inputs into martial weapons.

local function run_the_forge()
    util.set_mood("focused", 20)

    -- Consume raw ores to build military stock assets
    if world.has_material("iron_ore") then
        world.craft("recipe_iron_sword")
        util.log(self.name .. " hammered out a pristine iron arming sword on the anvil.")
    else
        -- If out of materials, attempt to buy stock from passing miners or merchants
        local nearby = world.nearby_entities()
        if nearby then
            for _, eid in ipairs(nearby) do
                local info = world.entity_info(eid)
                if info and info.alive and info.profession == "merchant" then
                    local trade = world.try_buy(eid, "iron_ore")
                    if trade and trade.done then
                        util.log(self.name .. " restocked on raw industrial iron ore packages.")
                        return
                    end
                end
            end
        end
    end
end

function do_tick()
    if world.phase == "night" then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
            return {util.event("profession_action", {profession = "blacksmith"})}
        end
        return {}
    end

    if world.tick % 16 == 0 then
        run_the_forge()
    end

    return {}
end

return do_tick()
