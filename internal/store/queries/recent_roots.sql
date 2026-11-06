-- The stream's roots (KK 2026-11-05: "we never show events"): the raw facts
-- of the caller's scope, newest first — what came in, so a face can tell what
-- the rules made of each. The company's own facts only: the installation's
-- rows (a pack's chart of accounts, its VAT rates) are the house's story,
-- not what happened in a company's books; unscoped, the whole log.
SELECT event_id, kind, event_type, to_char(occurred_at,'YYYY-MM-DD'), recorded_at,
       payload, cause_event_id, rule_id, rule_version, dedup_key, actor,
       activity_name, activity_version, company
FROM event e
WHERE e.kind = 'raw'
  AND ($1 = '' OR e.company = $1)
  -- the kernel's own verbs are not what happened in the books
  AND e.event_type NOT LIKE 'rule.%'
  AND e.event_type NOT LIKE 'activity.%'
  AND e.event_type NOT LIKE 'bundle.%'
  AND e.event_type NOT LIKE 'backfill.%'
  AND e.event_type NOT LIKE 'matching.%'
  AND e.event_type NOT LIKE 'reading.%'
  AND e.event_type NOT LIKE 'host.%'
  AND e.event_type NOT LIKE 'access.%'
  -- time passing is an event only where it made the books move: a month that opened and
  -- charged depreciation is a card; an uneventful day is not
  AND (e.event_type NOT LIKE 'time.%' OR EXISTS (SELECT 1 FROM event d WHERE d.cause_event_id = e.event_id))
  AND ($2::bigint = 0 OR e.event_id < $2::bigint)
ORDER BY e.event_id DESC
LIMIT $3
