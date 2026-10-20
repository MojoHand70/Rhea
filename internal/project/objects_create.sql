-- The JSON projection the packs' hand-written analysis SQL reads: every object with its state as
-- JSON text and its provenance.
DROP TABLE IF EXISTS objects;
CREATE TABLE objects (
    object_id       VARCHAR PRIMARY KEY,
    object_type     VARCHAR NOT NULL,
    type_version    INTEGER NOT NULL,
    state           VARCHAR NOT NULL,
    source_event_id BIGINT NOT NULL,
    rule_id         VARCHAR NOT NULL,
    rule_version    INTEGER NOT NULL,
    occurred_at     DATE NOT NULL,
    root_event_id   BIGINT NOT NULL,
    company         VARCHAR
)
