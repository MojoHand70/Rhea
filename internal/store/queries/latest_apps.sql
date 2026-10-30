-- The Apps a face is served: the newest ACTIVE version of every App, by id.
-- A draft above the running version does not suspend it; an App whose very
-- latest version is superseded is gone — a kept report its company removed
-- (2026-10-30) — as a rule whose latest version is superseded is retired.
-- The active row of a version answers before its draft.
SELECT a.spec
FROM (
    SELECT DISTINCT ON (app_id) app_id, status AS latest_status
    FROM app
    ORDER BY app_id, version DESC, CASE status WHEN 'draft' THEN 1 ELSE 0 END
) latest
JOIN LATERAL (
    SELECT spec FROM app r
    WHERE r.app_id = latest.app_id AND r.status = 'active'
    ORDER BY r.version DESC
    LIMIT 1
) a ON latest.latest_status <> 'superseded'
ORDER BY latest.app_id
