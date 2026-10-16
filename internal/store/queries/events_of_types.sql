-- Events of the named types, oldest first: the reads that ask the log a
-- narrow question (who approved what, what the host said) without loading
-- every raw fact. $2 is the caller's company; empty reads the whole log —
-- definitions and their approvals are installation-wide.
SELECT event_id, kind, event_type, to_char(occurred_at,'YYYY-MM-DD'), recorded_at,
       payload, cause_event_id, rule_id, rule_version, dedup_key, actor,
       activity_name, activity_version, company
FROM event
WHERE event_type = ANY($1) AND ($2 = '' OR company = $2)
ORDER BY event_id
