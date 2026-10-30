-- One App's newest row, of any status, with that status: what a removal
-- writes above, and whether the App shows. The active row of a version
-- answers before its draft.
SELECT spec, status FROM app WHERE app_id = $1
ORDER BY version DESC, CASE status WHEN 'draft' THEN 1 ELSE 0 END
LIMIT 1
