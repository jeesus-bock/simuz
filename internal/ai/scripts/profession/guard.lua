-- guard.lua
-- Municipal tactical enforcement patrolling locations and neutralising outlaws.

local OUTLAWS = { "bandit_chief", "thief", "cultist" }

local function patrol_and_enforce()
    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                -- Check against outlaw registry definitions
                for _, outlaw_prof in ipairs(OUTLAWS) do
                    if info.profession == outlaw_prof or info.faction == "rust_walkers" then
                        util.log(self.name .. " blows the law horn! 'Halt, criminal scum!' Confronting " .. info.name)
                        util.set_mood("furious", 30)

                        -- Neutralize the disruptive element
                        world.attack(self.id, eid)
                        return
                    end
                end
            end
        end
    end
end

local function do_tick()
    local acted = false

    -- Guard rotations enforce rules 24/7 across both day and night cycles
    if world.tick % 8 == 0 then
        patrol_and_enforce()
        acted = true
    end

    -- Standard patrol route walk mechanics across city checkpoints
    if world.tick % 30 == 0 and util.rand_int(100) < 50 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local next_patrol_zone = exits[util.rand_int(#exits) + 1]
            world.move_to(next_patrol_zone)
            util.log(self.name .. " moved to keep security rotation at: " .. next_patrol_zone)
            acted = true
        end
    end

    return acted
end

return do_tick()
