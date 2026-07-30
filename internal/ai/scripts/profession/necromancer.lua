-- necromancer.lua
-- Master engine script raising minions, managing soul vats, and casting curses.

local function run_dark_assembly()
    -- 1. If we have harvested bone components, assemble a new servant minion
    if world.has_material("ancient_skull") and world.has_material("dark_essence") then
        world.craft("recipe_summon_skeleton")
        util.log(self.name .. " weaves necrotic threads, raising a new skeletal thrall from the dirt.")
        world.quest_progress("legion_of_bone_q", "raise_minions", 1)
        util.set_mood("focused", 30)
        return
    end

    -- 2. Scan area to command minions or curse holy interlopers
    local nearby = world.nearby_entities()
    if nearby then
        for _, eid in ipairs(nearby) do
            if eid ~= self.id then
                local info = world.entity_info(eid)
                if info and info.alive then
                    -- Attack holy priests or elite guards encroaching on the ritual altar
                    if info.profession == "priest" or info.profession == "guard" then
                        util.log(self.name .. " extends a withered hand to blast " .. info.name .. " with rot curses!")
                        util.set_mood("furious", 20)

                        world.damage_location(self.id, 15) -- Apply heavy area shadow decay
                        world.attack(self.id, eid)
                        return
                    end
                end
            end
        end
    end
end

local function do_tick()
    -- Necromancers perform their main spell-weaving routines inside their dark crypts at night
    if world.phase == "day" then
        if self.home and self.loc_id ~= self.home then world.move_to(self.home) end
        return
    end

    if world.tick % 14 == 0 then
        run_dark_assembly()
    end
end

do_tick()
