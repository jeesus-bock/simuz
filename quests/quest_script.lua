-- Quest script placeholder for Simuz

-- Initialise quest state for a given entity
-- @param entity_id string ID of the entity
function init_quest(entity_id)
    -- TODO: set up quest progress, store in entity.Flags or a quest manager
    print("[Quest] Initialising quest for entity " .. entity_id)
end

-- Progress the quest to the next step
-- @param entity_id string ID of the entity
-- @param step string or number representing the next step
-- @return boolean indicating success
function progress_quest(entity_id, step)
    -- TODO: validate step, update quest state
    print("[Quest] Entity " .. entity_id .. " progresses to step " .. tostring(step))
    return true
end

-- Mark the quest as complete for the entity
-- @param entity_id string ID of the entity
function complete_quest(entity_id)
    -- TODO: finalize quest, grant rewards, clean up state
    print("[Quest] Quest completed for entity " .. entity_id)
end
