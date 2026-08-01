-- politician.lua
-- Species-specific political leader: human councilor, orc warlord,
-- dwarf thane, elf archon.  Interacts with diplomats, manages faction
-- relations, and controls diplomatic immunity.

local CIVILIZED = { human = true, orc = true, dwarf = true, elf = true }

local function find_diplomats()
    local nearby = world.nearby_entities()
    if not nearby then return {} end
    local diplomats = {}
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.profession == "diplomat" then
                table.insert(diplomats, info)
            end
        end
    end
    return diplomats
end

local function find_foreigners()
    local nearby = world.nearby_entities()
    if not nearby then return {} end
    local foreigners = {}
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.faction ~= "" and info.faction ~= self.faction and info.faction ~= "civilian" then
                table.insert(foreigners, info)
            end
        end
    end
    return foreigners
end

local function find_rivals()
    local nearby = world.nearby_entities()
    if not nearby then return {} end
    local rivals = {}
    for _, eid in ipairs(nearby) do
        if eid ~= self.id then
            local info = world.entity_info(eid)
            if info and info.alive and info.profession == "politician" and info.species ~= self.species then
                table.insert(rivals, info)
            end
        end
    end
    return rivals
end

-- -----------------------------------------------------------------------
-- Human Councilor: pragmatic, alliance-seeking, wealth-driven
-- -----------------------------------------------------------------------
local function human_politician()
    local diplomats = find_diplomats()
    local rivals = find_rivals()

    for _, dip in ipairs(diplomats) do
        if world.can_communicate(dip.id) then
            local lang = world.best_shared_language(dip.id)
            util.log(self.name .. " holds court with diplomat " .. dip.name .. ".")
            world.say_to(dip.id, "The council values stability. What terms do you bring?")
            util.set_mood("authoritative", 20)
            return true
        end
    end

    if #rivals > 0 then
        local rival = rivals[1]
        util.log(self.name .. " eyes rival leader " .. rival.name .. " with calculated suspicion.")
        util.set_mood("scheming", 15)
        return true
    end

    local foreigners = find_foreigners()
    if #foreigners > 0 then
        local target = foreigners[1]
        if world.can_communicate(target.id) then
            world.say_to(target.id, "Perhaps our factions can find common ground.")
            util.log(self.name .. " opens trade negotiations with " .. target.name .. ".")
            util.set_mood("diplomatic", 20)
            return true
        end
    end

    return false
end

-- -----------------------------------------------------------------------
-- Orc Warlord: strength-based, tribute-demanding, may execute diplomats
-- -----------------------------------------------------------------------
local function orc_warlord()
    local diplomats = find_diplomats()

    for _, dip in ipairs(diplomats) do
        local roll = util.rand_int(100)
        if roll < 25 then
            util.log(self.name .. " snarls at " .. dip.name .. ": 'Your words are weak! GUARDS!'")
            world.set_entity_relation(dip.id, "hostile")
            util.set_mood("furious", 30)
            world.attack(self.id, dip.id)
            return true
        elseif roll < 60 then
            util.log(self.name .. " demands tribute from diplomat " .. dip.name .. ".")
            local tribute = world.try_buy(dip.id, "gold_pouch")
            if tribute and tribute.done then
                util.log(self.name .. " accepts the tribute. The diplomat lives... for now.")
            end
            util.set_mood("dominant", 20)
            return true
        else
            if world.can_communicate(dip.id) then
                world.say_to(dip.id, "Speak, soft-skin. Make it worth my time.")
                util.log(self.name .. " reluctantly grants audience to " .. dip.name .. ".")
                util.set_mood("dismissive", 15)
            end
            return true
        end
    end

    local rivals = find_rivals()
    if #rivals > 0 then
        local rival = rivals[1]
        util.log(self.name .. " pounds chest at rival " .. rival.name .. ". 'This is MY territory!'")
        util.set_mood("aggressive", 25)
        return true
    end

    local nearby = world.nearby_entities()
    if nearby then
        for _, eid in ipairs(nearby) do
            if eid ~= self.id then
                local info = world.entity_info(eid)
                if info and info.alive and info.profession == "guard" and info.faction == self.faction then
                    util.log(self.name .. " orders " .. info.name .. " to patrol the perimeter.")
                    return true
                end
            end
        end
    end

    return false
end

-- -----------------------------------------------------------------------
-- Dwarf Thane: trade-focused, conservative, clan-bound
-- -----------------------------------------------------------------------
local function dwarf_thane()
    local diplomats = find_diplomats()

    for _, dip in ipairs(diplomats) do
        if world.can_communicate(dip.id) then
            world.say_to(dip.id, "State your business. Clan business is sacred.")
            util.log(self.name .. " grants a formal audience to diplomat " .. dip.name .. ".")
            util.set_mood("measured", 15)
            return true
        else
            util.log(self.name .. " waves off " .. dip.name .. ". 'I dinnae understand yer jabber.'")
            util.set_mood("gruff", 10)
            return true
        end
    end

    local foreigners = find_foreigners()
    for _, f in ipairs(foreigners) do
        if world.can_communicate(f.id) then
            world.say_to(f.id, "Got coin? Then we talk. No coin, no deal.")
            util.log(self.name .. " opens trade talks with " .. f.name .. ".")
            util.set_mood("transactional", 15)
            return true
        end
    end

    return false
end

-- -----------------------------------------------------------------------
-- Elf Archon: long-view strategist, nature-aligned, proud
-- -----------------------------------------------------------------------
local function elf_archon()
    local diplomats = find_diplomats()

    for _, dip in ipairs(diplomats) do
        if dip.species == "orc" or dip.species == "goblin" then
            util.log(self.name .. " regards " .. dip.name .. " with barely concealed disdain. 'Your kind scars the land.'")
            util.set_mood("aloof", 20)
            return true
        elseif world.can_communicate(dip.id) then
            world.say_to(dip.id, "Speak, mortal. Time is but a river—what matters is the current.")
            util.log(self.name .. " engages diplomat " .. dip.name .. " in measured discourse.")
            util.set_mood("serene", 15)
            return true
        end
    end

    local rivals = find_rivals()
    if #rivals > 0 then
        util.log(self.name .. " observes rival " .. rivals[1].name .. " with ancient patience.")
        util.set_mood("contemplative", 15)
        return true
    end

    return false
end

-- -----------------------------------------------------------------------
-- Main tick: dispatch to species-specific behaviour
-- -----------------------------------------------------------------------
function do_tick()
    if not CIVILIZED[self.species] then
        return {}
    end

    local acted = false

    if world.tick % 15 == 0 then
        if self.species == "human" then
            acted = human_politician()
        elseif self.species == "orc" then
            acted = orc_warlord()
        elseif self.species == "dwarf" then
            acted = dwarf_thane()
        elseif self.species == "elf" then
            acted = elf_archon()
        end

        if not acted then
            local exits = world.exits_from(self.loc_id)
            if exits and #exits > 0 then
                local dest = exits[util.rand_int(#exits) + 1]
                if dest ~= self.loc_id then
                    util.log(self.name .. " moves to inspect the domain.")
                    world.move_to(dest)
                    acted = true
                end
            end
        end
    end

    if acted then
        return {util.event("profession_action", {profession = "politician", species = self.species})}
    end
    return {}
end

return do_tick()
