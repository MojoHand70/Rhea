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
-- The worklist anti-joins on the cause, chains and object roots walk it, and
-- an object's explanation reads its amendments by id: indexed, or every
-- pending pass scans the whole log.
CREATE INDEX IF NOT EXISTS event_cause ON event (cause_event_id);
CREATE INDEX IF NOT EXISTS event_amended_object ON event ((payload->>'object_id')) WHERE event_type = 'object.amended';

-- Idempotent migrations for databases created before these columns.
ALTER TABLE event ADD COLUMN IF NOT EXISTS actor TEXT;
ALTER TABLE event ADD COLUMN IF NOT EXISTS activity_name TEXT;
ALTER TABLE event ADD COLUMN IF NOT EXISTS activity_version INT;

CREATE TABLE IF NOT EXISTS rule (
    rule_id        TEXT NOT NULL,
    version        INT NOT NULL,
    status         TEXT NOT NULL CHECK (status IN ('draft','active','superseded')),
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
-- A rule's company (KK, 2026-10-22: company rules and Rhea's global rules):
-- a company's own rule, booking its facts alone; empty or NULL, Rhea's rule
-- for every company. NULL is admitted so a backup written before the column
-- restores its rules as what they were, Rhea's.
ALTER TABLE rule ADD COLUMN IF NOT EXISTS company TEXT;

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
    status      TEXT NOT NULL CHECK (status IN ('draft','active','superseded')),
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
    notion     TEXT NOT NULL CHECK (notion IN ('list','detail','action','analysis','scheduling','matching','narrative','queue','words','statement','report','rights')),
    title      TEXT NOT NULL,
    domain     TEXT NOT NULL,
    function   TEXT NOT NULL,
    spec       JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (view_id, version)
);

-- Idempotent migration: databases created before the scheduling notion, the words and statement notions, the report (2026-10-17), the rights notion (2026-10-20)
-- (SPEC M5) or the matching grid (DIRECTION 2026-10-11) re-learn the CHECK
-- with the full notion library.
ALTER TABLE view_def DROP CONSTRAINT IF EXISTS view_def_notion_check;
ALTER TABLE view_def ADD CONSTRAINT view_def_notion_check
    CHECK (notion IN ('list','detail','action','analysis','scheduling','matching','narrative','queue','words','statement','report','rights'));

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
-- The key gained status once (the gate); rebuilt only where it still lacks
-- it, not on every start under an exclusive lock.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'object_type_pkey' AND array_length(conkey, 1) = 3) THEN
        ALTER TABLE object_type DROP CONSTRAINT IF EXISTS object_type_pkey;
        ALTER TABLE object_type ADD CONSTRAINT object_type_pkey PRIMARY KEY (name, version, status);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'view_def_pkey' AND array_length(conkey, 1) = 3) THEN
        ALTER TABLE view_def DROP CONSTRAINT IF EXISTS view_def_pkey;
        ALTER TABLE view_def ADD CONSTRAINT view_def_pkey PRIMARY KEY (view_id, version, status);
    END IF;
END $$;

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
        -- Row triggers never fire on TRUNCATE: a statement trigger closes
        -- the third verb.
        IF NOT EXISTS (
            SELECT 1 FROM pg_trigger WHERE tgname = t || '_no_truncate'
        ) THEN
            EXECUTE format(
                'CREATE TRIGGER %I BEFORE TRUNCATE ON %I
                 FOR EACH STATEMENT EXECUTE FUNCTION forbid_mutation()',
                t || '_no_truncate', t);
        END IF;
    END LOOP;
END;
$$;

-- The approval is the recorded event, never a resting status (DECISIONS
-- 2026-10-20): databases created before that admit 'approved' no more.
ALTER TABLE rule DROP CONSTRAINT IF EXISTS rule_status_check;
ALTER TABLE rule ADD CONSTRAINT rule_status_check CHECK (status IN ('draft','active','superseded'));
ALTER TABLE activity DROP CONSTRAINT IF EXISTS activity_status_check;
ALTER TABLE activity ADD CONSTRAINT activity_status_check CHECK (status IN ('draft','active','superseded'));

-- Invariant 9: isolation, enforced in the database. A scoped request runs as
-- rhea_app (no bypass) with the scope — a set of company keys — on the
-- connection; a row of event or object is visible, and may be written, when
-- its company is in the scope or it is the installation's (no company: the
-- pack's master data, the kernel's own verbs). The owner (the kernel's own
-- work: replay, backup, the loader) is a superuser and sees the whole log.
DO $$
BEGIN
    -- roles are cluster-wide and two databases may be born at once: a
    -- lost race is nothing
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'rhea_app') THEN
        BEGIN
            CREATE ROLE rhea_app NOLOGIN;
        EXCEPTION WHEN duplicate_object OR unique_violation THEN
            NULL;
        END;
    END IF;
END;
$$;
CREATE OR REPLACE FUNCTION rhea_scope() RETURNS text[] AS $$
    SELECT string_to_array(current_setting('rhea.scope', true), ',')
