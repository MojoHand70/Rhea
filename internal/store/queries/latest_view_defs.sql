-- The views a face is served: the newest ACTIVE version of every view, by id.
-- A draft above the running version does not suspend it; a view whose very
-- latest version is superseded is gone — a kept report's own view, removed
-- with its App (2026-10-30). The active row of a version answers before its
-- draft.
SELECT v.view_id, v.version, v.notion, v.title, v.domain, v.function, v.spec, COALESCE(v.company, '')
FROM (
    SELECT DISTINCT ON (view_id) view_id, status AS latest_status
    FROM view_def
    ORDER BY view_id, version DESC, CASE status WHEN 'draft' THEN 1 ELSE 0 END
) latest
JOIN LATERAL (
    SELECT * FROM view_def r
    WHERE r.view_id = latest.view_id AND r.status = 'active'
    ORDER BY r.version DESC
    LIMIT 1
) v ON latest.latest_status <> 'superseded'
ORDER BY latest.view_id
