-- The worklist: raw events that no rule has acted on — no derived event names
-- them as cause. The human view; the executor's set is pending.sql.
SELECT event_id, kind, event_type, to_char(occurred_at,'YYYY-MM-DD'), recorded_at,
       payload, cause_event_id, rule_id, rule_version, dedup_key, actor,
       activity_name, activity_version, company
FROM event e
WHERE e.kind = 'raw'
  -- the caller's scope: one company's facts, or the whole log when unscoped
  AND ($1 = '' OR e.company = $1 OR COALESCE(e.company, '') = '')
  -- the rule.*, activity.*, bundle.* and backfill.* namespaces are system verbs (approvals, draft
  -- requests), not business events waiting for explanation
  AND e.event_type NOT LIKE 'rule.%'
  AND e.event_type NOT LIKE 'activity.%'
  AND e.event_type NOT LIKE 'bundle.%'
  AND e.event_type NOT LIKE 'backfill.%'
  AND e.event_type NOT LIKE 'matching.%'
  AND e.event_type NOT LIKE 'reading.%'
  AND e.event_type NOT LIKE 'host.%'
  -- who was let in and who was refused: the kernel's own record, never a question
  AND e.event_type NOT LIKE 'access.%'
  -- time passing is infrastructure: an uneventful day is not a question
  AND e.event_type NOT LIKE 'time.%'
  -- a statement saying which account it is of is not a question either: where a pack proposes the
  -- account from it, the proposal is, waiting on Today for a person in Bankowość (2026-10-27)
  AND e.event_type <> 'bank.account.seen'
  AND NOT EXISTS (SELECT 1 FROM event d WHERE d.cause_event_id = e.event_id)
ORDER BY e.event_id
