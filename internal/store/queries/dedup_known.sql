-- Whether the log already holds a company's fact by its dedup key: a
-- channel's fact carrying files keeps them only when it is new.
SELECT EXISTS (SELECT 1 FROM event WHERE COALESCE(company, '') = $1 AND dedup_key = $2)
