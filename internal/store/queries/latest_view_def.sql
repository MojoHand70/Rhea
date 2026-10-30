-- One view's newest row, of any status, with that status. The active row of
-- a version answers before its draft.
SELECT view_id, version, notion, title, domain, function, spec, status, COALESCE(company, '')
FROM view_def WHERE view_id = $1
ORDER BY version DESC, CASE status WHEN 'draft' THEN 1 ELSE 0 END
LIMIT 1
