-- The raw root of each object's chain: walk the materializing event's causes
-- up to the fact nothing caused. One type at a time ($1), within the caller's
-- company ($2, empty: the whole log). A face finds an object's document — the
-- invoice a posting books, the bank line a payment entry came from — by the
-- root they share, without reading each chain.
WITH RECURSIVE up AS (
    SELECT o.object_id, e.event_id, e.cause_event_id
    FROM object o JOIN event e ON e.event_id = o.source_event_id
    WHERE o.object_type = $1 AND ($2 = '' OR o.company = $2)
  UNION ALL
    SELECT up.object_id, e.event_id, e.cause_event_id
    FROM up JOIN event e ON e.event_id = up.cause_event_id
)
SELECT object_id, event_id FROM up WHERE cause_event_id IS NULL
