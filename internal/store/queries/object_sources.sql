-- The record each object was made from: the object that the event its rule
-- fired on materialized or amended — the RMK record a spread's entry books,
-- the release a month's share books. One type at a time ($1), within the
-- caller's company ($2, empty: the whole log). An object a raw fact made
-- directly (an invoice's entry, beside the invoice) has none.
SELECT o.object_id, e.payload->>'object_id'
FROM object o
JOIN event e ON e.event_id = o.source_event_id
WHERE o.object_type = $1 AND ($2 = '' OR o.company = $2 OR COALESCE(o.company, '') = '')
  AND e.event_type IN ('object.materialized', 'object.amended')
  AND e.payload->>'object_id' IS NOT NULL
