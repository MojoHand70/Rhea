-- Whether a company's books hold a fact: a raw fact under its key that a
-- person, a door or a channel put there — not the kernel's own records (the
-- system verbs, the host's words, who was let in or refused) and not a day
-- the clock opened, which every company the log names receives
-- (rights.Claimed).
SELECT EXISTS (
    SELECT 1 FROM event
    WHERE company = $1
      AND kind = 'raw'
      AND event_type NOT LIKE 'time.%'
      AND event_type NOT LIKE 'rule.%'
      AND event_type NOT LIKE 'activity.%'
      AND event_type NOT LIKE 'bundle.%'
      AND event_type NOT LIKE 'backfill.%'
      AND event_type NOT LIKE 'matching.%'
      AND event_type NOT LIKE 'reading.%'
      AND event_type NOT LIKE 'host.%'
      AND event_type NOT LIKE 'access.%'
)
