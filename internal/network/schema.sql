-- The network's store: what installations published about how they explain
-- their business — rule shapes, never data. Append-only like every log in
-- Rhea; an installation's latest publication is its current knowledge.
CREATE TABLE IF NOT EXISTS publication (
    publication_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    installation   TEXT NOT NULL,
    shapes         JSONB NOT NULL,
    published_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- A simulated customer's publication is marked: it stays out of every count
-- a real client sees, and informs only other simulated runs.
ALTER TABLE publication ADD COLUMN IF NOT EXISTS synthetic BOOLEAN NOT NULL DEFAULT false;

CREATE OR REPLACE FUNCTION forbid_publication_mutation() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'publication is append-only (Rhea invariant 1)';
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'publication_append_only') THEN
        CREATE TRIGGER publication_append_only BEFORE UPDATE OR DELETE ON publication
            FOR EACH ROW EXECUTE FUNCTION forbid_publication_mutation();
    END IF;
END;
$$;

-- What installations were asked and had not been taught (DIRECTION 2026-10-17,
-- "what she was never taught reaches the house"): the host's general
-- description of the missing skill — never the person's words, a name or a
-- number, which stay in the installation. The Rhea team reads them counted
-- across installations. Append-only.
CREATE TABLE IF NOT EXISTS untaught (
    untaught_id  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    installation TEXT NOT NULL,
    missing      TEXT NOT NULL,
    noted_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    synthetic    BOOLEAN NOT NULL DEFAULT false
);

-- What kind of moment it was (2026-10-18): never taught, the books do not hold it, or her query
-- failed — and, for a failure, only the class of DuckDB's refusal (never the query, which may carry
-- the client's names).
ALTER TABLE untaught ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'untaught';
ALTER TABLE untaught ADD COLUMN IF NOT EXISTS detail TEXT NOT NULL DEFAULT '';

CREATE OR REPLACE FUNCTION forbid_untaught_mutation() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'untaught is append-only (Rhea invariant 1)';
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'untaught_append_only') THEN
        CREATE TRIGGER untaught_append_only BEFORE UPDATE OR DELETE ON untaught
            FOR EACH ROW EXECUTE FUNCTION forbid_untaught_mutation();
    END IF;
END;
$$;

-- The Rhea team's notes on a lesson she was never taught (the Learning Centre): where it stands —
-- open, planned for a pack, done, not for us — and a note. Filed under the lesson's key (its
-- description, letter case and outer spaces aside); append-only, the latest note is the lesson's.
CREATE TABLE IF NOT EXISTS lesson_note (
    lesson_note_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    lesson         TEXT NOT NULL,
    status         TEXT NOT NULL CHECK (status IN ('open', 'planned', 'done', 'not_for_us')),
    note           TEXT NOT NULL DEFAULT '',
    noted_by       TEXT NOT NULL,
    noted_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION forbid_lesson_note_mutation() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'lesson_note is append-only (Rhea invariant 1)';
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'lesson_note_append_only') THEN
        CREATE TRIGGER lesson_note_append_only BEFORE UPDATE OR DELETE ON lesson_note
            FOR EACH ROW EXECUTE FUNCTION forbid_lesson_note_mutation();
    END IF;
END;
$$;
