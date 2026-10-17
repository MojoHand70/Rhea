-- The collection's schemas as they stand in the copy: everything but DuckDB's own.
SELECT schema_name FROM information_schema.schemata
WHERE catalog_name = current_database() AND schema_name NOT IN ('main', 'information_schema', 'pg_catalog', 'temp')
