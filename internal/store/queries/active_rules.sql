-- The executing rule set: the newest ACTIVE version of every rule, in firing
-- order (priority ascending, rule_id tiebreak). A draft or approved version
-- in progress does NOT suspend the running version; only a rule whose very
-- latest version is superseded is retired.
SELECT a.rule_id, a.version, a.status, a.priority,
       to_char(a.effective_from,'YYYY-MM-DD'),
       a.created_by, a.description, a.spec, a.created_at
FROM (
    SELECT DISTINCT ON (rule_id) rule_id, status AS latest_status
    FROM rule
    ORDER BY rule_id, version DESC
) latest
JOIN LATERAL (
    SELECT * FROM rule r
    WHERE r.rule_id = latest.rule_id AND r.status = 'active'
    ORDER BY r.version DESC
    LIMIT 1
) a ON latest.latest_status <> 'superseded'
ORDER BY a.priority, a.rule_id
