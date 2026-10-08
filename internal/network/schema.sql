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
