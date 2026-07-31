-- courier.lua
-- Route-navigation script processing parcels and message delivery loops.

local function attempt_delivery()
    local nearby = world.nearby_entities()
    if not nearby then return false end

    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            -- Check if this specific entity is our assigned target contact
            if info and info.alive and info.profession == "guard" then
                world.talk_to(eid)
                local deliver = world.deliver_item(eid, "royal_scroll")
                if deliver then
                    util.log(self.name .. " successfully dispatched order documents to " .. info.name)
                    util.set_mood("happy", 25)
                    world.add_item("courier_token")
                    return true
                end
            end
        end
    end
    return false
end

function do_tick()
    if world.tick % 10 == 0 then
        local done = attempt_delivery()

        -- If no delivery target is found in this location, continuously move down the road
        if not done and util.rand_int(100) < 75 then
            local exits = world.exits_from(self.loc_id)
            if exits and #exits > 0 then
                local next_hop = exits[util.rand_int(#exits) + 1]
                world.move_to(next_hop)
                util.log(self.name .. " is traveling down route to sector: " .. next_hop)
            end
        end

        if done then
            return {util.event("profession_action", {profession = "courier"})}
        end
    end

    return {}
end

return do_tick()
