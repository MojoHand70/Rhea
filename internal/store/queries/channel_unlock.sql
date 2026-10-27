-- The channel's lock let go (channel_lock.sql).
SELECT pg_advisory_unlock(1128808782, hashtext($1))
