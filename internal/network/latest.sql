-- Every installation's current knowledge: its latest publication. Synthetic
-- publications (simulated customers) count only when asked for ($1).
SELECT DISTINCT ON (installation) installation, shapes
FROM publication
WHERE NOT synthetic OR $1
ORDER BY installation, publication_id DESC
