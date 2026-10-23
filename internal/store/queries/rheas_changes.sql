-- What Rhea changed since a day: every version of a rule of Rhea's (no
-- company) that became active on or after it, newest first. A company's own
-- rules are not Rhea's changes; every company reads these.
SELECT rule_id, version, status, priority, to_char(effective_from,'YYYY-MM-DD'),
       created_by, description, spec, warrant, created_at, COALESCE(company, ''), COALESCE(replaces, '')
FROM rule
WHERE COALESCE(company, '') = '' AND status = 'active' AND created_at >= $1::date
ORDER BY created_at DESC, rule_id
