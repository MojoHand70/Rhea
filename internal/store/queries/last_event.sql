-- The newest event the caller's scope sees: how far the executor has read
-- the log when it remembers which facts still wait.
SELECT COALESCE(max(event_id), 0) FROM event
