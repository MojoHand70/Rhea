-- Newest version per rule, in deterministic firing order: priority
-- ascending, rule_id as tiebreak (DECISIONS.md 2026-10-01). DISTINCT ON
-- forces its own ORDER BY, so the firing order is applied outside.
SELECT rule_id, version, status, priority, effective_from,
       created_by, description, spec, created_at
FROM (
    SELECT DISTINCT ON (rule_id)
           rule_id, version, status, priority,
           to_char(effective_from,'YYYY-MM-DD') AS effective_from,
           created_by, description, spec, created_at
    FROM rule
    ORDER BY rule_id, version DESC
) latest
ORDER BY priority, rule_id
