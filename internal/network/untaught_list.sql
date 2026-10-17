-- What installations were never taught: one row per description (letter case and outer spaces
-- aside), counted once per installation and once per time, most asked first.
-- $1: include simulated customers.
SELECT min(missing) AS missing, count(DISTINCT installation) AS installations, count(*) AS times, max(noted_at) AS last
FROM untaught
WHERE $1 OR NOT synthetic
GROUP BY lower(btrim(missing))
ORDER BY installations DESC, times DESC, last DESC
