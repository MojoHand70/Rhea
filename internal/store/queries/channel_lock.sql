-- One channel's passes, one at a time across processes: the advisory lock
-- keyed by the class "CHAN" and the channel's name (lock.go, LockChannel).
SELECT pg_advisory_lock(1128808782, hashtext($1))
