-- A record's history (KK 2026-11-05: the stream on every record as its
-- Historia): the roots of every chain that names the object — the fact that
-- made it, the payment that settled it, the close that counted it. Each
-- event naming the id in its payload walks up its causes to its raw root.
SELECT event_id, kind, event_type, to_char(occurred_at,'YYYY-MM-DD'), recorded_at,
       payload, cause_event_id, rule_id, rule_version, dedup_key, actor,
       activity_name, activity_version, company
FROM event
WHERE event_id IN (
    WITH RECURSIVE up AS (
        SELECT e.event_id, e.cause_event_id FROM event e
        WHERE ($1 = '' OR e.company = $1 OR COALESCE(e.company, '') = '')
          AND e.payload::text LIKE '%"' || $2 || '"%'
        UNION
        SELECT e.event_id, e.cause_event_id FROM event e JOIN up ON e.event_id = up.cause_event_id
    )
    SELECT event_id FROM up WHERE cause_event_id IS NULL
)
ORDER BY event_id DESC
LIMIT $3
