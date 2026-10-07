-- The processing set: raw events no rule has acted on — no derived event
-- names them as cause. This is what ProcessPending evaluates; the worklist
-- (unmatched.sql) is the human view of the same question and additionally
-- hides time.* — an uneventful day is not a question, but it stays pending
-- here so time events fire rules and failed firings retry as data. Quiet
-- days being re-scanned forever is an accepted cost (performance is an
-- explicit non-goal).
SELECT event_id, kind, event_type, to_char(occurred_at,'YYYY-MM-DD'), recorded_at,
       payload, cause_event_id, rule_id, rule_version, dedup_key, actor,
       activity_name, activity_version
FROM event e
WHERE e.kind = 'raw'
  -- the rule.*, activity.* and bundle.* namespaces are system verbs, never rule food
  AND e.event_type NOT LIKE 'rule.%'
  AND e.event_type NOT LIKE 'activity.%'
  AND e.event_type NOT LIKE 'bundle.%'
  AND NOT EXISTS (SELECT 1 FROM event d WHERE d.cause_event_id = e.event_id)
ORDER BY e.event_id
