-- Rhea system of record. Append-only everywhere except `object`, which is a
-- rebuildable projection cache owned by the executor/replayer (invariant 7).

CREATE TABLE IF NOT EXISTS event (
    event_id       BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kind           TEXT NOT NULL CHECK (kind IN ('raw', 'derived')),
    event_type     TEXT NOT NULL,
    occurred_at    DATE NOT NULL,
    recorded_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    payload        JSONB NOT NULL,
    cause_event_id BIGINT REFERENCES event(event_id),
    rule_id        TEXT,
    rule_version   INT,
    dedup_key      TEXT UNIQUE,
    -- who caused the event: cli:<user>, shell:<user>, agent:<model>, kernel.
    -- NULL means recorded before actors existed — honestly unknown.
    actor          TEXT,
    -- a derived event always knows what caused it and which rule fired
    CHECK (kind = 'raw' OR (cause_event_id IS NOT NULL AND rule_id IS NOT NULL))
);

-- Idempotent migration for databases created before the actor column.
ALTER TABLE event ADD COLUMN IF NOT EXISTS actor TEXT;

CREATE TABLE IF NOT EXISTS rule (
    rule_id        TEXT NOT NULL,
    version        INT NOT NULL,
    status         TEXT NOT NULL CHECK (status IN ('draft','approved','active','superseded')),
    priority       INT NOT NULL DEFAULT 100,
    effective_from DATE NOT NULL,
    created_by     TEXT NOT NULL,
    description    TEXT NOT NULL,
    spec           JSONB NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (rule_id, version)
);

CREATE TABLE IF NOT EXISTS object_type (
    name       TEXT NOT NULL,
    version    INT NOT NULL,
    spec       JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (name, version)
);

CREATE TABLE IF NOT EXISTS view_def (
    view_id    TEXT NOT NULL,
    version    INT NOT NULL,
    notion     TEXT NOT NULL CHECK (notion IN ('list','detail','action','analysis')),
    title      TEXT NOT NULL,
    domain     TEXT NOT NULL,
    function   TEXT NOT NULL,
    spec       JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (view_id, version)
);

-- Projection cache. Rebuildable from the event log at any time.
CREATE TABLE IF NOT EXISTS object (
    object_id       TEXT PRIMARY KEY,
    object_type     TEXT NOT NULL,
    type_version    INT NOT NULL,
    state           JSONB NOT NULL,
    source_event_id BIGINT NOT NULL REFERENCES event(event_id),
    rule_id         TEXT NOT NULL,
    rule_version    INT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Invariant 1: the system of record is append-only. UPDATE/DELETE on the
-- versioned tables fails loudly, in the database, no matter who connects.
CREATE OR REPLACE FUNCTION forbid_mutation() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'table % is append-only (Rhea invariant 1)', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY['event','rule','object_type','view_def'] LOOP
        IF NOT EXISTS (
            SELECT 1 FROM pg_trigger WHERE tgname = t || '_append_only'
        ) THEN
            EXECUTE format(
                'CREATE TRIGGER %I BEFORE UPDATE OR DELETE ON %I
                 FOR EACH ROW EXECUTE FUNCTION forbid_mutation()',
                t || '_append_only', t);
        END IF;
    END LOOP;
END;
$$;
