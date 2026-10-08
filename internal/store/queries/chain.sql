-- An event's whole derivation chain: every derived event whose causes lead
-- back to the root, generation by generation, in log order. Backfill replays
-- this to know what already fired where; it is the provenance walk inverted.
SELECT event_id, kind, event_type, to_char(occurred_at,'YYYY-MM-DD'), recorded_at,
       payload, cause_event_id, rule_id, rule_version, dedup_key, actor,
       activity_name, activity_version
FROM event
WHERE event_id IN (
    WITH RECURSIVE chain AS (
        SELECT event_id FROM event WHERE cause_event_id = $1 AND kind = 'derived'
        UNION ALL
        SELECT e.event_id FROM event e JOIN chain c ON e.cause_event_id = c.event_id
        WHERE e.kind = 'derived'
    )
    SELECT event_id FROM chain
)
ORDER BY event_id
