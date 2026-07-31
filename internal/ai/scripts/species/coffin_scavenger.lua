-- coffin_scavenger.lua
-- Behavioral script for an undead collector dragging corpses back to the bone pits.

local function harvest_graveyard()
    local nearby = world.nearby_entities()
    if not nearby then return end

    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        -- Look for dead bodies left behind from simulation fights
        if info and not info.alive then
            -- Ensure no one else is currently hauling this resource
            if not world.is_leashed(eid) then
                util.log(self.name .. " found a fresh corpse: " .. info.name .. ". Securing harvest hooks.")
                world.drag_entity(eid)
                util.mem_set("hauling_target", eid)
                return true
            end
        end
    end
    -- If no dead bodies, try bargaining with living Gravediggers for spare bones
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and info.profession == "gravedigger" then
            local trade = world.try_buy(eid, "ancient_skull")
            if trade and trade.done then
                util.log(self.name .. " bought a pristine ritual skull from a local gravedigger.")
                return true
            end
        end
    end
    return false
end

function do_tick()
    -- Inverse routine: Sleep during day, active at night
    if world.phase == "day" then
        if self.home and self.loc_id ~= self.home then
            -- If we were dragging a corpse home when dawn hit, finish hauling it
            local active_haul = util.mem_get("hauling_target")
            if active_haul then
                world.move_to(self.home)
                if self.loc_id == self.home then
                    world.undrag_entity(active_haul)
                    world.quest_progress("bone_pile_q", "bodies_delivered", 1)
                    util.mem_set("hauling_target", nil)
                    util.log(self.name .. " successfully locked the harvested corpse inside the crypt before dawn.")
                end
            else
                -- Otherwise safely head back to sleep inside the crypt tombs
                world.move_to(self.home)
            end
            return {util.event("move", {})}
        end
        return {}
    end

    -- Night Operations Loop
    local current_haul = util.mem_get("hauling_target")

    if not current_haul then
        -- We have free hands, go look for dead bodies or harvest targets every 10 ticks
        if world.tick % 10 == 0 then
            local found = harvest_graveyard()
            -- If the graveyard is empty, migrate to adjacent battlefield map sectors
            if not found and util.rand_int(100) < 50 then
                local exits = world.exits_from(self.loc_id)
                if exits and #exits > 0 then
                    local target_spot = exits[util.rand_int(#exits) + 1]
                    world.move_to(target_spot)
                    return {util.event("move", {})}
                end
            end
            return found and {util.event("hunt", {})} or {}
        end
    else
        -- We are dragging a body! Immediately path back to the home crypt
        if world.tick % 8 == 0 then
            if self.loc_id ~= self.home then
                world.move_to(self.home)
                util.log(self.name .. " is dragging a heavy corpse back through the mud lanes...")
            else
                -- Arrived home! Process the carcass conversion quest lines
                world.undrag_entity(current_haul)
                world.quest_progress("bone_pile_q", "bodies_delivered", 1)
                util.log(self.name .. " dumped the harvest into the necromancy meat-vats.")
                util.mem_set("hauling_target", nil)
                util.set_mood("relaxed", 30)
            end
            return {util.event("species_action", {})}
        end
    end
    return {}
end

return do_tick()