$$ LANGUAGE sql STABLE;
CREATE OR REPLACE FUNCTION rhea_in_scope(company TEXT) RETURNS boolean AS $$
    SELECT company IS NULL OR company = '' OR company = ANY (rhea_scope())
$$ LANGUAGE sql STABLE;
-- What a scope may write: its own companies' rows; the installation's rows
-- only from the installation's own scope (rhea.scope set and empty), never
-- from a company's — an office acting for no client wrote rows every
-- company saw.
CREATE OR REPLACE FUNCTION rhea_may_write(company TEXT) RETURNS boolean AS $$
    SELECT CASE
        WHEN company IS NULL OR company = '' THEN current_setting('rhea.scope', true) = ''
        ELSE company = ANY (rhea_scope())
    END
$$ LANGUAGE sql STABLE;
ALTER TABLE event ENABLE ROW LEVEL SECURITY;
ALTER TABLE object ENABLE ROW LEVEL SECURITY;
DO $$
DECLARE t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY['event','object'] LOOP
        IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = t AND policyname = t || '_scope') THEN
            EXECUTE format(
                'CREATE POLICY %I ON %I TO rhea_app USING (rhea_in_scope(company)) WITH CHECK (rhea_may_write(company))',
                t || '_scope', t);
        END IF;
        EXECUTE format('ALTER POLICY %I ON %I USING (rhea_in_scope(company)) WITH CHECK (rhea_may_write(company))',
            t || '_scope', t);
    END LOOP;
END;
$$;
-- A company's own rule is seen, and written, only in a scope that names the
-- company, as its facts are; Rhea's rules are every scope's.
ALTER TABLE rule ENABLE ROW LEVEL SECURITY;
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename = 'rule' AND policyname = 'rule_scope') THEN
        CREATE POLICY rule_scope ON rule TO rhea_app USING (rhea_in_scope(company));
    END IF;
END;
$$;
-- A rule's id keeps its company through every version: a company's own
-- reading of Rhea's rule is a rule of its own, never a version of Rhea's, and
-- no company writes a version of another's. Checked over every company's rows
-- (SECURITY DEFINER: a scope's own eyes miss the others'); the refusal names
-- no company.
CREATE OR REPLACE FUNCTION rule_company_fixed() RETURNS trigger AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.rule WHERE rule_id = NEW.rule_id
               AND COALESCE(company, '') <> COALESCE(NEW.company, '')) THEN
        RAISE EXCEPTION 'rule % belongs to another company or to Rhea: a company''s own reading is a rule of its own', NEW.rule_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = public;
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'rule_company_fixed') THEN
        CREATE TRIGGER rule_company_fixed BEFORE INSERT ON rule
            FOR EACH ROW EXECUTE FUNCTION rule_company_fixed();
    END IF;
END;
$$;
ALTER TABLE rule ENABLE ALWAYS TRIGGER rule_company_fixed;
-- A fact is the same fact within its company: the same file sent to two
-- clients is two facts, and one company's key never tells another that it
-- holds the document. NULL companies compare equal here.
ALTER TABLE event DROP CONSTRAINT IF EXISTS event_dedup_key_key;
DROP INDEX IF EXISTS event_company_dedup;
CREATE UNIQUE INDEX IF NOT EXISTS event_company_dedup_key ON event (COALESCE(company, ''), dedup_key) WHERE dedup_key IS NOT NULL;
-- The append-only triggers fire whatever the session's replication role
-- says: a superuser's SET session_replication_role = replica bypassed them.
DO $$
DECLARE t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY['event','rule','object_type','view_def','activity','bundle','blob'] LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ALWAYS TRIGGER %I', t, t || '_append_only');
        EXECUTE format('ALTER TABLE %I ENABLE ALWAYS TRIGGER %I', t, t || '_no_truncate');
    END LOOP;
END;
$$;
GRANT USAGE ON SCHEMA public TO rhea_app;
GRANT SELECT, INSERT ON ALL TABLES IN SCHEMA public TO rhea_app;
GRANT UPDATE, DELETE ON object TO rhea_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO rhea_app;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO rhea_app;

-- A derived event identifies its rule whole: the version too.
ALTER TABLE event DROP CONSTRAINT IF EXISTS event_check;
ALTER TABLE event DROP CONSTRAINT IF EXISTS event_derived_provenance;
ALTER TABLE event ADD CONSTRAINT event_derived_provenance
    CHECK (kind = 'raw' OR (cause_event_id IS NOT NULL AND rule_id IS NOT NULL AND rule_version IS NOT NULL));
-- Lookups and balances are keyed reads, never scans: a ref() by type and
-- field, a balance by account.
CREATE INDEX IF NOT EXISTS object_type_idx ON object (object_type);
CREATE INDEX IF NOT EXISTS object_posting_account ON object ((state->>'account')) WHERE object_type = 'posting';
