-- Whether a later fact names a record (the stream's "waits for a person":
-- a proposed account waits until the approval or the refusal names it; a
-- task until a person's decision does). One row when some event after the
-- given one carries the id in its payload, within the caller's scope.
SELECT 1 FROM event e
WHERE e.event_id > $2
  AND ($1 = '' OR e.company = $1 OR COALESCE(e.company, '') = '')
  AND e.payload::text LIKE '%"' || $3 || '"%'
LIMIT 1
