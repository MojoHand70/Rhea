-- What the copy was built from: the log's last event, the scope, the catalog — rebuilt only when it moved.
DROP TABLE IF EXISTS _rhea_meta;
CREATE TABLE _rhea_meta (built_from VARCHAR NOT NULL, objects BIGINT NOT NULL)
