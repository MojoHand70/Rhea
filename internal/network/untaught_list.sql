-- What installations could not get answered: one row per description (letter case and outer spaces
-- aside), counted once per installation and once per time, most asked first, with the kinds of
-- moment it was and the classes of refusal, and the lesson key the team's notes are filed under.
-- $1: include simulated customers.
SELECT min(missing) AS missing, lower(btrim(missing)) AS lesson,
       count(DISTINCT installation) AS installations, array_agg(DISTINCT installation ORDER BY installation) AS names,
       count(*) AS times, min(noted_at) AS first, max(noted_at) AS last,
       array_agg(DISTINCT kind ORDER BY kind) AS kinds,
       array_remove(array_agg(DISTINCT detail ORDER BY detail), '') AS details
FROM untaught
WHERE $1 OR NOT synthetic
GROUP BY lower(btrim(missing))
ORDER BY installations DESC, times DESC, last DESC
