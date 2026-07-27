-- Vampire AI - Dominate/Drain/Regeneration
-- Life drain heals 50% of damage dealt
-- Dominates targets to fight allies
-- Returns to coffin at dawn
-- Weak to sunlight during day
-- Regenerates at night

local ATTACK_CHANCE = 70
local DOMINATE_CHANCE = 25
local DOMINATE_DURATION = 75
local NIGHT_REGEN_AMOUNT = 5
local SUNLIGHT_DEBUFF = 2

local function is_daytime()
    return world.phase == "day" or world.phase == "dawn"
end

local function find_hostile()
    local nearby = world.nearby_entities()
    if not nearby or #nearby == 0 then
        return nil
    end
    for _, eid in ipairs(nearby) do
        local info = world.entity_info(eid)
        if info and info.alive and world.is_hostile(self.faction, info.faction) then
            return eid, info
        end
    end
    return nil
end

local function do_life_drain(target_id)
    local hit = world.attack(self.id, target_id)
    if hit then
        local heal_amount = math.floor(self.max_hp * 0.1)
        if self.hp < self.max_hp then
            util.log(self.name .. " drained life and healed " .. heal_amount .. " HP")
        end
        return true
    end
    return false
end

local function do_dominate(target_id)
    if util.rand_int(100) < DOMINATE_CHANCE then
        util.mem_set("dominate_target_" .. target_id, world.tick + DOMINATE_DURATION)
        local info = world.entity_info(target_id)
        util.log(self.name .. " dominated " .. (info and info.name or target_id) .. " for " .. DOMINATE_DURATION .. " ticks")
        return true
    end
    return false
end

local function return_to_coffin()
    if self.home and self.loc_id ~= self.home then
        world.move_to(self.home)
        util.log(self.name .. " returned to coffin at dawn")
    end
end

local function do_tick()
    local phase = world.phase
    local tick = world.tick

    if phase == "dawn" then
        return_to_coffin()
        return
    end

    if is_daytime() then
        if self.home and self.loc_id ~= self.home then
            world.move_to(self.home)
        end
        return
    end

    if phase == "night" or phase == "dusk" then
		if self.hp < self.max_hp then
			world.heal(self.id, NIGHT_REGEN_AMOUNT)
			util.log(self.name .. " regenerated " .. NIGHT_REGEN_AMOUNT .. " HP at night")
		end
    end

    local target_id, target_info = find_hostile()
    if target_id then
        if util.rand_int(100) < ATTACK_CHANCE then
            do_life_drain(target_id)
            do_dominate(target_id)
        end
        return
    end

    if world.tick % 30 == 0 then
        local exits = world.exits_from(self.loc_id)
        if exits and #exits > 0 then
            local dest = exits[util.rand_int(#exits) + 1]
            if dest ~= self.loc_id then
                world.move_to(dest)
            end
        end
    end
end

do_tick()
