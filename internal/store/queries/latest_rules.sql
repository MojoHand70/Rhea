-- Newest version per rule, in deterministic firing order:
-- priority ascending, rule_id as tiebreak (DECISIONS.md 2026-10-01).
SELECT DISTINCT ON (rule_id)
       rule_id, version, status, priority, to_char(effective_from,'YYYY-MM-DD'),
       created_by, description, spec, created_at
FROM rule
ORDER BY rule_id, version DESC
