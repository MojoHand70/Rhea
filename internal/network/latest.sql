-- Every installation's current knowledge: its latest publication.
SELECT DISTINCT ON (installation) installation, shapes
FROM publication
ORDER BY installation, publication_id DESC
