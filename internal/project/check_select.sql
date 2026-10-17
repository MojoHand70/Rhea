-- How DuckDB itself reads a query: one SELECT, or an error naming why not.
SELECT CAST(json_serialize_sql(CAST(? AS VARCHAR)) AS VARCHAR)
