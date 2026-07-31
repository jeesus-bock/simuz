-- smog_enforcer.lua
-- Behavioral script for an Orcish Cartel Enforcer collecting protection fees.

local REJECTED_FACTIONS = { "withered_root", "bleeding_quill" }

local function check_targets()
    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive then
                -- 1. Strike immediate blood enemies
                for _, rival in ipairs(REJECTED_FACTIONS) do
                    if info.faction == rival then
                        util.log(self.name .. " spotted a rival " .. rival .. "! Attacking!")
                        world.attack(self.id, eid)
                        util.set_mood("furious", 30)
                        return
                    end
                end

                -- 2. Shakedown non-cartel Blacksmiths or Merchants
                if info.profession == "blacksmith" or info.profession == "merchant" then
                    if info.faction ~= "smog_iron_cartel" then
                        -- Check memory so we don't shake down the same guy every single tick
                        local last_tax = util.mem_get("taxed_" .. eid)
                        if not last_tax then
                            util.log(self.name .. " is cornering " .. info.name .. " for independent forge taxes.")
                            world.talk_to(eid)

                            -- Force them to buy protection (try_sell protection tokens)
                            local result = world.try_sell(eid, "cartel_stamp")
                            if result and result.done then
                                util.log(self.name .. " successfully collected protection fees from " .. info.name)
                                util.mem_set("taxed_" .. eid, world.tick)
                                util.set_mood("relaxed", 20)
                            else
                                -- If they refuse to pay, crack their knees
                                util.log(info.name .. " refused to pay the Smog tax! Smashing foundries!")
                                world.attack(self.id, eid)
                                util.set_mood("furious", 15)
                            end
                            return
                        end
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
            return {util.event("flee", {})}
        end
        return {}
    end

    local events = {}
    if world.tick % 12 == 0 then
        check_targets()
        table.insert(events, util.event("profession_action", {profession = "enforcer"}))
    end

    if world.tick % 40 == 0 and util.rand_int(100) < 40 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local next_loc = exits[util.rand_int(#exits) + 1]
            world.move_to(next_loc)
            util.log(self.name .. " marched to patrol location: " .. next_loc)
            table.insert(events, util.event("move", {destination = next_loc}))
        end
    end
    return events
end

return do_tick()
