-- What installations were never taught: one row per description (letter case and outer spaces
-- aside), counted once per installation and once per time, most asked first, with the lesson key
-- the team's notes are filed under. $1: include simulated customers.
SELECT min(missing) AS missing, lower(btrim(missing)) AS lesson,
       count(DISTINCT installation) AS installations, array_agg(DISTINCT installation ORDER BY installation) AS names,
       count(*) AS times, min(noted_at) AS first, max(noted_at) AS last
FROM untaught
WHERE $1 OR NOT synthetic
GROUP BY lower(btrim(missing))
ORDER BY installations DESC, times DESC, last DESC
