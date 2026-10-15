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
    -- which declared door emitted the event: the provenance symmetry
    -- (DIRECTION 2026-10-05) — derived events carry (rule_id, rule_version),
    -- raw events through an activity carry (activity_name, activity_version).
    -- NULL: an adapter, a pack load, or history from before doors existed.
    activity_name    TEXT,
    activity_version INT,
    -- a derived event always knows what caused it and which rule fired
    CHECK (kind = 'raw' OR (cause_event_id IS NOT NULL AND rule_id IS NOT NULL))
);

-- The partition key (DIRECTION 2026-10-08, scale rule 1): the company whose
-- fact an event is, inherited by everything derived from it. Carried on every
-- call as the caller's scope; isolation and sharding ride on it later.
-- NULL: history from before the key, or a fact appended without a scope.
ALTER TABLE event ADD COLUMN IF NOT EXISTS company TEXT;
CREATE INDEX IF NOT EXISTS event_company ON event (company, event_id);

-- Idempotent migrations for databases created before these columns.
ALTER TABLE event ADD COLUMN IF NOT EXISTS actor TEXT;
ALTER TABLE event ADD COLUMN IF NOT EXISTS activity_name TEXT;
ALTER TABLE event ADD COLUMN IF NOT EXISTS activity_version INT;

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
-- A rule's own warrant: the source its reading rests on (core.Warrant).
ALTER TABLE rule ADD COLUMN IF NOT EXISTS warrant JSONB;

CREATE TABLE IF NOT EXISTS object_type (
    name       TEXT NOT NULL,
    version    INT NOT NULL,
    spec       JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (name, version)
);

-- Declared verbs (SPEC §2 Activity): versioned like rules, same lifecycle,
-- same gate. `active` means offered and triggerable — activities are never
-- consulted during replay, so superseding one touches no history.
CREATE TABLE IF NOT EXISTS activity (
    name        TEXT NOT NULL,
    version     INT NOT NULL,
    status      TEXT NOT NULL CHECK (status IN ('draft','approved','active','superseded')),
    domain      TEXT NOT NULL,
    description TEXT NOT NULL,
    spec        JSONB NOT NULL,
    created_by  TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (name, version)
);

CREATE TABLE IF NOT EXISTS view_def (
    view_id    TEXT NOT NULL,
    version    INT NOT NULL,
    notion     TEXT NOT NULL CHECK (notion IN ('list','detail','action','analysis','scheduling','matching','narrative','queue')),
    title      TEXT NOT NULL,
    domain     TEXT NOT NULL,
    function   TEXT NOT NULL,
    spec       JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (view_id, version)
);

-- Idempotent migration: databases created before the scheduling notion
-- (SPEC M5) or the matching grid (DIRECTION 2026-10-11) re-learn the CHECK
-- with the full notion library.
ALTER TABLE view_def DROP CONSTRAINT IF EXISTS view_def_notion_check;
ALTER TABLE view_def ADD CONSTRAINT view_def_notion_check
    CHECK (notion IN ('list','detail','action','analysis','scheduling','matching','narrative','queue'));

-- Definitions under the gate (DECISIONS 2026-10-08, the bundle): object
-- types and views gain the lifecycle rules and activities have. Their version
-- is the schema's identity (objects record type_version), so approval does
-- not mint a new version: it appends the `active` row of the same version,
-- rejection the `superseded` one. Rows from before the gate were installed
-- live, and are active.
ALTER TABLE object_type ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE view_def ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE object_type DROP CONSTRAINT IF EXISTS object_type_status_check;
ALTER TABLE object_type ADD CONSTRAINT object_type_status_check
    CHECK (status IN ('draft','active','superseded'));
ALTER TABLE view_def DROP CONSTRAINT IF EXISTS view_def_status_check;
ALTER TABLE view_def ADD CONSTRAINT view_def_status_check
    CHECK (status IN ('draft','active','superseded'));
ALTER TABLE object_type DROP CONSTRAINT IF EXISTS object_type_pkey;
ALTER TABLE object_type ADD CONSTRAINT object_type_pkey PRIMARY KEY (name, version, status);
ALTER TABLE view_def DROP CONSTRAINT IF EXISTS view_def_pkey;
ALTER TABLE view_def ADD CONSTRAINT view_def_pkey PRIMARY KEY (view_id, version, status);

-- The scope of one approval: draft members activated whole or not at all.
CREATE TABLE IF NOT EXISTS bundle (
    bundle_id   TEXT NOT NULL,
    version     INT NOT NULL,
    status      TEXT NOT NULL CHECK (status IN ('draft','active','superseded')),
    description TEXT NOT NULL,
    members     JSONB NOT NULL,
    created_by  TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (bundle_id, version)
);
-- Why the bundle is proposed (DIRECTION, "Rhea suggests standards").
ALTER TABLE bundle ADD COLUMN IF NOT EXISTS warrant JSONB;
-- The conversation: the question a bundle answers, and the rejected bundle it redrafts.
ALTER TABLE bundle ADD COLUMN IF NOT EXISTS question TEXT;
ALTER TABLE bundle ADD COLUMN IF NOT EXISTS after_bundle TEXT;

-- The blob store (DIRECTION 2026-10-10, "the file is the fact"): every file
-- a door or a channel receives, kept whole under its content hash. A
-- document.received event names a row here, and the store refuses the event
-- without it. Append-only like the log: a file re-read later by a newer
-- channel yields an amendment, never a rewrite.
CREATE TABLE IF NOT EXISTS blob (
    hash       TEXT PRIMARY KEY,          -- "sha256:" + hex of the bytes
    media_type TEXT NOT NULL,
    size       BIGINT NOT NULL,
    bytes      BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
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
ALTER TABLE object ADD COLUMN IF NOT EXISTS company TEXT;
CREATE INDEX IF NOT EXISTS object_company ON object (company, object_type);

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
    FOREACH t IN ARRAY ARRAY['event','rule','object_type','view_def','activity','bundle','blob'] LOOP
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
