-- The offered verb set: the newest ACTIVE version of every activity, by name.
-- Same semantics as active_rules.sql: a draft in progress does not suspend
-- the offered version; only an activity whose very latest version is
-- superseded is withdrawn.
SELECT a.name, a.version, a.status, a.domain, a.description, a.spec,
       a.created_by, a.created_at
FROM (
    SELECT DISTINCT ON (name) name, status AS latest_status
    FROM activity
    ORDER BY name, version DESC
) latest
JOIN LATERAL (
    SELECT * FROM activity x
    WHERE x.name = latest.name AND x.status = 'active'
    ORDER BY x.version DESC
    LIMIT 1
) a ON latest.latest_status <> 'superseded'
ORDER BY a.name
